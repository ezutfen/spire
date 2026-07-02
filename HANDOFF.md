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

- Date: `2026-07-02`
- Branch: `further-modernization`
- Goal: implement Option 2 as an in-place migration to `Vue 3 + Vite + Pinia` while keeping the Go backend and HTTP API stable
- State: frontend foundation is migrated and now builds on plain Vue 3 (no `@vue/compat`, no `configureCompat({ MODE: 2 })`); SPA packaging moved from `packr` to `go:embed`; Wire-based bootstrap removed in favor of explicit constructor composition; connections/user modal flows and the current admin update/configuration slices now run on the Vue 3-safe path; admin dashboard/zone/timer lifecycle hooks and removed-API (`$set`/`.native`) usages migrated off Vue 2 conventions; full Vue 2 lifecycle-hook/`$set`/`.native`/`.sync` sweep complete across editor-heavy and shared routes (0 remaining occurrences for those searches); `EQTabs`/`EQTab` rewritten off the removed `$children` instance property via `provide`/`inject` registration; first-party frontend `util.format` usage removed behind a browser-safe helper; `EqZoneMap` moved off `vue2-leaflet` to `@vue-leaflet/vue-leaflet`; targeted Playwright smoke coverage now validates the migrated `ServerConfig`, releases modal, player event log settings save flow, `ZoneServers` query restore, `FileLogs` listing/filter/stream, and the connections manage-developer modal flow; vitest unit-test run now excludes the Playwright smoke directory so `npm test` stays green

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

### 12. EQTabs `$children` Removal + Admin Configuration Routes

The single remaining real `this.$children` usage (`frontend/src/components/eq-ui/EQTabs.vue`) is a **removed** Vue 3 instance property — under the migration build it returns an empty array, so `eq-tabs` tab navigation (and the `selected`-based activation) was broken for all 12 consumers (incl. `ServerConfig.vue` and the major editors). Replaced it with the idiomatic Vue 3 parent/child discovery pattern, then cleaned the three admin configuration routes flagged as the next-step targets.

- Rewrote `EQTabs`/`EQTab` off `$children` using `provide`/`inject` registration:
  - [frontend/src/components/eq-ui/EQTabs.vue](/home/zutfen/code/spire/frontend/src/components/eq-ui/EQTabs.vue) now exposes an `eqTabsApi` (register/unregister) via `provide()` and manages the reactive `tabs` array directly; removed the `created() { this.tabs = this.$children }` line
  - [frontend/src/components/eq-ui/EQTab.vue](/home/zutfen/code/spire/frontend/src/components/eq-ui/EQTab.vue) injects `eqTabsApi`, registers itself in `mounted()`, and unregisters in `beforeUnmount()`
  - Nesting (e.g. outer/inner `eq-tabs` in `ServerConfig`) resolves correctly because each `EQTabs` provides its own API and `inject` picks the nearest ancestor provider
- Migrated the admin configuration routes off BootstrapVue form inputs and Node debt:
  - [frontend/src/views/admin/configuration/LogSettings.vue](/home/zutfen/code/spire/frontend/src/views/admin/configuration/LogSettings.vue): replaced `b-form-input` with a native `<input>`; removed the browser-unsafe `import util from "util"` and converted `util.format(...)` to a template literal
  - [frontend/src/views/admin/configuration/ServerRules.vue](/home/zutfen/code/spire/frontend/src/views/admin/configuration/ServerRules.vue): replaced `b-form-input` with a native `<input>`
  - [frontend/src/views/admin/configuration/ServerConfig.vue](/home/zutfen/code/spire/frontend/src/views/admin/configuration/ServerConfig.vue): removed stray `console.log("trigger")` debug noise (route was already on native inputs and `eq-tabs`/`eq-tab`, now fixed via the `EQTabs` change above)

### 13. Pure Vue 3 Cutover

- Replaced the final 28 `:inputData.sync="..."` bindings with `v-model:inputData="..."` across the remaining editor / selector parents:
  - [frontend/src/views/Calculators.vue](/home/zutfen/code/spire/frontend/src/views/Calculators.vue)
  - [frontend/src/views/Components.vue](/home/zutfen/code/spire/frontend/src/views/Components.vue)
  - [frontend/src/views/items/Items.vue](/home/zutfen/code/spire/frontend/src/views/items/Items.vue)
  - [frontend/src/views/items/ItemEditor.vue](/home/zutfen/code/spire/frontend/src/views/items/ItemEditor.vue)
  - [frontend/src/views/npcs/NpcEditor.vue](/home/zutfen/code/spire/frontend/src/views/npcs/NpcEditor.vue)
  - [frontend/src/views/spells/SpellEditor.vue](/home/zutfen/code/spire/frontend/src/views/spells/SpellEditor.vue)
  - [frontend/src/views/spells/components/SpellItemSelector.vue](/home/zutfen/code/spire/frontend/src/views/spells/components/SpellItemSelector.vue)
  - [frontend/src/views/tasks/components/TaskItemSelector.vue](/home/zutfen/code/spire/frontend/src/views/tasks/components/TaskItemSelector.vue)
  - [frontend/src/components/selectors/ItemSelector.vue](/home/zutfen/code/spire/frontend/src/components/selectors/ItemSelector.vue)
