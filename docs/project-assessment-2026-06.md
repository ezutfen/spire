# Spire Project Assessment

Date: June 16, 2026

## Executive Summary

### Verdict

Spire is still worth investing in, but not as a "keep doing the same thing" codebase.

The product direction remains coherent: a rich EverQuest Emulator toolkit with a strong local-first workflow, a hosted mode, deep domain-specific editing, and a large amount of generated CRUD surface. That is valuable and should be preserved. The main problem is that the delivery layer around that value has aged faster than the domain layer itself.

The backend is not the first place I would rewrite. The frontend and build/tooling boundary are. The project is carrying too much risk in Vue 2, Vue CLI, BootstrapVue, legacy asset packaging, dated auth patterns, and brittle quality gates. Those issues do not mean the product is unsalvageable. They do mean the current stack is increasingly expensive to trust, upgrade, and contribute to.

### Overall Recommendation

Keep the core product direction and the Go backend/domain model. Modernize aggressively at the frontend and tooling boundary first. Do not start with a backend rewrite or a microservices split.

### Top 5 Risks

1. **Frontend stack obsolescence**
   The SPA still runs on Vue 2, Vue CLI 3-era tooling, BootstrapVue, class-component patterns, and `--openssl-legacy-provider`. Vue 2 reached end of life on December 31, 2023, and Vue CLI is in maintenance mode. This is the largest concentration of future breakage risk.

2. **Dependency and security debt**
   `npm audit --omit=dev` already reports material production dependency risk, including 17 total production vulnerabilities: 1 critical, 4 high, 5 moderate, and 7 low. The full audit is much worse at 153 total vulnerabilities: 10 critical, 53 high, 72 moderate, and 18 low. The old toolchain is pulling in a large vulnerable dependency tree.

3. **Weak quality gates**
   `go test ./...` does not pass cleanly. Some failures are environment-bound, but others are basic build/vet or test failures:
   - `internal/banner` fails vet because of `fmt.Println` usage with a redundant newline.
   - `internal/eqtraders` fails vet because `fmt.Println` is being used like `Printf`.
   - `internal/integration` assumes a live database.
   - `internal/structs` has real failing tests.
   This means green-ness is not a trustworthy signal right now.

4. **Legacy runtime/build plumbing**
   The project still depends on `packr` for SPA packaging and `google/wire` for dependency injection codegen. `google/wire` was archived and made read-only on August 25, 2025. These are not urgent production outages, but they are clear maintenance liabilities.

5. **High-change-cost UI hotspots**
   Several hand-written frontend modules are very large and likely difficult to change safely. Examples include:
   - `frontend/src/views/spells/SpellEditor.vue`
   - `frontend/src/views/tasks/TaskEditor.vue`
   - `frontend/src/views/items/ItemEditor.vue`
   - `frontend/src/app/spells.ts`
   These are not just "big files"; they are likely where future modernization cost accumulates.

## Architecture and Maintainability Review

### What the Project Is

Spire is not a simple web app. It is a mixed-mode product with:

- A Go monolith that serves HTTP APIs and the frontend SPA.
- A local desktop-like executable path that launches itself for local use.
- A hosted/web mode with auth and connection management.
- Dockerized development and release flows.
- Heavy code generation on both backend and frontend.

That complexity is partly justified. The product genuinely needs deep domain behavior, local deployment friendliness, and a broad editing surface. The problem is not that Spire has many capabilities. The problem is that too many concerns are collapsed into one delivery stack with aging tools.

### What Looks Structurally Sound

- **Go monolith as product core**: still a reasonable shape for this product. The app is domain-rich but not obviously at a scale where service decomposition would pay for itself.
- **Generated CRUD surface**: reasonable in principle for a schema-heavy editor product. The large generated API/model surface is a feature of the product, not automatically a flaw.
- **Local-first workflow**: still a strong differentiator and worth protecting.

### What Looks Expensive to Maintain

- **Frontend runtime and tooling**
  - Vue 2.7.x
  - Vue Router 3
  - Vuex 3
  - BootstrapVue 2
  - Vue CLI 3-era structure
  - class decorators and legacy component patterns

  This is a lot of legacy surface to carry at once. Any one of these could be manageable; all of them together create compounding upgrade drag.

