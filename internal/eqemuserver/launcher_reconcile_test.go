package eqemuserver

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/EQEmu/spire/internal/logger"
)

func int32Set(vals ...int32) map[int32]bool {
	m := make(map[int32]bool, len(vals))
	for _, v := range vals {
		m[v] = true
	}
	return m
}

// TestReconcile_KillsStaleDynamicWithinGrace verifies the core prod-failure
// scenario: a dynamic zone child that is alive but unregistered for longer than
// grace (and old enough) is selected for termination.
func TestReconcile_KillsStaleDynamicWithinGrace(t *testing.T) {
	now := time.Now()
	grace := 120 * time.Second
	local := []zoneProcInfo{
		{Pid: 10, Age: 30 * time.Minute}, // stale, old -> kill
		{Pid: 11, Age: 30 * time.Minute}, // healthy (registered)
		{Pid: 12, Age: 10 * time.Second}, // young, unregistered -> grace protects
	}
	registered := int32Set(11)

	// pid 10 has been unregistered for a while already
	prev := map[int32]time.Time{10: now.Add(-5 * time.Minute)}

	kill, next, stats := decideStaleZoneKills(local, registered, prev, now, grace, 2)

	if len(kill) != 1 || kill[0] != 10 {
		t.Fatalf("expected to kill only stale pid 10, got %v", kill)
	}
	if stats.Eligible != 1 || stats.Killed != 1 {
		t.Fatalf("unexpected stats: %+v", stats)
	}
	// pid 12 (young) should be tracked but not killed
	if _, ok := next[12]; !ok {
		t.Fatalf("young unregistered pid 12 should be tracked with a grace timer")
	}
	// pid 10 timer retained so it is retried promptly if the kill fails
	if _, ok := next[10]; !ok {
		t.Fatalf("killed pid 10 should retain its timer for retry-on-failure")
	}
}

// TestReconcile_DoesNotKillBootingZone verifies a freshly spawned dynamic zone
// that has not yet registered is not killed until it has exceeded grace in both
// age and continuous unregistration.
func TestReconcile_DoesNotKillBootingZone(t *testing.T) {
	now := time.Now()
	grace := 120 * time.Second

	// spawned 60s ago, never registered -> still within grace
	local := []zoneProcInfo{{Pid: 20, Age: 60 * time.Second}}
	kill, _, stats := decideStaleZoneKills(local, int32Set(), nil, now, grace, 2)
	if len(kill) != 0 {
		t.Fatalf("booting zone within grace must not be killed, got %v", kill)
	}
	if stats.NewTimers != 1 {
		t.Fatalf("expected a new grace timer, got stats %+v", stats)
	}

	// now aged past grace AND continuously unregistered past grace -> eligible
	now2 := now.Add(grace + time.Second)
	local[0].Age = grace + time.Second
	kill2, _, stats2 := decideStaleZoneKills(local, int32Set(), map[int32]time.Time{20: now}, now2, grace, 2)
	if len(kill2) != 1 || kill2[0] != 20 {
		t.Fatalf("stale zone past grace should be killed, got %v", kill2)
	}
	if stats2.NewTimers != 0 {
		t.Fatalf("timer should already exist, got NewTimers %d", stats2.NewTimers)
	}
}

// TestReconcile_StaticZonesNeverKilled verifies static zones are excluded from
// dynamic reconciliation even when stale.
func TestReconcile_StaticZonesNeverKilled(t *testing.T) {
	now := time.Now()
	grace := 120 * time.Second
	local := []zoneProcInfo{
		{Pid: 30, Age: 1 * time.Hour, StaticName: "bazaar"}, // stale static
		{Pid: 31, Age: 1 * time.Hour},                       // stale dynamic
	}
	prev := map[int32]time.Time{
		30: now.Add(-1 * time.Hour),
		31: now.Add(-1 * time.Hour),
	}
	kill, _, stats := decideStaleZoneKills(local, int32Set(), prev, now, grace, 2)
	if len(kill) != 1 || kill[0] != 31 {
		t.Fatalf("only the stale dynamic should be killed, got %v", kill)
	}
	if stats.Eligible != 1 {
		t.Fatalf("static must not count as eligible, got stats %+v", stats)
	}
}

// TestReconcile_PerCycleCap verifies the per-cycle kill cap defers excess kills.
func TestReconcile_PerCycleCap(t *testing.T) {
	now := time.Now()
	grace := 120 * time.Second
	local := []zoneProcInfo{
		{Pid: 1, Age: 1 * time.Hour},
		{Pid: 2, Age: 1 * time.Hour},
		{Pid: 3, Age: 1 * time.Hour},
		{Pid: 4, Age: 1 * time.Hour},
	}
	prev := map[int32]time.Time{}
	for _, z := range local {
		prev[z.Pid] = now.Add(-1 * time.Hour)
	}

	kill, _, stats := decideStaleZoneKills(local, int32Set(), prev, now, grace, 2)
	if len(kill) != 2 {
		t.Fatalf("per-cycle cap of 2 should yield 2 kills, got %d", len(kill))
	}
	if stats.Eligible != 4 || stats.Deferred != 2 {
		t.Fatalf("unexpected stats: %+v", stats)
	}
}

