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

- Date: `2026-06-18`
- Branch: `feature/option-2-migration-foundation`
- Goal: implement Option 2 as an in-place migration to `Vue 3 + Vite + Pinia` while keeping the Go backend and HTTP API stable
- State: frontend foundation is migrated and building; SPA packaging moved from `packr` to `go:embed`; Wire-based bootstrap removed in favor of explicit constructor composition; connections/user modal flows and the current admin update/configuration slices now run on the Vue 3-safe path; admin dashboard/zone/timer lifecycle hooks and removed-API (`$set`/`.native`) usages migrated off Vue 2 conventions; full Vue 2 lifecycle-hook/`$set`/`.native` sweep complete across editor-heavy and shared routes (0 remaining occurrences); `EQTabs`/`EQTab` rewritten off the removed `$children` instance property via `provide`/`inject` registration (tab navigation now works under Vue 3 across all 12 consumers); the three admin configuration routes (`LogSettings`/`ServerRules`/`ServerConfig`) cleaned off BootstrapVue form inputs and Node `util` debt; the remaining `.sync` compat debt has been swept to Vue 3 `v-model:inputData` (0 `.sync` occurrences remain in `frontend/src`); the player-event-log, update-releases, zone-server, and file-log admin routes cleaned of browser-unsafe `util` and simple BootstrapVue wrappers; the server-update branch control, zone-log copy, discord-webhook add/delete, launcher-options numeric inputs, client-assets download link, and dashboard system-info admin seams converted off BootstrapVue and Node `util` (0 admin-side `util` debt remains); a stale 12MB untracked packr box artifact (`internal/http/spa/a_spa-packr.go`) removed so `go build ./...` passes cleanly again; the repo-wide browser-unsafe Node `util` debt cleared in one pass via a new browser-safe `util.format` shim ([frontend/src/app/utility/util-shim.ts](/home/zutfen/code/spire/frontend/src/app/utility/util-shim.ts)) — all 60 `import ... from "util"` lines across `frontend/src` repointed to it, the Vite `util` alias repointed to the shim as a safety net, and the npm `util` polyfill dependency dropped from [frontend/package.json](/home/zutfen/code/spire/frontend/package.json) (0 stray `util` imports remain; `util.format` parity verified against Node for all 100 call-site patterns)

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

### 13. `.sync` Modifier Sweep

Completed the remaining Vue 2 `.sync` compat sweep by converting all 28 `:inputData.sync="..."` call sites to Vue 3 `v-model:inputData="..."`. This was a mechanical drop-in because the receiving components already emit `update:inputData`. Result: **0 `.sync` occurrences remain** in `frontend/src`.

- Converted the calculator/editor call sites:
  - [frontend/src/views/Calculators.vue](/home/zutfen/code/spire/frontend/src/views/Calculators.vue)
  - [frontend/src/views/Components.vue](/home/zutfen/code/spire/frontend/src/views/Components.vue)
  - [frontend/src/views/items/ItemEditor.vue](/home/zutfen/code/spire/frontend/src/views/items/ItemEditor.vue)
  - [frontend/src/views/npcs/NpcEditor.vue](/home/zutfen/code/spire/frontend/src/views/npcs/NpcEditor.vue)
  - [frontend/src/views/spells/SpellEditor.vue](/home/zutfen/code/spire/frontend/src/views/spells/SpellEditor.vue)
- Converted the shared selector/item-picker call sites:
  - [frontend/src/views/items/Items.vue](/home/zutfen/code/spire/frontend/src/views/items/Items.vue)
  - [frontend/src/views/tasks/components/TaskItemSelector.vue](/home/zutfen/code/spire/frontend/src/views/tasks/components/TaskItemSelector.vue)
  - [frontend/src/views/spells/components/SpellItemSelector.vue](/home/zutfen/code/spire/frontend/src/views/spells/components/SpellItemSelector.vue)
  - [frontend/src/components/selectors/ItemSelector.vue](/home/zutfen/code/spire/frontend/src/components/selectors/ItemSelector.vue)