- **Monolith plus delivery coupling**
  The same application layer owns API serving, desktop boot, SPA packaging, websocket startup, reverse proxy behavior, and other operational concerns. That increases blast radius for routine changes.

- **Legacy packaging**
  The backend still uses `packr` to package SPA assets. That is no longer a good fit now that Go has `embed`.

- **Generated DI**
  Wire still works for now, but it no longer has an active upstream. For a codebase of this size, explicit constructors are now the safer long-term path.

### Oversized Hotspots

The codebase has a lot of generated surface, but there are also genuine hand-written hotspots that deserve attention. The generated footprint is large enough that modernization should preserve boundaries instead of trying to rewrite everything at once:

- roughly 464 files under `frontend/src/app/api`
- roughly 498 files across `internal/http/crudcontrollers` and `internal/models`

Alongside that generated surface, there are hand-written hotspots that deserve attention:

- `frontend/src/app/spells.ts` is especially large and likely mixes domain translation, display logic, and editor behavior.
- `frontend/src/views/spells/SpellEditor.vue`, `TaskEditor.vue`, and `ItemEditor.vue` are very large UI surfaces with likely high regression risk.
- `internal/eqemuserver/installer.go` is large enough to deserve decomposition even if the surrounding backend remains a monolith.
- `internal/eqtraders/scrape_cmd.go` is both large and currently noisy enough to trip vet, which is a bad sign for maintainability.

### Justified Complexity vs Accidental Complexity

**Justified**

- Rich domain editors for EQEmu content.
- Generated API/model surface for large schema coverage.
- Local and hosted usage paths.
- Docker-backed dev environment for reproducibility.

**Accidental**

- Keeping the SPA on multiple legacy frontend layers at once.
- Relying on both outdated packaging and outdated DI codegen.
- Quality checks that fail for avoidable reasons.
- Old CI/release assumptions still reflected in files like `.drone.yml` and release scripts.

## Quality, Security, and Operability Review

### Test Reliability

Current test reliability is weak.

- `go test ./...` is not a dependable pass/fail signal.
- Some failures are infrastructure assumptions rather than product regressions.
- Some failures are real code/test health problems.

The most important conclusion is not "there are some failing tests." It is that maintainers cannot trust the default test command as a clean gate today.

### CI/CD and Build Health

- CI is based on Drone via `.drone.yml`, which suggests an older release pipeline shape.
- Release automation still centers on Dockerized builds plus manual-ish version checks and `gh-release`.
- The workspace container mixes Go, Node, Java, release tooling, `packr`, `wire`, and global npm utilities in one image.

This is workable, but it is not lean, modern, or especially reproducible. It also increases the chance that build stability depends on container drift rather than repository truth.

### Dependency Health

Frontend dependency health is the project’s sharpest operational risk.

- `npm audit --omit=dev` shows 17 production vulnerabilities.
- Full `npm audit` shows 153 total vulnerabilities driven largely by old Vue CLI/Webpack-era tooling.
- The workspace currently does not have frontend dependencies installed locally, so `npm ls --depth=0` reports unmet dependencies. That does not prove the lockfile is wrong, but it does confirm contributor setup friction and limits quick local validation.

Backend dependency health is better, but not clean:

- `dgrijalva/jwt-go` is still present in auth code even though `golang-jwt/jwt` is also in the repo.
- Archived or legacy infrastructure packages remain in active use (`wire`, `packr`).

### Auth and Session Handling

Auth deserves a modernization pass even if the backend is not rewritten.

Observed concerns:

- Custom JWT creation/validation paths.
- Long-lived JWT expiration policy.
- Mixed dependency usage around JWT libraries.
- Cache-based user context behavior.
- Hosted-mode auth layered into a codebase that also defaults to local/desktop behavior.

Nothing here proves an immediate exploit in Spire itself, but it is dated enough that I would treat auth hardening as part of stabilization, not as optional cleanup.

### Packaging and Runtime Delivery

The project is still carrying a pre-`embed` style asset story:

- Build frontend separately.
- Use `packr` to package assets into the binary.
- Also download large asset bundles externally during install/setup.

