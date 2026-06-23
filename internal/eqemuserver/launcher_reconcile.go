package eqemuserver

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

const (
	// defaultStaleZoneGraceSeconds is applied when the operator has not
	// configured a grace period. It is long enough to cover a slow zone boot
	// or a transient world reconnect without killing a healthy zone.
	defaultStaleZoneGraceSeconds = 120

	// defaultMaxStaleZoneKillsPerCycle bounds kills per reconciliation pass so
	// the pool is drained-and-replenished gradually rather than all at once.
	defaultMaxStaleZoneKillsPerCycle = 2

	// staleZoneKillWindow is the rolling window over which kills are rate
	// limited as a guardrail against runaway reconciliation.
	staleZoneKillWindow = 5 * time.Minute

	// staleZoneKillWindowMultiple scales the rolling-window kill cap relative
	// to the per-cycle cap (window cap = perCycleCap * multiple).
	staleZoneKillWindowMultiple = 5

	// staleZoneTerminateGrace is how long we wait after a graceful SIGTERM
	// before escalating to SIGKILL for a stale zone child.
	staleZoneTerminateGrace = 5 * time.Second

	// staleZoneTerminatePoll is the interval used while waiting for a
	// terminated process to exit.
	staleZoneTerminatePoll = 500 * time.Millisecond
)

// zoneProcInfo is the launcher-local view of a single zone child process.
type zoneProcInfo struct {
	Pid        int32
	StaticName string // non-empty when this zone was launched for a configured static zone
	Age        time.Duration
}

// reconcileStats summarizes one reconciliation decision for observability.
type reconcileStats struct {
	Eligible  int // stale dynamics old enough to kill this cycle
	Killed    int // dynamics selected for termination this cycle (pre window cap)
	Deferred  int // eligible but deferred due to per-cycle cap
	NewTimers int // dynamics that started a grace timer this cycle
	Tracked   int // dynamics currently tracked as unregistered (timer active)
}

// zoneStaticArg returns the configured static zone short-name this zone process
// was launched for (e.g. the "bazaar" in `zone bazaar`), or "" for a dynamic
// (argument-less) zone. This is the single source of truth shared by
// pollProcessCounts (counting / boot loop) and reconciliation (kill
// classification) so the two paths can never diverge on what is static.
func (l *Launcher) zoneStaticArg(proc ProcessDetails) string {
	// cwd can contain spaces; strip it before splitting the command line
	cmdline := strings.ReplaceAll(proc.Cmdline, proc.Cwd, "")
	cmdline = strings.TrimSpace(cmdline)
	arg := strings.Split(cmdline, " ")
	if len(arg) <= 1 {
		return ""
	}
	for _, z := range l.staticZonesToBoot {
		if z == arg[1] {
			return arg[1]
		}
	}
	return ""
}

// decideStaleZoneKills is the pure, side-effect-free reconciliation decision.
//
// Inputs:
//   - local:      every zone child process currently alive on this launcher
//     (dynamic and static), each tagged with age and static name.
//   - registered: the authoritative set of OS PIDs that world reports as
//     registered, usable zone servers (from GetZoneList).
//   - prevSince:  prior per-pid "first seen unregistered" timestamps (state).
//   - now/grace:  evaluation time and the boot/reconnect grace period.
//   - perCycleCap:max dynamics to terminate in a single pass.
//
// It returns the list of dynamic PIDs to kill, the updated state map to persist
// (timers only retained for pids still alive), and summary stats.
//
// Static zones are never selected for killing here; v1 reconciliation only
// heals the dynamic pool. A dynamic zone is eligible only if it is both older
// than grace AND has been continuously unregistered for at least grace.
func decideStaleZoneKills(
	local []zoneProcInfo,
	registered map[int32]bool,
	prevSince map[int32]time.Time,
	now time.Time,
	grace time.Duration,
	perCycleCap int,
) (killPids []int32, nextSince map[int32]time.Time, stats reconcileStats) {
	nextSince = make(map[int32]time.Time)

	type candidate struct {
		pid             int32
		unregisteredFor time.Duration
	}
	var eligible []candidate

	for _, zp := range local {
		// static zones are handled by the static boot loop; never auto-kill here
		if zp.StaticName != "" {
			continue
		}

		if registered[zp.Pid] {
			// healthy / registered this cycle -> drop any pending timer
			continue
		}

		// unregistered dynamic: start or continue its grace timer
		since, hadTimer := prevSince[zp.Pid]
		if !hadTimer {
			since = now
			stats.NewTimers++
		}
		nextSince[zp.Pid] = since

		unregisteredFor := now.Sub(since)
		// require BOTH continuous unregistration for grace AND minimum process
		// age, so slow boots and brief reconnects are not killed
		if unregisteredFor >= grace && zp.Age >= grace {
			eligible = append(eligible, candidate{pid: zp.Pid, unregisteredFor: unregisteredFor})
		}
	}

	// kill oldest-stale first for deterministic, gradual convergence
	sort.Slice(eligible, func(i, j int) bool {
		return eligible[i].unregisteredFor > eligible[j].unregisteredFor
	})

	if perCycleCap < 1 {
		perCycleCap = defaultMaxStaleZoneKillsPerCycle
	}
	for i, c := range eligible {
		if i >= perCycleCap {
			break
		}
		killPids = append(killPids, c.pid)
	}

	stats.Eligible = len(eligible)
	stats.Killed = len(killPids)
	stats.Deferred = len(eligible) - len(killPids)
	stats.Tracked = len(nextSince)
	return killPids, nextSince, stats
}

