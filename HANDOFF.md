# HANDOFF

This file is the single source of truth for cross-session handoff during the Option 2 migration work.

Update this file after each completed step.

## Update Rules

When a step is completed, update all of the following before ending the session:

1. `Current Status`
2. `Completed Steps`
3. `Verification`
4. `Next Step`
5. `Open Risks / Warnings`

Keep entries short, factual, and action-oriented.

## Current Status

- Date: `2026-06-17`
- Branch: `feature/option-2-migration-foundation`
- Goal: implement Option 2 as an in-place migration to `Vue 3 + Vite + Pinia` while keeping the Go backend and HTTP API stable
- State: frontend foundation is migrated and building; SPA packaging moved from `packr` to `go:embed`; Wire-based bootstrap removed in favor of explicit constructor composition; connections/user modal flows have started moving onto the local Vue 3 modal bridge

## Completed Steps

### 1. Frontend Foundation

- Replaced Vue CLI build/dev flow with Vite in [frontend/package.json](/home/zutfen/code/spire/frontend/package.json) and [frontend/vite.config.ts](/home/zutfen/code/spire/frontend/vite.config.ts)
- Added Vite entry HTML in [frontend/index.html](/home/zutfen/code/spire/frontend/index.html)
- Moved app boot to Vue 3 in [frontend/src/main.ts](/home/zutfen/code/spire/frontend/src/main.ts)
- Upgraded router boot to Vue Router 4 in [frontend/src/router.ts](/home/zutfen/code/spire/frontend/src/router.ts)
- Added Vue compatibility/runtime bridge pieces:
  - [frontend/src/plugins/legacy-bootstrap.ts](/home/zutfen/code/spire/frontend/src/plugins/legacy-bootstrap.ts)
  - [frontend/src/components/runtime/TrustedHtml.vue](/home/zutfen/code/spire/frontend/src/components/runtime/TrustedHtml.vue)
  - [frontend/src/app/event-bus/event-bus.ts](/home/zutfen/code/spire/frontend/src/app/event-bus/event-bus.ts)
  - [frontend/src/app/env/runtime-env.ts](/home/zutfen/code/spire/frontend/src/app/env/runtime-env.ts)
  - [frontend/src/stores/app.ts](/home/zutfen/code/spire/frontend/src/stores/app.ts)
  - [frontend/src/stores/session.ts](/home/zutfen/code/spire/frontend/src/stores/session.ts)

### 2. Low-Risk Route / Shell Compatibility

- Patched boot-critical and low-risk screens to work under Vite/Vue 3:
  - [frontend/src/App.vue](/home/zutfen/code/spire/frontend/src/App.vue)
  - [frontend/src/views/Login.vue](/home/zutfen/code/spire/frontend/src/views/Login.vue)
  - [frontend/src/views/Home.vue](/home/zutfen/code/spire/frontend/src/views/Home.vue)
  - [frontend/src/views/Doc.vue](/home/zutfen/code/spire/frontend/src/views/Doc.vue)
  - [frontend/src/views/server-developer/Releases.vue](/home/zutfen/code/spire/frontend/src/views/server-developer/Releases.vue)
  - [frontend/src/views/admin/server-update/UpdateReleases.vue](/home/zutfen/code/spire/frontend/src/views/admin/server-update/UpdateReleases.vue)
  - [frontend/src/components/layout/Navbar.vue](/home/zutfen/code/spire/frontend/src/components/layout/Navbar.vue)
  - [frontend/src/components/layout/docs/DocNavbar.vue](/home/zutfen/code/spire/frontend/src/components/layout/docs/DocNavbar.vue)
  - [frontend/src/components/modals/AppUpdateModal.vue](/home/zutfen/code/spire/frontend/src/components/modals/AppUpdateModal.vue)

### 3. Frontend Build Compatibility Cleanup

- Added browser-safe replacements for legacy Node-style imports used by the old frontend
- Modernized debounce export compatibility in [frontend/src/app/utility/debounce.js](/home/zutfen/code/spire/frontend/src/app/utility/debounce.js)
- Kept the generated/build output path at `frontend/dist`

### 4. SPA Packaging Migration

- Replaced runtime `packr` usage with `go:embed`
- Added embedded frontend filesystem in [frontend/embed.go](/home/zutfen/code/spire/frontend/embed.go)
- Switched SPA serving to embedded `fs.FS` in:
  - [internal/http/spa/spa.go](/home/zutfen/code/spire/internal/http/spa/spa.go)
  - [internal/http/spa/packer.go](/home/zutfen/code/spire/internal/http/spa/packer.go)