### 14. Admin Log/Release/Zone/File Compat Cleanup

Completed the next handoff-targeted admin slice by removing the remaining browser-unsafe `util` imports from the targeted routes, replacing the timing-sensitive release-notes modal flow with direct local state, and converting the remaining simple BootstrapVue form/button/input-group/pagination usage in these views to native Vue 3-safe markup.

- Cleaned the player event log explorer/settings routes:
  - [frontend/src/views/admin/player-event-logs/PlayerEventLogs.vue](/home/zutfen/code/spire/frontend/src/views/admin/player-event-logs/PlayerEventLogs.vue): replaced the two `b-form-input` filters with native `<input>`, replaced the filter-delete `b-button`, removed the Node `util` import from payload formatting, and replaced `b-pagination` with explicit page buttons driven by a computed `totalPages`
  - [frontend/src/views/admin/player-event-logs/PlayerEventLogSettings.vue](/home/zutfen/code/spire/frontend/src/views/admin/player-event-logs/PlayerEventLogSettings.vue): replaced the search `b-form-input` with a native `<input>` and removed the `util.format(...)` notification formatting
- Cleaned the admin releases route:
  - [frontend/src/views/admin/server-update/UpdateReleases.vue](/home/zutfen/code/spire/frontend/src/views/admin/server-update/UpdateReleases.vue): replaced the `b-modal` + `$bvModal.show(...)` release-notes flow with local `releaseNotesVisible` state and [frontend/src/components/eq-ui/EQModal.vue](/home/zutfen/code/spire/frontend/src/components/eq-ui/EQModal.vue), converted the release action `b-button`s to native `<button>`, removed the stray `console.log("trigger")`, and removed the `util` import from the crash-link opener
- Cleaned the zone/file admin routes:
  - [frontend/src/views/admin/ZoneServers.vue](/home/zutfen/code/spire/frontend/src/views/admin/ZoneServers.vue): replaced the search `b-form-input` with a native `<input>` and removed the `util` import from the player tooltip formatter
          - [frontend/src/views/admin/FileLogs.vue](/home/zutfen/code/spire/frontend/src/views/admin/FileLogs.vue): replaced the search `b-input-group`, file/filter/action `b-button`s, and `b-spinner` with native Bootstrap-markup equivalents

### 15. Admin Update/Zone/Webhook/Launcher Compat + Stale packr Cleanup

Completed the next handoff-targeted admin slice by converting the remaining BootstrapVue (`b-input-group`/`b-button`/`b-form-input`) usage in the listed admin routes to native Bootstrap markup and removing the last admin-side browser-unsafe Node `util` imports. Also removed a stale untracked packr artifact that was silently breaking `go build ./...`.

- Cleaned the server-update branch control into native markup in [frontend/src/views/admin/server-update/ServerUpdate.vue](/home/zutfen/code/spire/frontend/src/views/admin/server-update/ServerUpdate.vue): replaced `b-input-group`/`b-input-group-append`/`b-button` with `<div class="input-group">`/`<div class="input-group-append">`/`<button class="btn btn-white btn-sm">` (matching the FileLogs Step-14 pattern)
- Cleaned the zone-log copy control in [frontend/src/views/admin/ZoneLogs.vue](/home/zutfen/code/spire/frontend/src/views/admin/ZoneLogs.vue): replaced the copy-to-clipboard `b-button` with a native `<button class="btn btn-white btn-sm ml-1">`
- Cleaned the discord-webhook add/delete actions in [frontend/src/views/admin/configuration/DiscordWebhooks.vue](/home/zutfen/code/spire/frontend/src/views/admin/configuration/DiscordWebhooks.vue): replaced both `b-button`s (add-webhook and per-row delete) with native `<button>`s, preserving the `variant`→`btn-*` class mapping (`outline-warning btn-dark` / `primary btn-dark`) and `size`/class/style props
- Cleaned the launcher-options numeric inputs in [frontend/src/views/admin/components/LauncherOptions.vue](/home/zutfen/code/spire/frontend/src/views/admin/components/LauncherOptions.vue): replaced both `b-form-input` (min zone processes / log retention days) with native `<input type="number" class="form-control">`
- Removed the last admin-side browser-unsafe Node `util` imports and converted `util.format(...)` to template literals:
  - [frontend/src/views/admin/tools/ClientAssets.vue](/home/zutfen/code/spire/frontend/src/views/admin/tools/ClientAssets.vue): removed `import util from "util"`, download URL now a template literal (result: 0 admin-side `util` imports remain)
  - [frontend/src/views/admin/components/DashboardSystemInfo.vue](/home/zutfen/code/spire/frontend/src/views/admin/components/DashboardSystemInfo.vue): removed `import * as util from 'util'`, `osDisplay` now a template literal