// reconcileStaleZoneProcesses reconciles local zone child process liveness with
// world's authoritative registered zoneserver list and terminates stale
// dynamic zone children so the dynamic pool can be replenished.
//
// It reuses the zone-child snapshot captured by this cycle's pollProcessCounts
// (no second process-table scan). The caller guarantees `list` was fetched
// successfully (world reachable); this method performs no action while the
// feature is disabled, while world reports zero usable zones, or while a
// stop/restart timer is pending.
func (l *Launcher) reconcileStaleZoneProcesses(list WorldZoneList) {
	if !l.reconcileStaleZones {
		return
	}

	// do not reconcile while a graceful stop/restart is counting down
	if l.GetStopTimer() > 0 {
		return
	}

	// authoritative: the set of OS PIDs world considers usable zone servers
	registered := make(map[int32]bool)
	for _, z := range list.Data {
		if z.ZoneOsPid > 0 {
			registered[int32(z.ZoneOsPid)] = true
		}
	}

	// snapshot the zone children captured by this cycle's pollProcessCounts
	l.pollProcessMutex.Lock()
	local := append([]zoneProcInfo(nil), l.currentZoneProcInfos...)
	l.pollProcessMutex.Unlock()

	// guard: world is reachable but reports zero usable zones (still booting,
	// just restarted, or a degraded/partial telnet response). Treating every
	// live zone as stale in that state could drain the pool, so defer until
	// world reports at least one registered zone.
	if len(registered) == 0 {
		l.logger.Warn().
			Any("local_zone_children", len(local)).
			Msg("Zone reconcile: world returned no registered zones; deferring (world may be booting or degraded)")
		return
	}

	now := time.Now()
	grace := time.Duration(l.staleZoneGraceSeconds) * time.Second

	// decide under lock; state (timers + kill window) is launcher-wide
	l.reconcileMutex.Lock()
	prevSince := l.staleZoneSince
	killPids, nextSince, stats := decideStaleZoneKills(local, registered, prevSince, now, grace, l.maxStaleZoneKillsPerCycle)
	l.staleZoneSince = nextSince

	// prune the rolling kill window
	windowCutoff := now.Add(-staleZoneKillWindow)
	kept := l.reconcileKillTimes[:0]
	for _, t := range l.reconcileKillTimes {
		if t.After(windowCutoff) {
			kept = append(kept, t)
		}
	}
	l.reconcileKillTimes = kept
	windowCap := l.maxStaleZoneKillsPerCycle * staleZoneKillWindowMultiple
	if windowCap < 1 {
		windowCap = defaultMaxStaleZoneKillsPerCycle
	}
	l.reconcileMutex.Unlock()

	if stats.NewTimers > 0 {
		l.logger.Warn().
			Any("count", stats.NewTimers).
			Any("grace_seconds", l.staleZoneGraceSeconds).
			Msg("Zone reconcile: dynamic zone(s) alive but not registered with world, grace timer started")
	}

	// surface stale static zones (observed, never killed in v1)
	for _, zp := range local {
		if zp.StaticName == "" || registered[zp.Pid] {
			continue
		}
		l.logger.Debug().
			Any("pid", zp.Pid).
			Any("static_zone", zp.StaticName).
			Msg("Zone reconcile: static zone alive but not registered with world (not killed; static reconciliation disabled)")
	}

	killed := 0
	for _, pid := range killPids {
		l.reconcileMutex.Lock()
		windowFull := len(l.reconcileKillTimes) >= windowCap
		l.reconcileMutex.Unlock()
		if windowFull {
			l.logger.Warn().
				Any("pid", pid).
				Any("window_cap", windowCap).
				Any("window", staleZoneKillWindow.String()).
				Msg("Zone reconcile: stale zone eligible but rolling kill window cap reached, deferring")
			continue
		}

		age := time.Duration(0)
		for _, zp := range local {
			if zp.Pid == pid {
				age = zp.Age
				break
			}
		}
		unregisteredFor := grace
		if t, ok := nextSince[pid]; ok {
			unregisteredFor = now.Sub(t)
		}

		l.logger.Info().
			Any("pid", pid).
			Any("age", age.Round(time.Second).String()).
			Any("unregistered_for", unregisteredFor.Round(time.Second).String()).
			Any("grace", grace.String()).
			Msg("Zone reconcile: killing stale dynamic zone (alive but not registered with world)")

		ok, err := l.terminateZoneProcess(pid)
		if err != nil {
			l.logger.Error().Err(err).Any("pid", pid).Msg("Zone reconcile: error terminating stale zone")
			continue
		}
		if !ok {
			continue
		}

		killed++
		// NOTE: we intentionally do NOT mutate currentProcessCounts here. The
		// cached count is owned by pollProcessCounts (which also resets it from
		// other call sites, e.g. rpcZoneCountDynamic); decrementing it from the
		// reconciler would race a concurrent reset and corrupt the count. The
		// next Supervisor cycle's pollProcessCounts() recounts reality and the
		// boot loop then replenishes -- the ~1s delay is immaterial vs. a
		// failure that develops over minutes, and this keeps the kill path
		// concurrency-safe.

		l.reconcileMutex.Lock()
		l.reconcileKillTimes = append(l.reconcileKillTimes, now)
		l.reconcileMutex.Unlock()
		l.broadcastReconcileKill(pid, age, unregisteredFor)
	}

	if stats.Eligible > 0 {
		l.logger.Info().
			Any("eligible", stats.Eligible).
			Any("killed", killed).
			Any("deferred", stats.Eligible-killed).
			Any("tracked", stats.Tracked).
			Msg("Zone reconcile: cycle summary")
	}
}

