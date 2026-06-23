# Launcher Stale-Zone Reconciliation (Self-Healing)

Design note, code summary, verification, and risk assessment for the fix to the
monomyth **"No zoneserver available to boot up"** failure where child zone PIDs
remain alive but are no longer registered with world as usable zone servers.

> Scope: dev-first (`monomyth-dev`). Do **not** deploy to prod until verified on
> dev. This change is independent of the EQEmu-side `KillProcessOnDynamicShutdown`
> setting; it assumes stale child PIDs can occur for any reason.

---

## 1. Root cause

The Spire launcher `Supervisor()` loop decides how many dynamic zone processes to
keep alive using the **local alive process count** only:

```go
// internal/eqemuserver/launcher.go (Supervisor, dynamic boot loop)
for l.getProcessCounts(zoneProcessName)-l.currentZoneStatics < (l.zoneAssignedDynamics + l.minZoneProcesses) {
    l.startServerProcess(zoneProcessName)
}
```

`getProcessCounts(zoneProcessName)` counts zone children found via `gopsutil`
(`pollProcessCounts`) — i.e. **process liveness**. A child that is *alive but no
longer registered with world* (hung in libc malloc/futex, lost its world socket,
reaped from world's list, etc.) still increments this counter, so the launcher
believes the dynamic pool is satisfied and **never boots a replacement**.

Meanwhile world's authoritative `zone_count` keeps draining
(`Removed Zone Server connection ... total zone_count [...]`) until the pool is
exhausted and world logs `No zoneserver available to boot up`. `ps` still shows
the stale `/home/eqemu/server/bin/zone` PIDs under `/tmp/spire-launcher`, so
process liveness is an insufficient health signal.

## 2. Authoritative health signal (chosen)

**World's registered zoneserver list**, obtained in-process via the existing
telnet API the launcher already calls every `Supervisor` cycle:

```
api get_zone_list   ->  WorldZoneList.Data[].zone_os_pid
```

Each entry is exactly what world considers a usable zone server and carries the
zone's **OS PID** (`ZoneOsPid`). This is the single authoritative, in-process
signal of "a usable zone server" — no new credentials, no prod-only access, no
UI interaction, and it is already fetched once per cycle.

Therefore:

> **A launcher-local zone child process is healthy ⇔ its PID appears as a
> `ZoneOsPid` in world's `get_zone_list` response.**
>
> A dynamic zone child that is alive but whose PID is absent from that list for
> longer than the grace period is **stale** and is killed so the normal boot loop
> replaces it.

## 3. Stale-child criteria

A local `zone` child `P` is selected for termination only when **all** hold:

1. `P` is a **dynamic** zone (launched without a static short-name argument).
   Static zones are never killed by this feature (see §6).
2. `P` is **alive** (present in `gopsutil` process list with base name `zone`).
3. `P`'s PID is **absent** from world's registered `ZoneOsPid` set
   (`zone_os_pid > 0`) — and world was reachable this cycle (the zone list was
   fetched successfully; otherwise the pass is skipped entirely).
4. `P` has been **continuously** unregistered for ≥ `staleZoneGraceSeconds`.
5. `P`'s **process age** is ≥ `staleZoneGraceSeconds` (covers slow boots /
   reconnects even if a timer was set at spawn time).

Conditions 4 and 5 together protect: slow-booting zones, transient world list
dips, and brief reconnects.

## 4. Reconciliation loop

Runs once per `Supervisor` cycle, **after** `GetZoneList()` succeeds and
**before** the dynamic boot loop (`internal/eqemuserver/launcher.go`,
`Supervisor()`). Reconciliation only kills stale children; it does **not** mutate
the shared process count (that is owned by `pollProcessCounts`, which resets it
from other call sites — mutating it here would race). The killed child is
recounted by the next cycle's `pollProcessCounts()`, so the existing boot loop
replenishes the pool on the following cycle (~1s later — immaterial vs. a failure
that develops over minutes, and concurrency-safe).