// TestReconcile_HealedZoneClearsTimer verifies a zone that re-registers clears
// its pending stale timer (no flapping / over-eager kill after a transient dip).
func TestReconcile_HealedZoneClearsTimer(t *testing.T) {
	now := time.Now()
	grace := 120 * time.Second
	local := []zoneProcInfo{{Pid: 40, Age: 1 * time.Hour}}
	prev := map[int32]time.Time{40: now.Add(-10 * time.Minute)}

	// now registered again
	_, next, _ := decideStaleZoneKills(local, int32Set(40), prev, now, grace, 2)
	if _, ok := next[40]; ok {
		t.Fatalf("re-registered zone should have its timer cleared, still tracked: %+v", next)
	}
}

// TestReconcile_GonePidPruned verifies timers for processes that have exited
// are dropped from persisted state.
func TestReconcile_GonePidPruned(t *testing.T) {
	now := time.Now()
	grace := 120 * time.Second
	// pid 50 is no longer in `local` (exited) but had a timer
	prev := map[int32]time.Time{50: now.Add(-5 * time.Minute)}
	_, next, _ := decideStaleZoneKills(nil, int32Set(), prev, now, grace, 2)
	if _, ok := next[50]; ok {
		t.Fatalf("gone pid should be pruned from state, still present: %+v", next)
	}
}

// TestReconcile_RequiresBothAgeAndUnregistered verifies that a zone old in age
// but only briefly unregistered (e.g. a transient world list dip) is not killed.
func TestReconcile_RequiresBothAgeAndUnregistered(t *testing.T) {
	now := time.Now()
	grace := 120 * time.Second
	local := []zoneProcInfo{{Pid: 60, Age: 1 * time.Hour}} // old process
	// but only unregistered for 10s (transient dip)
	prev := map[int32]time.Time{60: now.Add(-10 * time.Second)}
	kill, _, _ := decideStaleZoneKills(local, int32Set(), prev, now, grace, 2)
	if len(kill) != 0 {
		t.Fatalf("transient unregistration must not trigger a kill, got %v", kill)
	}
}

// zoneListWithPids builds a WorldZoneList whose registered zones carry the given
// OS PIDs (via JSON so the anonymous Data element need not be spelled out).
func zoneListWithPids(pids ...int) WorldZoneList {
	parts := make([]string, 0, len(pids))
	for _, p := range pids {
		parts = append(parts, fmt.Sprintf(`{"zone_os_pid":%d}`, p))
	}
	raw := `{"data":[` + strings.Join(parts, ",") + `]}`
	var list WorldZoneList
	_ = json.Unmarshal([]byte(raw), &list)
	return list
}

func newReconcileTestLauncher() *Launcher {
	l := NewLauncher(logger.NewAppLogger(), nil, nil, nil, nil, nil, nil)
	l.reconcileStaleZones = true
	l.staleZoneGraceSeconds = 120
	l.maxStaleZoneKillsPerCycle = 2
	return l
}

// TestReconcileMethod_EmptyWorldListDefers verifies the deploy-safety guard:
// when world is reachable but reports zero registered zones (booting/degraded),
// reconciliation defers and starts no grace timers.
func TestReconcileMethod_EmptyWorldListDefers(t *testing.T) {
	l := newReconcileTestLauncher()
	l.currentZoneProcInfos = []zoneProcInfo{{Pid: 123, Age: 1 * time.Hour}}
	l.currentProcessCounts = map[string]int{zoneProcessName: 1}

	l.reconcileStaleZoneProcesses(WorldZoneList{})

	if len(l.staleZoneSince) != 0 {
		t.Fatalf("expected no grace timers when world list is empty, got %v", l.staleZoneSince)
	}
	if l.currentProcessCounts[zoneProcessName] != 1 {
		t.Fatalf("count must not change when deferring, got %d", l.currentProcessCounts[zoneProcessName])
	}
}

// TestReconcileMethod_DoesNotMutateProcessCounts verifies the reconciler never
// touches the shared currentProcessCounts (which is owned by pollProcessCounts
// and reset from other call sites). Mutating it here would race a concurrent
// reset and corrupt the count; replenishment must happen via the next poll.
func TestReconcileMethod_DoesNotMutateProcessCounts(t *testing.T) {
	l := newReconcileTestLauncher()
	const fakePid = int32(9999999) // does not exist
	l.currentZoneProcInfos = []zoneProcInfo{{Pid: fakePid, Age: 1 * time.Hour}}
	l.currentProcessCounts = map[string]int{zoneProcessName: 5}
	// pid has been unregistered well past grace
	l.staleZoneSince = map[int32]time.Time{fakePid: time.Now().Add(-1 * time.Hour)}
	// world reports a different pid as registered, so fakePid is stale
	list := zoneListWithPids(1)

	l.reconcileStaleZoneProcesses(list)

	if _, ok := l.staleZoneSince[fakePid]; !ok {
		t.Fatalf("grace timer should be retained for the stale pid")
	}
	if l.currentProcessCounts[zoneProcessName] != 5 {
		t.Fatalf("reconcile must not mutate currentProcessCounts, got %d", l.currentProcessCounts[zoneProcessName])
	}
}

// TestReconcileMethod_StopTimerSkips verifies reconciliation is skipped while a
// stop/restart timer is pending.
func TestReconcileMethod_StopTimerSkips(t *testing.T) {
	l := newReconcileTestLauncher()
	l.currentZoneProcInfos = []zoneProcInfo{{Pid: 123, Age: 1 * time.Hour}}
	l.SetStopTimer(60)

	l.reconcileStaleZoneProcesses(zoneListWithPids(1))

	if len(l.staleZoneSince) != 0 {
		t.Fatalf("expected no action while stop timer pending, got %v", l.staleZoneSince)
	}
}