- Removed the stale untracked packr box artifact `internal/http/spa/a_spa-packr.go` (12MB, generated, no build tag, imported `github.com/gobuffalo/packr` which is no longer in `go.mod`/`go.sum`). It was the sole reason `go build ./...` failed; deletion is safe (the file was never git-tracked; SPA serving uses the tracked `go:embed` files `spa.go`/`packer.go`)

### 16. Repo-Wide Node `util` Debt Cleared via Browser-Safe Shim

Completed the highest-leverage next-step target from the handoff by clearing all remaining browser-unsafe Node `util` debt in a single pass. The handoff's "~15 imports" estimate turned out to be **60 files** (the prior estimate was based on a whitespace-strict regex that missed the alignment-padded `import util            from "util"` lines). `util.format` is the **only** method used across the codebase (**100 call sites**, all `%s`-only printf formatting). Rather than converting ~100 call sites to template literals file-by-file, this step introduces one shim module and repoints every import.

- Added the browser-safe shim [frontend/src/app/utility/util-shim.ts](/home/zutfen/code/spire/frontend/src/app/utility/util-shim.ts): a faithful `format()` that mirrors Node's `util.format` semantics (`%s`, `%d`/`%i`, `%f`, `%j`/`%o`/`%O`, `%c`, `%%`; extra args appended space-separated; missing args leave their specifier intact; `%%` literal). Exports both named `format` and `default { format }` so both `import util from ...` (default) and `import * as util from ...` (namespace) styles keep working
- Repointed all 60 `import ... from "util"` / `from 'util'` sources in `frontend/src` to `from "@/app/utility/util-shim"` (both default and `* as` binding styles preserved). Touched: `app/spells.ts`, `app/db-schema.ts`, `app/api/spire-websocket.ts`, `app/api/spire-query-builder.ts`, `app/api/eqemu-websocket-client.ts`, `router.ts`, `App.vue`, plus selectors/previews/tools/asset-viewers/items/spells/npcs/merchants/bots/tasks/quest-api-explorer/client-files/connections/Home/server-developer views
- Repointed the Vite alias in [frontend/vite.config.ts](/home/zutfen/code/spire/frontend/vite.config.ts): `util: 'util'` (which mapped the bare specifier to the npm polyfill) → `util: fileURLToPath(new URL('./src/app/utility/util-shim.ts', import.meta.url))` so any stray/future bare `util` import resolves browser-safely instead of to the Node polyfill
- Dropped the now-unused npm `util@^0.12.5` polyfill dependency from [frontend/package.json](/home/zutfen/code/spire/frontend/package.json) (debt cleared at the manifest level)
- Verified `util.format` parity against Node's implementation: 18/19 test cases match exactly; the single divergence is the rendering of a raw object/array passed to `%s` (shim uses `JSON.stringify`-style, Node uses `util.inspect`-style). Confirmed by exhaustive arg scan that **no real call site passes a non-primitive/non-pre-stringified value to `%s`** — every `%s` argument is a string, number, boolean, array element, property access, or `JSON.stringify(...)`/`.join(...)` result, so parity holds for 100% of actual usage