// captureZoneIdentity returns the process create-time iff `pid` currently refers
// to a live zone process, else (0, false). It builds a fresh process object so
// the create-time is read from the OS (not a struct-cached value), which makes
// this a reliable PID-reuse check: if the PID was reused by any other process
// (or a different zone), the create-time differs.
func (l *Launcher) captureZoneIdentity(pid int32) (int64, bool) {
	p, err := process.NewProcess(pid)
	if err != nil {
		return 0, false
	}
	proc := l.getProcessDetails(p)
	if proc.BaseProcessName != zoneProcessName {
		return 0, false
	}
	ct, err := p.CreateTime()
	if err != nil || ct <= 0 {
		return 0, false
	}
	return ct, true
}

// terminateZoneProcess gracefully terminates (SIGTERM) then force-kills
// (SIGKILL) a zone child by PID. It captures the process identity at entry and
// re-validates it immediately before SIGKILL to defend against PID reuse during
// the escalation grace window. Returns ok=true if the process was
// terminated/killed (or was already gone), ok=false if it no longer refers to a
// zone process.
func (l *Launcher) terminateZoneProcess(pid int32) (bool, error) {
	startTime, ok := l.captureZoneIdentity(pid)
	if !ok {
		// already reaped, or no longer a zone process
		return false, nil
	}

	p, err := process.NewProcess(pid)
	if err != nil {
		return false, nil
	}

	if err := p.Terminate(); err != nil {
		l.logger.Debug().Err(err).Any("pid", pid).Msg("Zone reconcile: SIGTERM returned error")
	}

	deadline := time.Now().Add(staleZoneTerminateGrace)
	for time.Now().Before(deadline) {
		alive, err := p.IsRunning()
		if err != nil || !alive {
			l.logger.Info().Any("pid", pid).Msg("Zone reconcile: stale zone terminated gracefully")
			return true, nil
		}
		time.Sleep(staleZoneTerminatePoll)
	}

	// escalate to SIGKILL only if this is still the exact zone process we
	// targeted; if the PID was reused mid-wait, the original already exited and
	// there is nothing safe to force-kill.
	if cur, ok := l.captureZoneIdentity(pid); !ok || cur != startTime {
		l.logger.Info().
			Any("pid", pid).
			Msg("Zone reconcile: pid no longer the original zone before SIGKILL, aborting")
		return true, nil
	}

	if err := p.Kill(); err != nil {
		l.logger.Debug().Err(err).Any("pid", pid).Msg("Zone reconcile: SIGKILL returned error")
	}
	time.Sleep(staleZoneTerminatePoll)
	if alive, _ := p.IsRunning(); !alive {
		l.logger.Warn().Any("pid", pid).Msg("Zone reconcile: stale zone force-killed after grace")
		return true, nil
	}

	return false, fmt.Errorf("zone process %d did not exit after SIGTERM and SIGKILL", pid)
}

// broadcastReconcileKill emits a UI notification so reconciliation actions are
// visible in the Spire dashboard, not only in logs.
func (l *Launcher) broadcastReconcileKill(pid int32, age, unregisteredFor time.Duration) {
	if l.websocketMgr == nil {
		return
	}
	payload := map[string]any{
		"pid":              pid,
		"age_seconds":      int(age.Round(time.Second).Seconds()),
		"unregistered_for": int(unregisteredFor.Round(time.Second).Seconds()),
		"reason":           "stale dynamic zone (alive but not registered with world)",
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}
	l.websocketMgr.Broadcast("zoneReconcileKill", string(data))
}