That is more moving pieces than necessary. This should be simplified. The right medium-term move is to use Go `embed` for the SPA bundle and keep large external content assets explicitly separate.

### Local Development Friction

Local development looks heavier than it should be for contributors:

- Multiple make targets and Docker assumptions.
- Database seed steps.
- Asset download steps.
- Frontend setup divergence from the rest of modern Vue tooling.
- Incomplete trust in tests as a quick validation step.

This is manageable for a core maintainer, but it raises the onboarding cost for occasional contributors.

## Stack Pivot Recommendations

### Option 1: Conservative

**Keep Go + Echo monolith. Replace only the frontend build/runtime stack incrementally.**

What this means:

- Keep the current API contract.
- Keep the backend structure mostly intact.
- Migrate frontend tooling first: Vue CLI to Vite, keep Vue-based UI, start isolating large screens.
- Delay `packr` and Wire removal until after the frontend is stabilized.

Pros:

- Lowest disruption.
- Fastest route to better build speed and lower frontend risk.
- Preserves most current behavior.

Cons:

- Leaves backend/tooling debt in place longer.
- May create a partially modern frontend still constrained by old backend composition and packaging.

Best when:

- The priority is short-term risk reduction with minimal product churn.

### Option 2: Recommended

**Keep Go backend and domain model. Migrate frontend to Vue 3 + Vite + Pinia, replace the BootstrapVue path, remove `packr`, remove Wire, keep HTTP API stable.**

What this means:

- Preserve the monolith and domain services.
- Treat the HTTP API as a compatibility boundary.
- Move from Vue 2 + Vue CLI + Vuex + class components to Vue 3 + Vite + Pinia + Composition API.
- Replace BootstrapVue with a Vue 3-compatible component strategy.
- Replace `packr` with Go `embed`.
- Replace Wire with explicit constructors/manual composition.

Pros:

- Best balance of payoff and containment.
- Removes the highest concentration of future maintenance risk.
- Modernizes contributor experience without rewriting the domain core.
- Protects the existing API so migration can be phased screen by screen.

Cons:

- Requires meaningful UI migration work.
- Large editor screens will take deliberate refactoring, not just mechanical framework conversion.

Best when:

- The goal is to keep Spire viable for several more years without a full rewrite.

### Option 3: Bold

**Keep domain logic in Go, but separate product layers more clearly and turn SPA delivery/desktop concerns into thinner adapters.**

What this means:

- Keep the core business/domain logic in Go.
- Split app boot/runtime responsibilities more explicitly:
  - domain services
  - HTTP API delivery
  - desktop/local launcher concerns
  - hosted/auth concerns
  - asset serving/packaging
- Possibly move the frontend into a more independent app lifecycle while still shipping with the product.

Pros:

- Cleanest long-term architecture.
- Reduces coupling between product modes.
- Makes future hosted/local evolution easier.

Cons:

- More expensive than needed right now.
- Risk of over-architecting before the frontend/tooling debt is retired.
- Can distract from the highest-value modernization work.

Best when:

- The product is expected to grow materially in hosted mode, team size, or operational complexity.

### Recommendation

Choose **Option 2**.

It addresses the highest-risk surface first, preserves the domain investment, and avoids the trap of rewriting the most stable part of the system before the least stable part.

## Prioritized Roadmap

### Next 30 Days

- `must` Make `go test ./...` honest again.
  - Fix the `internal/banner` vet failure.
  - Fix the `internal/eqtraders` vet failures.
  - Separate DB-dependent integration tests from the default unit-test path.
  - Triage and fix or quarantine the failing `internal/structs` tests.

- `must` Establish a frontend dependency baseline.
  - Perform a clean install in a reproducible environment.
  - Confirm whether the lockfile is still healthy.
  - Record a known-good Node/npm version.

- `must` Stabilize auth dependencies.
  - Standardize on one JWT library.
  - Review token lifetime and hosted-mode auth assumptions.

- `must` Document the current architecture and runtime modes.
  - One short maintainer document is enough.
  - Focus on local mode, hosted mode, SPA packaging, and codegen boundaries.

- `should` Add a lightweight CI gate that separates:
  - unit tests
  - integration tests
  - frontend install/build
  - generation drift checks