## Verification

Last verified successfully (`2026-06-18`, after Step 16):

- `cd frontend && npm run build` (built in ~7s; `node_modules` already present from prior `npm install --legacy-peer-deps --package-lock=false`)
- `go build ./...`
- `go build ./internal/http/spa`
- `rg -n "beforeDestroy\(|destroyed\(\)|this\.\$set\(|\.native" frontend/src --glob '!**/vendors/**'` → 0 matches (app code; only the third-party `assets/vendors/chart.bundle.min.js` matches the `.native` substring, which is out of scope and matches prior Step-11 results)
- `rg -n "\$children" frontend/src --glob '!**/vendors/**'` → 0 matches
- `rg -n "\.sync=" frontend/src --type-add 'vue:*.vue' --type vue` → 0 matches
- `rg -n "\$listeners|\$scopedSlots|Vue\.set" frontend/src --glob '!**/vendors/**'` → 0 matches
- `rg -n "from ['\"]util['\"]" frontend/src` → 0 matches (repo-wide Node `util` imports fully cleared)
- `rg -l "@/app/utility/util-shim" frontend/src | wc -l` → 60 (all repointed imports resolve to the shim)
- orphan check: files using `util.format` but not importing the shim (excluding the shim itself) → 0 (all 100 call sites resolve through the repointed binding)
- `util.format` parity: 18/19 cases match Node exactly; the single divergence (raw object/array to `%s`) does not occur at any real call site (verified by exhaustive arg scan)
- `git diff --check`

## Open Risks / Warnings

- App still boots under Vue's migration build (`configureCompat({ MODE: 2 })`). The targeted sweeps cleared all `beforeDestroy`/`destroyed`/`$set`/`.native` usages, the single real `$children` usage, and the repo-wide `.sync` debt in `frontend/src` (verified: 0 matches each). Other Vue 2-only instance APIs may still be present outside the scanned set and would surface when scanning for compat-mode deprecation warnings in the browser; those remain intentionally out of scope until a dedicated compat-warning pass
- **Repo-wide Node `util` debt is fully cleared**: Step 16 introduced [frontend/src/app/utility/util-shim.ts](/home/zutfen/code/spire/frontend/src/app/utility/util-shim.ts) and repointed all 60 `import ... from "util"` lines in `frontend/src` to it (0 stray `util` imports remain). The npm `util@^0.12.5` polyfill was dropped from `package.json`, and the Vite `util` alias now points at the shim as a safety net. The shim's `format()` is parity-verified against Node for all 100 `util.format` call-site patterns; the only divergence (raw object/array passed to `%s`) does not occur at any real call site
- Fresh frontend dependency installs still need the legacy peer resolver while Vue 2 bridge packages remain in tree: plain `npm ci` currently fails on the expected `vue-class-component@7.2.6` peer conflict against Vue 3, while `npm install --legacy-peer-deps --package-lock=false` restored a working local build. Note: the npm `util` polyfill was dropped from `package.json` in Step 16 but `frontend/package-lock.json` still lists it (the lockfile is non-authoritative per the `--package-lock=false` install procedure and was intentionally not hand-edited to avoid corrupting its dependency tree/checksums); it will clear on the next clean lockfile regeneration
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
- Step 12 `EQTabs`/`EQTab` `provide`/`inject` rewrite is build-verified but not yet browser-smoke-tested; tab selection (incl. nested `eq-tabs` in `ServerConfig` and the v-for loginserver tabs), hover-to-select, and the `selected` query-string restore should be confirmed in the browser. The three admin configuration routes are likewise build-verified only
- Step 14 admin log/release/zone/file cleanup is build-verified only; browser smoke should confirm the release-notes modal open/close flow, player-event-log page switching/filter deletion, zone-server search, and file-log search/watch controls
- Step 15 admin update/zone/webhook/launcher compat + `util` removal is build-verified only; browser smoke should confirm the server-update branch select+Set append control, zone-log copy-to-clipboard button, discord-webhook add/delete buttons, and launcher-options numeric inputs. `ServerProcessButtonComponent.vue` still contains `b-button`/`b-spinner` usage (preflight/start/stop/restart controls) but was intentionally left out of this slice as it was already touched in Step 7 and is not a blocker
- Step 16 repo-wide `util` shim migration is build-verified only; browser smoke should confirm that string-formatted outputs render identically (notably router path interpolation in `router.ts`, websocket URIs in `app/api/spire-websocket.ts`/`eqemu-websocket-client.ts`, query-builder `where` clauses in `app/api/spire-query-builder.ts`, asset/key formatting in selectors/asset-viewers, and the markdown YouTube embeds in `Home.vue`). The shim's `format()` is parity-tested against Node for all `%s` call-site patterns, so behavioral regressions are not expected, but an end-to-end smoke of a few editor/preview flows is worthwhile
- `go build ./...` now passes cleanly after removing the stale untracked packr artifact (Step 15); prior "GO BUILD OK" lines in earlier verification runs were masked by piping through `tail`, so the packr breakage had been silently present since the Step-4 embed migration. Going forward, run `go build ./...` without a pipe (or check `$?`) to catch this class of regression
- Large editor-heavy routes are not yet intentionally re-architected; current success is foundation-first
- Vue 2 specialty libraries are still present as dependency debt even though the app now builds on the new shell
- `docs/project-assessment-2026-06.md` still references Wire historically; that is acceptable unless we want the assessment updated to reflect implementation progress