```
GetZoneList() ok? ─ no ─► skip (world down/restarting)            [guardrail]
        │ yes
stop timer pending? ─ yes ─► skip (graceful stop/restart in progress) [guardrail]
        │ no
build registered{ zone_os_pid } from list
registered empty? ─ yes ─► skip+warn (world booting/degraded; would drain pool) [guardrail]
        │ no
snapshot zone children captured by this cycle's pollProcessCounts (no 2nd scan)
decideStaleZoneKills(...)            # pure: grace + per-cycle cap, oldest-stale first
apply rolling 5m kill-window cap      # guardrail against runaway
for each selected pid:
    SIGTERM -> wait <=5s
    re-validate pid identity (base name + create-time) before SIGKILL  [PID-reuse defense]
    SIGKILL if still the same zone process
    broadcast "zoneReconcileKill" to Spire UI
(next Supervisor cycle: pollProcessCounts recounts reality, boot loop replaces)
```

Static/dynamic classification uses one shared helper (`zoneStaticArg`) called by
both `pollProcessCounts` (boot/count path) and reconciliation (kill path), so the
two can never diverge. The zone-child snapshot (`currentZoneProcInfos`) is
captured once per cycle inside `pollProcessCounts` (which already enumerates
processes), so reconciliation adds **no second process-table scan**.

## 5. Guardrails

| Guardrail | Mechanism |
|---|---|
| World down / restarting | `GetZoneList()` failure returns before reconcile; no action. |
| World reachable but empty list | `len(registered) == 0` ⇒ skip + warn (world booting/degraded; prevents pool drain). |
| Stop / restart pending | `GetStopTimer() > 0` ⇒ skip. |
| Booting / reconnecting zone | Grace: requires both age ≥ grace and continuous-unregistered ≥ grace. |
| Healthy zone with clients | Stays in world's registered list ⇒ never selected. |
| Runaway kills | Per-cycle cap (`maxStaleZoneKillsPerCycle`, default 2) **and** rolling 5-minute window cap (`perCycleCap × 5`, default 10). |
| PID reuse | Identity captured at entry (base name + create-time); re-validated immediately before SIGKILL; aborts if the PID no longer refers to the same zone process. |
| Force-kill of wrong process | SIGTERM first, wait ≤5s, then SIGKILL only after identity re-check passes. |
| Restart-loop avoidance | Reconcile never mutates the shared process count; replacements come from the existing boot loop on the next cycle after `pollProcessCounts` recounts reality. Per-cycle + rolling-window caps bound kills; failed kills retain the timer for prompt retry. |
| Double-spawn / count corruption | Reconcile does not touch `currentProcessCounts` (owned/reset by `pollProcessCounts`), so there is no race with a concurrent poll reset and no over-spawn. |
| Static zones | Never killed in v1 (reported at debug only); static replacement remains the static boot loop's responsibility. |

## 6. Static-zone behavior (unchanged / conservative)

Static zones (launched as `zone <shortname>`, configured via `staticZones`) are
**excluded** from kill selection. A stale static is logged at `Debug`
("static zone alive but not registered with world (not killed; static
reconciliation disabled)") so it is observable without taking action. Static
replacement continues to be driven by the existing static boot loop. Enabling
static kill is a deliberate, separate follow-up (would need its own conservative
criteria, e.g. zero players).

## 7. Tunables (`eqemu_config.json` → `web-admin.launcher`)

All have safe in-memory defaults; **no config file rewrite is forced** (defaults
apply when keys are absent). Add keys only to override.

| Key | Type | Default | Meaning |
|---|---|---|---|
| `reconcileStaleZones` | bool | `true` | Master switch. |
| `staleZoneGraceSeconds` | int | `120` | Grace before a dynamic child is killed; also min age. |
| `maxStaleZoneKillsPerCycle` | int | `2` | Max kills per Supervisor pass. |

Example (dev) override to make a reproduction faster:

```jsonc
"web-admin": { "launcher": {
  "reconcileStaleZones": true,
  "staleZoneGraceSeconds": 30,
  "maxStaleZoneKillsPerCycle": 2
}}
```

## 8. Code summary (diff)

Files changed/added:

- `internal/eqemuserverconfig/config.go` — 3 tunable fields on
  `WebAdminLauncherConfig` (`ReconcileStaleZones *bool`,
  `StaleZoneGraceSeconds int`, `MaxStaleZoneKillsPerCycle int`).
- `internal/eqemuserver/launcher.go` — launcher state (timers, kill window,
  mutex) + init + tunable loading in `loadServerConfig()` + one call to
  `reconcileStaleZoneProcesses(list)` in `Supervisor()`.
- `internal/eqemuserver/launcher_reconcile.go` (new) — pure
  `decideStaleZoneKills`, `reconcileStaleZoneProcesses`,
  `terminateZoneProcess`, `zoneProcessStaticName`, `broadcastReconcileKill`.
- `internal/eqemuserver/launcher_reconcile_test.go` (new) — 7 decision tests.
- `internal/eqemuserver/launcher_reconcile_scenario_test.go` (new) — multi-cycle
  convergence test simulating the prod drain.

## 9. Verification

### 9.1 Build / vet / unit tests (runnable without an EQEmu stack)

```bash
cd /home/zutfen/code/spire
go build ./...
go vet ./internal/eqemuserver/ ./internal/eqemuserverconfig/
go test -count=1 -v -run TestReconcile ./internal/eqemuserver/
```

Result (actual, 11 tests):

```
=== RUN   TestReconcile_ScenarioPoolDrainRecovery
    recovered full pool of 8 registered dynamics after 132 simulated cycles (120s grace + gradual capped replenish)
--- PASS: TestReconcile_ScenarioPoolDrainRecovery (0.00s)
=== RUN   TestReconcile_KillsStaleDynamicWithinGrace            --- PASS
=== RUN   TestReconcile_DoesNotKillBootingZone                 --- PASS
=== RUN   TestReconcile_StaticZonesNeverKilled                 --- PASS
=== RUN   TestReconcile_PerCycleCap                            --- PASS
=== RUN   TestReconcile_HealedZoneClearsTimer                  --- PASS
=== RUN   TestReconcile_GonePidPruned                          --- PASS
=== RUN   TestReconcile_RequiresBothAgeAndUnregistered        --- PASS
=== RUN   TestReconcileMethod_EmptyWorldListDefers            --- PASS   (deploy-safety: empty world list => defer)
=== RUN   TestReconcileMethod_NoDecrementWhenKillFails         --- PASS   (count not decremented on failed kill)
=== RUN   TestReconcileMethod_StopTimerSkips                   --- PASS   (no action while stop timer pending)
PASS
ok  	github.com/EQEmu/spire/internal/eqemuserver
```

`go build ./...` ⇒ OK. `go vet` reports only pre-existing warnings in
`launcher_cmd.go` (unexported fields with json tags); none in changed files.

The scenario test proves the prod failure is healed at the algorithm level: a
pool of 8 with 5 silently-stale children fully recovers to 8 registered
dynamics, healthy zones are never killed, and the cap prevents restart loops.

### 9.2 Dev reproduction on `monomyth-dev` (operator runbook)

Build + install the changed launcher on dev (do **not** touch prod):

```bash
# on the dev host, in the spire checkout
GOOS=linux GOARCH=amd64 go build -o spire-linux-amd64
# install per your dev layout, e.g.:
cp spire-linux-amd64 /opt/spire/spire
```

Lower the grace to make the test observable in minutes (dev only):

```jsonc
// eqemu_config.json  ->  web-admin.launcher
"staleZoneGraceSeconds": 30,
"maxStaleZoneKillsPerCycle": 2
```

Start the launcher:

```bash
/tmp/spire-launcher eqemu-server:launcher start   # or your dev invocation
```

Create a **stale dynamic zone** (alive but unregistered) on dev — pick the
lowest-risk reproducer available:

- **Option A (preferred, deterministic):** identify a dynamic zone pid that is
  *not* serving clients, then detach it from world without killing the process:
  ```bash
  # find a dynamic (arg-less) zone pid not in world's registered list
  ZPID=$(pgrep -f '/bin/zone$' | head -n1)
  # sever its world socket so world drops the registration but the proc lives
  gdb -p "$ZPID" -batch -ex 'call close(<its world tcp fd>)' -ex detach
  ```
  (If `gdb` fd-closing is impractical, Option B.)