- Added a browser-safe formatter helper and tests:
  - [frontend/src/app/utility/string-format.ts](/home/zutfen/code/spire/frontend/src/app/utility/string-format.ts)
  - [frontend/src/app/utility/string-format.test.ts](/home/zutfen/code/spire/frontend/src/app/utility/string-format.test.ts)
- Replaced first-party `util.format(...)` / `import ... from "util"` usage across the frontend with `stringFormat(...)`, including the targeted admin files (`PlayerEventLogs`, `PlayerEventLogSettings`, `UpdateReleases`, `ZoneServers`) plus the shared routes/helpers that still depended on the polyfill
- Removed the Vue compat cutover scaffolding:
  - deleted `configureCompat({ MODE: 2 })` from [frontend/src/main.ts](/home/zutfen/code/spire/frontend/src/main.ts)
  - removed the `vue: '@vue/compat'` and `util: 'util'` aliases from [frontend/vite.config.ts](/home/zutfen/code/spire/frontend/vite.config.ts)
  - removed `@vue/compat` and `util` from [frontend/package.json](/home/zutfen/code/spire/frontend/package.json)
- Fixed the post-compat map build blocker by swapping [frontend/src/components/EqZoneMap.vue](/home/zutfen/code/spire/frontend/src/components/EqZoneMap.vue) from `vue2-leaflet` to `@vue-leaflet/vue-leaflet` and updating the frontend dependency set accordingly
- Removed clearly stray debug noise in the touched admin / mirrored release views:
  - [frontend/src/views/admin/server-update/UpdateReleases.vue](/home/zutfen/code/spire/frontend/src/views/admin/server-update/UpdateReleases.vue)
  - [frontend/src/views/server-developer/Releases.vue](/home/zutfen/code/spire/frontend/src/views/server-developer/Releases.vue)
  - [frontend/src/views/admin/player-event-logs/PlayerEventLogs.vue](/home/zutfen/code/spire/frontend/src/views/admin/player-event-logs/PlayerEventLogs.vue)
  - [frontend/src/views/admin/FileLogs.vue](/home/zutfen/code/spire/frontend/src/views/admin/FileLogs.vue)

### 14. Vue 3 Runtime Smoke Coverage

- Added Playwright smoke coverage and harness support in:
  - [frontend/playwright.config.ts](/home/zutfen/code/spire/frontend/playwright.config.ts)
  - [frontend/tests/smoke/helpers.ts](/home/zutfen/code/spire/frontend/tests/smoke/helpers.ts)
  - [frontend/tests/smoke/vue3-smoke.spec.ts](/home/zutfen/code/spire/frontend/tests/smoke/vue3-smoke.spec.ts)
  - [frontend/package.json](/home/zutfen/code/spire/frontend/package.json)
- Fixed browser-runtime blockers surfaced by the new smoke suite:
  - replaced browser-invalid `require(...)` usage in [frontend/src/views/admin/components/ServerProcessButtonComponent.vue](/home/zutfen/code/spire/frontend/src/views/admin/components/ServerProcessButtonComponent.vue)
  - added [frontend/src/app/assets/class-race-icon-url.ts](/home/zutfen/code/spire/frontend/src/app/assets/class-race-icon-url.ts) and switched icon consumers off dynamic `require(...)`:
    [frontend/src/views/admin/ZoneServers.vue](/home/zutfen/code/spire/frontend/src/views/admin/ZoneServers.vue),
    [frontend/src/views/admin/player-event-logs/PlayerEventLogs.vue](/home/zutfen/code/spire/frontend/src/views/admin/player-event-logs/PlayerEventLogs.vue),
    [frontend/src/views/admin/components/PlayersOnlineComponent.vue](/home/zutfen/code/spire/frontend/src/views/admin/components/PlayersOnlineComponent.vue)
  - replaced `require("uuid/v4")` in [frontend/src/components/layout/NavSectionComponent.vue](/home/zutfen/code/spire/frontend/src/components/layout/NavSectionComponent.vue) with `crypto.randomUUID()` plus a fallback
  - fixed async runtime-template registration in [frontend/src/views/server-developer/Releases.vue](/home/zutfen/code/spire/frontend/src/views/server-developer/Releases.vue) and [frontend/src/views/admin/server-update/UpdateReleases.vue](/home/zutfen/code/spire/frontend/src/views/admin/server-update/UpdateReleases.vue) using `defineAsyncComponent(...)`
  - normalized nested config defaults and delayed initial tab rendering in [frontend/src/views/admin/configuration/ServerConfig.vue](/home/zutfen/code/spire/frontend/src/views/admin/configuration/ServerConfig.vue) so hidden tab content no longer crashes on missing subtrees under Vue 3
- Verified smoke flows:
  - `ServerConfig` query restore + hover tab selection
  - releases analytics + release notes modal
  - player event log settings save + reload notification
  - `ZoneServers` query restore + filtered player rendering

