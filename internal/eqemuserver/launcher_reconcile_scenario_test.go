package eqemuserver

import (
	"testing"
	"time"
)

// TestReconcile_ScenarioPoolDrainRecovery simulates the prod failure and the
// launcher Supervisor loop at the decision level:
//   - a dynamic pool is maintained at `target` live dynamics
//   - some children become "stale" (alive but dropped from world's registered
//     list) -- this is exactly what caused "No zoneserver available to boot up"
//   - each Supervisor cycle: reconcile decides kills (capped), killed stale
//     children are reaped, and the boot loop spawns replacements that register
//     after a short boot delay
//
// It asserts the stale pool is fully healed, healthy zones are never killed,
// and the loop converges (no restart loop / no over-kill).
func TestReconcile_ScenarioPoolDrainRecovery(t *testing.T) {
	const target = 8     // maintained dynamic pool size (analogue of minZoneProcesses)
	const staleCount = 5 // children that went stale (alive, unregistered)
	const graceSec = 120
	const perCycleCap = 2
	const bootDelay = 10 * time.Second // freshly-booted zone takes 10s to register

	type child struct {
		pid        int32
		registered bool
		age        time.Duration
	}

	now := time.Unix(1_700_000_000, 0)
	nextPid := int32(1000)
	pendingRegister := map[int32]time.Time{}

	// boot the initial pool: all registered & old
	pool := make([]child, 0, target+staleCount)
	for i := 0; i < target; i++ {
		pool = append(pool, child{pid: nextPid, registered: true, age: 1 * time.Hour})
		nextPid++
	}
	// mark `staleCount` of them stale (world dropped their registration, but the
	// process is still alive -- the prod condition)
	for i := 0; i < staleCount; i++ {
		pool[i].registered = false
	}

	prevSince := map[int32]time.Time{}
	grace := time.Duration(graceSec) * time.Second

	const maxCycles = 400
	for cycle := 0; cycle < maxCycles; cycle++ {
		local := make([]zoneProcInfo, 0, len(pool))
		registered := make(map[int32]bool, len(pool))
		for _, c := range pool {
			local = append(local, zoneProcInfo{Pid: c.pid, Age: c.age})
			if c.registered {
				registered[c.pid] = true
			}
		}

		kill, next, _ := decideStaleZoneKills(local, registered, prevSince, now, grace, perCycleCap)
		prevSince = next
		if len(kill) > perCycleCap {
			t.Fatalf("cycle %d: killed %d > cap %d", cycle, len(kill), perCycleCap)
		}

		killedSet := make(map[int32]bool, len(kill))
		for _, pid := range kill {
			killedSet[pid] = true
			// sanity: never kill a healthy (registered) zone
			if registered[pid] {
				t.Fatalf("cycle %d: healthy zone pid %d was killed", cycle, pid)
			}
		}

		// reap killed children, age the survivors
		survivors := make([]child, 0, len(pool))
		for _, c := range pool {
			if killedSet[c.pid] {
				continue
			}
			c.age += time.Second
			survivors = append(survivors, c)
		}
		pool = survivors

		// boot loop: spawn replacements up to target. New children are alive
		// immediately but register after bootDelay.
		for len(pool) < target {
			pool = append(pool, child{pid: nextPid, registered: false, age: 0})
			pendingRegister[nextPid] = now.Add(bootDelay)
			nextPid++
		}

		// advance simulated time; let pending boots register when due
		now = now.Add(1 * time.Second)
		for i := range pool {
			if !pool[i].registered {
				if due, ok := pendingRegister[pool[i].pid]; ok && !now.Before(due) {
					pool[i].registered = true
					delete(pendingRegister, pool[i].pid)
				}
			}
		}

		// converged when all live dynamics are registered again
		allHealthy := len(pool) == target
		for _, c := range pool {
			if !c.registered {
				allHealthy = false
				break
			}
		}
		if allHealthy {
			t.Logf("recovered full pool of %d registered dynamics after %d simulated cycles "+
				"(%ds grace + gradual capped replenish)", target, cycle+1, graceSec)
			return
		}
	}

	t.Fatalf("did not converge within %d cycles; pool=%v", maxCycles, pool)
}