### Next 90 Days

- `must` Migrate the frontend build from Vue CLI to Vite.
- `must` Start the Vue 3 migration with a compatibility-preserving screen strategy.
- `must` Replace BootstrapVue usage with a Vue 3-supported path.
- `must` Introduce Pinia and begin retiring Vuex 3.
- `should` Break up the largest editor and utility modules before or during migration.
- `should` Replace `packr` with Go `embed` for the SPA bundle.
- `should` Replace Wire with explicit constructor wiring.
- `should` Reduce release image/tooling sprawl in the workspace container and build scripts.

### Longer-Term Optional Pivots

- `optional` Separate desktop/local launch concerns from hosted/web concerns more explicitly.
- `optional` Introduce a cleaner internal app layering for boot/runtime ownership.
- `optional` Revisit whether some scrapers/installers belong in separate packages or tools.
- `optional` Reassess whether the hosted experience should remain in the same deployment unit once the frontend modernization is complete.

## Direct Answers to Key Decision Questions

### Can We Keep This Project as a Single Go App?

Yes, with conditions.

Conditions:

- Modernize the frontend/tooling stack.
- Simplify asset packaging.
- Make tests trustworthy.
- Reduce coupling in app boot/runtime responsibilities over time.

I would not split this into microservices now.

### Should We Rewrite the Frontend?

Yes, but as a phased migration, not a greenfield redesign.

The payoff is high because the current frontend stack is the densest source of obsolescence risk. The right move is to preserve the existing API and migrate screen by screen, starting with infrastructure and shared patterns before the largest editors.

### Should We Split Services?

Not now.

There is not enough evidence that operational scale or team topology justifies microservices. The current problems are mostly tooling, packaging, coupling, and maintainability inside one product, not service-boundary problems.

### What Should Be Fixed First for Safer Ongoing Contribution?

1. Restore trust in the default test/build path.
2. Stabilize frontend install/build reproducibility.
3. Standardize and harden auth dependencies.
4. Begin frontend tooling migration.
5. Remove `packr` and Wire after the new frontend path is established.

### What Should Not Be Rewritten Yet?

- Core Go domain logic.
- Generated CRUD/backend schema coverage model.
- The current HTTP API contract, unless a later migration phase makes a deliberate breaking-change decision.
- The local-first product concept.

## Protected Areas During Modernization

These areas should be preserved while modernizing:

- Backend API shape and route semantics.
- Existing EQEmu domain behaviors.
- Local executable workflow.
- Generated schema coverage strategy, unless replaced with an equally comprehensive system.

## Validation Basis

This assessment is based on direct inspection of the repository plus targeted validation checks:

- Repo structure and manifests:
  - Go module at `go.mod`
  - frontend manifests at `frontend/package.json` and `frontend/package-lock.json`
  - Docker/dev/release files including `Makefile`, `.drone.yml`, `containers/workspace/Dockerfile`, and `scripts/build-release.sh`
- Entrypoints and runtime shape:
  - `main.go`
  - `boot/app.go`
  - `internal/http/http.go`
  - `internal/http/spa/*.go`
  - `internal/auth/controller.go`
  - `internal/user/context_middleware.go`
- Quality checks:
  - `go test ./...`
  - `cd frontend && npm audit --omit=dev --json`
  - `cd frontend && npm audit --json`
  - `cd frontend && npm ls --depth=0`
- Codebase hotspot inspection:
  - large hand-written frontend editors and utility files
  - large backend installer/scraper files
  - generated API/model surfaces

## External References

- Vue 2 EOL announcement: <https://v2.vuejs.org/eol/>
- Vue CLI maintenance mode: <https://cli.vuejs.org/>
- BootstrapVue Vue 3 transition guidance: <https://bootstrap-vue.org/vue3/>
- `google/wire` archived repository status: <https://github.com/google/wire/pulls>

## Final Recommendation

Spire should be treated as a valuable product with an aging shell, not as a failed architecture.

The right move is to preserve the backend and domain investment, modernize the frontend and build boundary aggressively, restore trust in tests and releases, and only then decide whether deeper runtime separation is worth the extra cost.