### 15. Expanded Runtime Smoke Coverage + Test Harness Cleanup

- Added two more Playwright smoke flows covering high-risk Vue 3 routes that were previously unverified:
  - `FileLogs` listing/filter/stream: renders the file listing, filters by log type (`Zone`), and opens a watched log file that streams contents via the polling timer in [frontend/src/views/admin/FileLogs.vue](/home/zutfen/code/spire/frontend/src/views/admin/FileLogs.vue)
  - `UserConnections` + manage-developer modal: renders a user-owned database connection and opens the `manage-developer-modal` (the migrated `nextTick()` + `$bvModal.show(...)` flow) in [frontend/src/views/connections/UserConnections.vue](/home/zutfen/code/spire/frontend/src/views/connections/UserConnections.vue)
- Extended the smoke harness with file-log listing/contents, log-search, and a richer `/api/v1/connections` fixture (with a developer + owner relationship) in [frontend/tests/smoke/helpers.ts](/home/zutfen/code/spire/frontend/tests/smoke/helpers.ts)
- Fixed the `npm test` (vitest) regression introduced when the Playwright smoke spec was added: vitest was picking up `tests/smoke/*.spec.ts` and failing. Added a `test.exclude` for `tests/smoke/**` in [frontend/vite.config.ts](/home/zutfen/code/spire/frontend/vite.config.ts) so the unit-test run stays green
- Recorded the required `--legacy-peer-deps` install behavior in [frontend/.npmrc](/home/zutfen/code/spire/frontend/.npmrc) so fresh `npm install` runs resolve the legacy Vue 2 dependency peer conflicts (`vue-class-component` / `vue-property-decorator`) without manual flags
- Verified smoke flows now also cover `FileLogs` and the connections manage-developer modal

## Verification

Last verified successfully:

- `2026-07-02` (after Step 15):
  - `cd frontend && npm test` (4 unit tests pass; smoke dir excluded from vitest)
  - `cd frontend && npm run test:smoke` (6 Playwright smoke tests pass)
  - `cd frontend && npm run build`
- `2026-06-26` (after Step 14):
  - `cd frontend && npm run test:smoke`
- `2026-06-24` (after Step 13):
  - `go build ./...`
  - `go build ./internal/http/spa`
  - `cd frontend && npm test`
  - `cd frontend && npm run build`
  - `rg -n "\.sync=" frontend/src --type-add 'vue:*.vue' --type vue` → 0 matches
  - `rg -n "util\.format\(|from [\"']util[\"']" frontend/src` → 0 matches
  - `rg -n "@vue/compat|configureCompat|MODE: 2" frontend -g '!frontend/package-lock.json'` → 0 matches
  - `rg -n "beforeDestroy\(|destroyed\(\)|this\.\$set\(|\.native" frontend/src` (`.vue` files) → 0 matches
  - `rg -n "\$children" frontend/src` (excluding `assets/vendors/`) → 0 matches

## Open Risks / Warnings

- Frontend build still emits non-fatal warnings:
  - deprecated Sass legacy JS API
  - deprecated Vue deep selector syntax (`>>>` / `/deep/`)
  - existing duplicate `case 503` warning in `frontend/src/app/spells.ts`
- `/img/eq-wallpaper-1.b2319219.jpg` still reports a runtime-resolution warning during `vite build`
- The pure Vue 3 cutover is only partially browser-smoke-tested so far:
  - covered by Playwright smoke: `ServerConfig`, releases modal, player event log settings save/reload, `ZoneServers` query restore, `FileLogs` listing/filter/stream, connections manage-developer modal
  - still not covered: `EqZoneMap` interactions, the player event log explorer (`PlayerEventLogs.vue` event grid + runtime-template raw view), launcher/update flows that rely on live process state, and editor hover flows previously affected by `.native`
- Vue 2-era dependency debt still exists outside the build blocker fixed here (`vue2-ace-editor`, `vue2-dropzone`, `vuex`, `vue-property-decorator`, `vue-class-component`); the app now builds without compat, but those packages should be revalidated or retired in future cleanup
- `docs/project-assessment-2026-06.md` still references Wire historically; that is acceptable unless we want the assessment updated to reflect implementation progress

## Next Step

Recommended next phase:

- Continue expanding runtime verification across the remaining high-risk Vue 3 routes, using Playwright smoke where practical and a short manual pass for the harder interactive flows
- Goal: confirm the remaining admin/editor runtime behavior now that compat mode is gone

Suggested next targets:

- `frontend/src/views/admin/player-event-logs/PlayerEventLogs.vue` event grid + runtime-template raw-event view (needs richer fixtures: player event logs, characters, AA preload)
- [frontend/src/components/EqZoneMap.vue](/home/zutfen/code/spire/frontend/src/components/EqZoneMap.vue) leaflet interactions
- launcher/update flows that still rely on live process state or richer backend responses
- editor hover flows previously affected by `.native`

## Session Notes

- Do not revert unrelated user changes
- Keep `frontend/dist` as the backend-consumed artifact shape
- Preserve existing HTTP API contracts, route paths, auth redirects, and query-string behavior