- **Option B (known eqemu path):** reproduce via the eqemu dynamic-shutdown idle
  path that leaves a child alive-but-unregistered, with a short idle timer, and
  confirm via `api get_zone_list` that the pid disappears while `ps` still lists
  it.
- **Option C (simulate):** `kill -STOP <zone pid>` to freeze a zone so world
  drops it as unresponsive; reconciliation should still detect+replace it.

Observe reconciliation acting (tail launcher logs):

```
... Zone reconcile: dynamic zone(s) alive but not registered with world, grace timer started   count=1 grace_seconds=30
... Zone reconcile: killing stale dynamic zone (alive but not registered with world)  pid=12345 age=12m0s unregistered_for=31s grace=30s
... Zone reconcile: stale zone terminated gracefully   pid=12345
... Zone reconcile: cycle summary   eligible=1 killed=1 deferred=0 tracked=0
... Starting Dynamic Zone   bootedTotalDynamics=... targetDynamics=...
```

Then confirm recovery end-to-end:

```bash
# world received a fresh New Zone Server connection and zone_count returned to floor
echo "api get_server_counts" | nc <world telnet ip> <port>      # zone_count back to expected
# the stale pid is gone and a new zone pid exists
ps -ef | grep '[/]bin/zone' | wc -l
```

Expected across several cycles: no repeat kills of the same pid, no kill of
healthy/static zones, `zone_count` returns to the configured floor and stays
there.

### 9.3 Normal-case checks (dev)

- **Active zone with clients:** a registered dynamic with players is in
  `get_zone_list` ⇒ never selected. Verify no kill log for it.
- **Static zone (e.g. bazaar):** never killed; only a `Debug` stale notice if
  unregistered.
- **Slow booting zone:** a freshly-spawned zone younger than grace is not killed
  even while unregistered (covered by `TestReconcile_DoesNotKillBootingZone`).
- **World restart:** while world is down, `GetZoneList()` fails ⇒ reconcile is
  skipped (no mass unsafe kills). Verify during a dev world bounce.

## 10. Risk assessment

| Risk | Likelihood | Mitigation |
|---|---|---|
| Killing a healthy zone due to a transient/partial world list | Low | Requires continuous unregistration ≥ grace **and** age ≥ grace; rolling kill cap limits blast radius; world-unreachable ⇒ skip. |
| World reachable but returns empty/partial list | Low | Explicit `len(registered) == 0` guard defers reconcile (no timers started) until world reports ≥1 usable zone. Partial-list residual risk remains mitigated by grace + caps. |
| Killing a zone with active clients | Very low | Such zones stay registered; grace gives time to re-register. Cannot be fully excluded without per-zone player data when unregistered — accepted, documented. |
| Restart loop / thrash / count drift | Very low | Reconcile never mutates the shared `currentProcessCounts` (which `pollProcessCounts` resets from other call sites), so there is no TOCTOU with a concurrent poll reset; replacements come from the boot loop on the next cycle. Per-cycle + rolling-window caps bound kills. |
| PID reuse / killing wrong process | Very low | Identity (base name + create-time) captured at entry and re-validated immediately before SIGKILL; aborts if the PID no longer maps to the same zone process. |
| Accidental prod config rewrite | Low | Tunables default in-memory; no `Save` triggered for them. |
| Distributed / leaf nodes | N/A for v1 | Reconciliation wired into the non-distributed `Supervisor` path (the prod failure mode). Distributed/leaf reconciliation is a documented follow-up. |
| Static zones drained | None | Statics are explicitly excluded from kill selection in v1. |

## 11. Follow-ups (out of scope for this change)

- Reconciliation on distributed root / leaf nodes (leaf hosts zones; needs
  address-filtered registered set per node).
- Optional static-zone reconciliation behind a separate conservative flag
  (zero-players + long grace).
- Exposing tunables in the Spire launcher-options admin UI.
