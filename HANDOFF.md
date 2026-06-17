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
- State: frontend foundation is migrated and building; SPA packaging moved from `packr` to `go:embed`; Wire-based bootstrap removed in favor of explicit constructor composition; connections/user modal flows and the current admin update/configuration slices now run on the Vue 3-safe path; admin dashboard/zone/timer lifecycle hooks and removed-API (`$set`/`.native`) usages migrated off Vue 2 conventions; full Vue 2 lifecycle-hook/`$set`/`.native` sweep complete across editor-heavy and shared routes (0 remaining occurrences)

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

### 7. Admin Modal / Pagination Runtime Cleanup

- Switched local EQ modal teardown to Vue 3 lifecycle hooks in [frontend/src/components/eq-ui/EQModal.vue](/home/zutfen/code/spire/frontend/src/components/eq-ui/EQModal.vue)
- Removed the last dead BootstrapVue modal event and fixed delayed stop/restart notification handling in [frontend/src/views/admin/components/ServerProcessButtonComponent.vue](/home/zutfen/code/spire/frontend/src/views/admin/components/ServerProcessButtonComponent.vue)
- Updated the player event log explorer to:
  - keep `requesting` reactive and always released after request failures
  - restart its auto-refresh timer when the selected interval changes
  - use Vue 3 lifecycle hooks and modern Highlight.js element highlighting
  in [frontend/src/views/admin/player-event-logs/PlayerEventLogs.vue](/home/zutfen/code/spire/frontend/src/views/admin/player-event-logs/PlayerEventLogs.vue)

### 8. Admin Form / Log Viewer Compatibility

- Replaced unsupported `b-form-tags` / `b-form-tag` / `b-form-select` usage with a native Vue 3-safe static-zone tag picker in [frontend/src/views/admin/components/LauncherOptions.vue](/home/zutfen/code/spire/frontend/src/views/admin/components/LauncherOptions.vue)
- Kept launcher option defaults and prop-sync behavior centralized in the same component so server config updates still post through the existing API
- Switched the file log viewer teardown to Vue 3 lifecycle hooks and clear timer state explicitly in [frontend/src/views/admin/FileLogs.vue](/home/zutfen/code/spire/frontend/src/views/admin/FileLogs.vue)

### 9. Admin Update / Configuration Route Compatibility

- Replaced unsupported `b-select :options` usage with a native select in [frontend/src/views/admin/server-update/ServerUpdate.vue](/home/zutfen/code/spire/frontend/src/views/admin/server-update/ServerUpdate.vue)
- Switched that route to Vue 3 lifecycle hooks, removed dead ANSI-conversion setup, and made build/clean/cancel stream status handling safer when fetch requests fail in [frontend/src/views/admin/server-update/ServerUpdate.vue](/home/zutfen/code/spire/frontend/src/views/admin/server-update/ServerUpdate.vue)
- Added missing row keys, centralized in-game log reload behavior, and made webhook list loading resilient to API failures in [frontend/src/views/admin/configuration/DiscordWebhooks.vue](/home/zutfen/code/spire/frontend/src/views/admin/configuration/DiscordWebhooks.vue)

### 10. Admin/Shared Lifecycle & Removed-API Migration

Note: the app boots under Vue's migration build (`configureCompat({ MODE: 2 })` in [frontend/src/main.ts](/home/zutfen/code/spire/frontend/src/main.ts)), so Vue 2 lifecycle hooks (`beforeDestroy`/`destroyed`), `this.$set`, and `.native` still *function* but emit deprecation warnings and will break when compat mode is dropped. This step migrates the admin routes and shared shell off those conventions:

- Renamed Vue 2 lifecycle hooks to their Vue 3 equivalents so timer/listener teardown actually runs forward-compatibly:
  - [frontend/src/views/admin/ZoneLogs.vue](/home/zutfen/code/spire/frontend/src/views/admin/ZoneLogs.vue) (`beforeDestroy`/`destroyed` → `beforeUnmount`/`unmounted`, closes the zone-log WebSocket on teardown)
  - [frontend/src/views/admin/ZoneServers.vue](/home/zutfen/code/spire/frontend/src/views/admin/ZoneServers.vue) (`beforeDestroy` → `beforeUnmount`, clears the zone polling interval)
  - [frontend/src/views/admin/tools/DatabaseBackup.vue](/home/zutfen/code/spire/frontend/src/views/admin/tools/DatabaseBackup.vue) (`destroyed` → `unmounted`)
  - [frontend/src/views/admin/layout/AdminHeader.vue](/home/zutfen/code/spire/frontend/src/views/admin/layout/AdminHeader.vue) (`beforeDestroy` → `beforeUnmount`)
  - [frontend/src/views/admin/components/PlayersOnlineComponent.vue](/home/zutfen/code/spire/frontend/src/views/admin/components/PlayersOnlineComponent.vue) (`beforeDestroy` → `beforeUnmount`)
  - [frontend/src/views/admin/components/DashboardNetworkingInfo.vue](/home/zutfen/code/spire/frontend/src/views/admin/components/DashboardNetworkingInfo.vue) (`beforeDestroy` → `beforeUnmount`)
  - [frontend/src/views/admin/components/DashboardCpuInfo.vue](/home/zutfen/code/spire/frontend/src/views/admin/components/DashboardCpuInfo.vue) (`beforeDestroy` → `beforeUnmount`)
  - [frontend/src/views/admin/components/DashboardProcessCounts.vue](/home/zutfen/code/spire/frontend/src/views/admin/components/DashboardProcessCounts.vue) (`beforeDestroy` → `beforeUnmount`)
  - [frontend/src/App.vue](/home/zutfen/code/spire/frontend/src/App.vue) (`destroyed` → `unmounted`)
  - [frontend/src/components/LoaderFakeProgress.vue](/home/zutfen/code/spire/frontend/src/components/LoaderFakeProgress.vue) (`beforeDestroy` → `beforeUnmount`)
  - [frontend/src/components/eq-ui/EQDebug.vue](/home/zutfen/code/spire/frontend/src/components/eq-ui/EQDebug.vue) (`destroyed` → `unmounted`)
  - [frontend/src/components/DbConnectionStatusPill.vue](/home/zutfen/code/spire/frontend/src/components/DbConnectionStatusPill.vue) (`destroyed` → `unmounted`)