## Next Step

Recommended next phase:

- Continue Phase 2 frontend migration: finish the remaining shared/admin compat cleanup around lingering BootstrapVue-style inputs/buttons (the browser-unsafe Node shim debt is now fully cleared via the Step-16 `util.format` shim)
- Goal: keep shrinking the compat surface before any browser warning pass or compat-mode removal attempt

The Step-16 repo-wide `util` shim migration is complete (60 imports repointed, npm `util` dep dropped, 0 stray `util` imports remain). The Step-15 suggested targets (`ServerUpdate.vue`, `ZoneLogs.vue`, `DiscordWebhooks.vue`, `LauncherOptions.vue`, `ClientAssets.vue`, `DashboardSystemInfo.vue`) are also complete. Suggested next targets:

- `frontend/src/views/admin/components/ServerProcessButtonComponent.vue` (still uses `b-button`/`b-spinner` in the preflight/start/stop/restart controls — the last significant BootstrapVue cluster in the admin shell; already touched in Step 7 so re-check before editing)
- `frontend/src/views/client-files/ClientFiles.vue` and `frontend/src/components/DbColumnFilter.vue` / `DbConnectionStatusPill.vue` (these now import the `util` shim rather than Node `util`, but may still carry BootstrapVue markup clusters worth converting to native inputs/buttons)
- A broader BootstrapVue (`b-button`/`b-form-input`/`b-input-group`/`b-modal`/`b-spinner`/`b-pagination`) sweep across the shared/editor routes — the admin shell is largely cleared, but the editor-heavy and selector routes still contain BootstrapVue usage that would emit compat-mode warnings

The Vue 2 lifecycle / `$set` / `.native` sweep (Step 11), the `$children` removal (Step 12), the `.sync` sweep (Step 13), the Step 14 admin compat cleanup, the Step 15 admin compat + packr cleanup, and the Step 16 repo-wide `util` shim migration above are complete. The remaining Vue 2-only instance API scan stays clean with `rg -n "\\\$listeners|\\\$scopedSlots|Vue\\.set" frontend/src` (0 matches on 2026-06-18).

## Session Notes

- Do not revert unrelated user changes
- Keep `frontend/dist` as the backend-consumed artifact shape
- Preserve existing HTTP API contracts, route paths, auth redirects, and query-string behavior