- Removed `packr` from active build/release paths:
  - [Makefile](/home/zutfen/code/spire/Makefile)
  - [windows-build-spire-release.exe.bat](/home/zutfen/code/spire/windows-build-spire-release.exe.bat)
  - [containers/workspace/Dockerfile](/home/zutfen/code/spire/containers/workspace/Dockerfile)
  - [containers/prod/Dockerfile](/home/zutfen/code/spire/containers/prod/Dockerfile)
  - [scripts/build-release.sh](/home/zutfen/code/spire/scripts/build-release.sh)

### 5. Wire Removal

- Promoted the generated Wire bootstrap into normal source in [boot/wire_gen.go](/home/zutfen/code/spire/boot/wire_gen.go)
- Deleted [boot/wire.go](/home/zutfen/code/spire/boot/wire.go)
- Removed Wire-only provider set scaffolding from the remaining boot files
- Deleted now-empty Wire helper files:
  - [boot/inject_database_resolver.go](/home/zutfen/code/spire/boot/inject_database_resolver.go)
  - [boot/inject_encryption.go](/home/zutfen/code/spire/boot/inject_encryption.go)
  - [boot/inject_logger.go](/home/zutfen/code/spire/boot/inject_logger.go)
  - [boot/inject_services.go](/home/zutfen/code/spire/boot/inject_services.go)
- Updated the CRUD controller template in [internal/model/templates/inject_http_crud_controller.tmpl](/home/zutfen/code/spire/internal/model/templates/inject_http_crud_controller.tmpl)
- Updated contributor docs in [README.md](/home/zutfen/code/spire/README.md)

### 6. Phase 2 Modal Flow Compatibility

- Fixed the local modal bridge to honor `#modal-header` slots in [frontend/src/plugins/legacy-bootstrap.ts](/home/zutfen/code/spire/frontend/src/plugins/legacy-bootstrap.ts)
- Reworked connection modal open flows to use `nextTick()` and direct `$bvModal.show(...)` calls instead of timing-sensitive directive registration in [frontend/src/views/connections/UserConnections.vue](/home/zutfen/code/spire/frontend/src/views/connections/UserConnections.vue)
- Added safer empty-object defaults for conditional modal props in:
  - [frontend/src/views/connections/AddDeveloperModal.vue](/home/zutfen/code/spire/frontend/src/views/connections/AddDeveloperModal.vue)
  - [frontend/src/views/connections/ManageDiscordConnectionModal.vue](/home/zutfen/code/spire/frontend/src/views/connections/ManageDiscordConnectionModal.vue)
  - [frontend/src/views/user/ResetUserPasswordModal.vue](/home/zutfen/code/spire/frontend/src/views/user/ResetUserPasswordModal.vue)
- Removed temporary watcher / force-update noise from the connections route shell in [frontend/src/views/connections/UserConnections.vue](/home/zutfen/code/spire/frontend/src/views/connections/UserConnections.vue)

## Verification

Last verified successfully:

- `go build`
- `go build ./internal/http/spa`
- `cd frontend && npm run build`

## Open Risks / Warnings

- Frontend build still emits non-fatal warnings:
  - deprecated Sass legacy JS API
  - deprecated Vue deep selector syntax (`>>>` / `/deep/`)
  - existing duplicate `case 503` warning in `frontend/src/app/spells.ts`
- Connections/user modal fixes are build-verified but not yet browser-smoke-tested end-to-end
- Large editor-heavy routes are not yet intentionally migrated; current success is foundation-first
- Vue 2 specialty libraries are still present as dependency debt even though the app now builds on the new shell
- `docs/project-assessment-2026-06.md` still references Wire historically; that is acceptable unless we want the assessment updated to reflect implementation progress

## Next Step

Recommended next phase:

- Continue Phase 2 frontend migration: finish the remaining admin/shared modal-heavy routes
- Goal: replace real BootstrapVue usage on ordinary CRUD/admin flows with the new local Vue 3 wrapper layer and remove timing-sensitive modal behavior

Suggested first targets:

- `frontend/src/views/admin/components/ServerProcessButtonComponent.vue`
- `frontend/src/views/admin/player-event-logs/PlayerEventLogs.vue`
- `frontend/src/views/admin/*` routes with modal/tabs/pagination usage

## Session Notes

- Do not revert unrelated user changes
- Keep `frontend/dist` as the backend-consumed artifact shape
- Preserve existing HTTP API contracts, route paths, auth redirects, and query-string behavior