- Replaced removed `this.$set(...)` calls with direct reactive assignment (Proxy-based reactivity in Vue 3) in [frontend/src/views/admin/ZoneServers.vue](/home/zutfen/code/spire/frontend/src/views/admin/ZoneServers.vue) (player-toggle map and in-place zone updates)
- Removed a redundant `@click.native` handler from the player event log auto-refresh toggle in [frontend/src/views/admin/player-event-logs/PlayerEventLogs.vue](/home/zutfen/code/spire/frontend/src/views/admin/player-event-logs/PlayerEventLogs.vue) (relying on the existing `v-model`, which is honored by compat mode's Vue 2 `value`/`input` convention)

### 11. Vue 2 Convention Sweep (Lifecycle Hooks + `.native`)

Completed the bounded sweep of the remaining Vue 2 conventions enumerated by `rg -n "beforeDestroy\(|destroyed\(\)|this\.\$set\(|\.native" frontend/src` across editor-heavy and shared routes. Result: **0 occurrences remain**. The app still boots under `configureCompat({ MODE: 2 })`, but these usages no longer emit deprecation warnings and will not break when compat mode is dropped.

- Renamed Vue 2 lifecycle hooks to their Vue 3 equivalents (`beforeDestroy` → `beforeUnmount`, `destroyed` → `unmounted`) so teardown actually runs forward-compatibly:
  - [frontend/src/views/zone/Zone.vue](/home/zutfen/code/spire/frontend/src/views/zone/Zone.vue)
  - [frontend/src/views/tasks/components/TaskTimerCountdown.vue](/home/zutfen/code/spire/frontend/src/views/tasks/components/TaskTimerCountdown.vue)
  - [frontend/src/views/asset-viewers/SpellAnimationViewer.vue](/home/zutfen/code/spire/frontend/src/views/asset-viewers/SpellAnimationViewer.vue)
  - [frontend/src/views/asset-viewers/EmitterViewer.vue](/home/zutfen/code/spire/frontend/src/views/asset-viewers/EmitterViewer.vue)
  - [frontend/src/views/quest-editor/QuestEditor.vue](/home/zutfen/code/spire/frontend/src/views/quest-editor/QuestEditor.vue)
  - [frontend/src/views/quest-api-explorer/QuestApiExplorer.vue](/home/zutfen/code/spire/frontend/src/views/quest-api-explorer/QuestApiExplorer.vue)
  - [frontend/src/components/LoaderCastBarTimer.vue](/home/zutfen/code/spire/frontend/src/components/LoaderCastBarTimer.vue)
  - [frontend/src/views/npcs/NPCs.vue](/home/zutfen/code/spire/frontend/src/views/npcs/NPCs.vue)
  - [frontend/src/components/EqZoneMap.vue](/home/zutfen/code/spire/frontend/src/components/EqZoneMap.vue)
  - [frontend/src/views/items/ItemEditor.vue](/home/zutfen/code/spire/frontend/src/views/items/ItemEditor.vue)
  - [frontend/src/views/sage/Sage.vue](/home/zutfen/code/spire/frontend/src/views/sage/Sage.vue)
  - [frontend/src/components/preview/EQZoneCardPreview.vue](/home/zutfen/code/spire/frontend/src/components/preview/EQZoneCardPreview.vue)
- Removed the deprecated `.native` event modifier from component listeners. All affected listeners were on components (`eq-tabs`, `eq-window-simple`, `b-form-select`, `b-form-input`) whose single-root / `...attrs`-spread behavior forwards a plain `@mouseover` to the root element, preserving the previous native-listener behavior:
  - [frontend/src/views/tasks/TaskEditor.vue](/home/zutfen/code/spire/frontend/src/views/tasks/TaskEditor.vue)
  - [frontend/src/views/spells/SpellEditor.vue](/home/zutfen/code/spire/frontend/src/views/spells/SpellEditor.vue) (3 usages)
  - [frontend/src/views/npcs/NpcEditor.vue](/home/zutfen/code/spire/frontend/src/views/npcs/NpcEditor.vue)
  - [frontend/src/views/items/ItemEditor.vue](/home/zutfen/code/spire/frontend/src/views/items/ItemEditor.vue)

## Verification

Last verified successfully (`2026-06-17`, after Step 11):

- `go build ./...`
- `go build ./internal/http/spa`
- `cd frontend && npm run build`
- `rg -n "beforeDestroy\(|destroyed\(\)|this\.\$set\(|\.native" frontend/src` → 0 matches

## Open Risks / Warnings

- App still boots under Vue's migration build (`configureCompat({ MODE: 2 })`). The targeted sweep in Step 11 cleared all `beforeDestroy`/`destroyed`/`$set`/`.native` usages (verified: 0 matches). Other Vue 2-only instance APIs may still be present (e.g. `this.$children`, `this.$listeners`, `this.$scopedSlots`, `Vue.set`, `.sync`) and would surface when scanning for compat-mode deprecation warnings in the browser; these were intentionally out of scope for this sweep
- Frontend build still emits non-fatal warnings:
  - deprecated Sass legacy JS API
  - deprecated Vue deep selector syntax (`>>>` / `/deep/`)
  - existing duplicate `case 503` warning in `frontend/src/app/spells.ts`
- Connections/user modal fixes are build-verified but not yet browser-smoke-tested end-to-end
- Admin modal/timer fixes are build-verified but not yet browser-smoke-tested end-to-end
- Admin dashboard/zone/lifecycle hook fixes (`$set`, `.native`, `beforeDestroy`/`destroyed`) are build-verified but not yet browser-smoke-tested end-to-end
- Launcher options static-zone add/remove flow is build-verified but not yet browser-smoke-tested end-to-end
- Server update branch switching/build controls and Discord webhook CRUD are build-verified but not yet browser-smoke-tested end-to-end
- Step 11 lifecycle-hook/`.native` sweep is build-verified but not yet browser-smoke-tested; the `@mouseover` fallthrough behavior (eq-tabs/eq-window-simple root `div`, b-form-select/b-form-input via `...attrs` spread) should be confirmed in the browser for the editor/preview hover flows
- Large editor-heavy routes are not yet intentionally re-architected; current success is foundation-first
- Vue 2 specialty libraries are still present as dependency debt even though the app now builds on the new shell
- `docs/project-assessment-2026-06.md` still references Wire historically; that is acceptable unless we want the assessment updated to reflect implementation progress

## Next Step

Recommended next phase:

- Continue Phase 2 frontend migration: finish the remaining admin/shared modal-heavy routes
- Goal: replace real BootstrapVue usage on ordinary CRUD/admin flows with the new local Vue 3 wrapper layer and remove timing-sensitive modal behavior

Suggested first targets:

- `frontend/src/views/admin/configuration/LogSettings.vue`
- `frontend/src/views/admin/configuration/ServerRules.vue`
- `frontend/src/views/admin/configuration/ServerConfig.vue`
- `frontend/src/views/admin/*` routes with modal/tabs/pagination usage

The Vue 2 lifecycle / `$set` / `.native` sweep (Step 11) is complete. As a follow-up to further reduce compat-mode deprecation noise, scan the browser console and sweep the remaining Vue 2-only instance APIs (e.g. `this.$children`, `this.$listeners`, `this.$scopedSlots`, `Vue.set`, `.sync`) with `rg -n "\\\$children|\\\$listeners|\\\$scopedSlots|Vue\\.set|\\.sync=" frontend/src --type-add 'vue:*.vue' --type vue`.

## Session Notes

- Do not revert unrelated user changes
- Keep `frontend/dist` as the backend-consumed artifact shape
- Preserve existing HTTP API contracts, route paths, auth redirects, and query-string behavior
