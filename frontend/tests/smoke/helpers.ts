import type { Page, Route } from '@playwright/test'

type SmokeHarnessOptions = {
  tabHoverEnabled?: boolean
}

const jsonHeaders = {
  'access-control-allow-origin': '*',
  'content-type': 'application/json',
}

const serverConfigFixture = {
  server: {
    world: {
      longname: 'Spire Test Server',
      shortname: 'spiretest',
      address: '203.0.113.10',
      localaddress: '192.168.1.10',
      key: 'ABCDEFGHIJKLMNOPQRSTUVXYZ1234567890ABCD',
      telnet: {
        enabled: true,
        ip: '0.0.0.0',
        port: '9000',
      },
      tcp: {
        ip: '0.0.0.0',
        port: '9001',
      },
      loginserver1: {
        account: 'server-admin',
        password: 'secret',
        legacy: '0',
        host: 'login.eqemu.dev',
        port: '5998',
      },
    },
    zones: {
      defaultstatus: '0',
      ports: {
        low: 7000,
        high: 7500,
      },
    },
    ucs: {
      host: '127.0.0.1',
      port: '7778',
    },
    database: {
      db: 'peq',
      host: 'db.internal',
      port: '3306',
      username: 'spire',
      password: 'password',
    },
    content_database: {
      db: 'peq_content',
      host: 'content.internal',
      port: '3306',
      username: 'content',
      password: 'password',
    },
    qsdatabase: {
      db: 'peq_logs',
      host: 'qs.internal',
      port: '3306',
      username: 'qs',
      password: 'password',
    },
  },
}

const releaseAnalyticsFixture = {
  data: [
    {
      tag_name: 'v1.2.4',
      name: 'v1.2.4',
      body: '* Added Vue 3 smoke coverage\n* Fixed release modal flow',
      published_at: '2026-06-20T12:00:00Z',
      assets: [
        {
          name: 'eqemu-windows-x64.zip',
          download_count: 42,
          browser_download_url: 'https://example.com/eqemu-windows-x64.zip',
        },
        {
          name: 'eqemu-linux-x64.tar.gz',
          download_count: 24,
          browser_download_url: 'https://example.com/eqemu-linux-x64.tar.gz',
        },
      ],
    },
  ],
}

const crashCountsFixture = {
  crash_report_counts: [
    { server_version: '1.2.4', crash_count: 5 },
    { server_version: '1.2.4-dev', crash_count: 2 },
  ],
  unique_crash_counts: [
    { server_version: '1.2.4', signature: 'zone-crash' },
    { server_version: '1.2.4', signature: 'login-crash' },
  ],
}

const playerEventLogSettingsFixture = [
  {
    id: 101,
    event_name: 'Player Login',
    event_enabled: 1,
    etl_enabled: 1,
    retention_days: 7,
    discord_webhook_id: 0,
    log_to_discord: 0,
    log_category_id: 101,
  },
  {
    id: 102,
    event_name: 'Player Logout',
    event_enabled: 1,
    etl_enabled: 0,
    retention_days: 14,
    discord_webhook_id: 1,
    log_to_discord: 1,
    log_category_id: 102,
  },
]

const discordWebhooksFixture = [
  {
    id: 1,
    webhook_name: 'Operations',
  },
]

const zoneServerListFixture = [
  {
    id: 1,
    zone_name: 'qeynos',
    zone_long_name: 'Qeynos Hills',
    number_players: 1,
    zone_os_pid: 1337,
    is_static_zone: true,
    zone_id: 2,
    instance_id: 0,
    zone_server_address: '127.0.0.1',
    client_port: 7001,
    compile_date: '2026-06-20',
    compile_time: '12:00',
    compile_version: '1.2.4',
    cpu: '12',
    memory: 209715200,
    elapsed: 7200,
    clients: [
      {
        id: 500,
        name: 'Alyra',
        class: 1,
        race: 1,
        guild: 'Knights of Testing',
      },
    ],
  },
  {
    id: 2,
    zone_name: 'freeport',
    zone_long_name: 'West Freeport',
    number_players: 0,
    zone_os_pid: 1441,
    is_static_zone: false,
    zone_id: 9,
    instance_id: 0,
    zone_server_address: '127.0.0.2',
    client_port: 7002,
    compile_date: '2026-06-20',
    compile_time: '12:00',
    compile_version: '1.2.4',
    cpu: '8',
    memory: 157286400,
    elapsed: 3600,
    clients: [],
  },
]

const zonesFixture = [
  {
    id: 1,
    short_name: 'qeynos',
    long_name: 'Qeynos Hills',
    zoneidnumber: 2,
    expansion: 0,
  },
  {
    id: 2,
    short_name: 'freeport',
    long_name: 'West Freeport',
    zoneidnumber: 9,
    expansion: 0,
  },
]

const fileLogsFixture = [
  {
    path: 'zone-eqemu.log',
    modified_time: 1750982400,
    size: 4096,
  },
  {
    path: 'world-eqemu.log',
    modified_time: 1750982401,
    size: 8192,
  },
  {
    path: 'crashes-2026-06-20.log',
    modified_time: 1750982402,
    size: 1024,
  },
]

const fileLogContentsFixture = {
  contents: '[WorldServer] [Info] Zone booted\n[WorldServer] [Warn] High load\n',
  cursor: 128,
}

const userConnectionsFixture = {
  data: [
    {
      id: 1,
      active: 1,
      server_database_connection_id: 10,
      database_connection: {
        id: 1,
        name: 'Test Production DB',
        db_host: 'db.internal',
        db_port: '3306',
        db_name: 'peq',
        db_username: 'spire',
        content_db_username: '',
        content_db_host: '',
        content_db_port: '',
        content_db_name: '',
        created_by: 1,
        user_server_database_connections: [
          {
            user: {
              id: 2,
              user_name: 'DevAlice',
              avatar: 'data:image/gif;base64,R0lGODlhAQABAAAAACw=',
              deleted_at: null,
            },
          },
        ],
      },
    },
  ],
}

const playerEventLogsFixture = [
  {
    id: 9001,
    event_type_id: 29,
    event_type_name: 'SAY',
    character_id: 500,
    zone_id: 2,
    event_data: JSON.stringify({ message: 'Hail, traveler', target: 'a_guard' }),
    created_at: '2026-07-01T12:00:00Z',
  },
  {
    id: 9000,
    event_type_id: 10,
    event_type_name: 'WENT_ONLINE',
    character_id: 500,
    zone_id: 0,
    event_data: '{}',
    created_at: '2026-07-01T11:30:00Z',
  },
]

const playerEventLogsCountFixture = { count: 2 }

const characterDataBulkFixture = [
  {
    id: 500,
    name: 'Alyra',
    class: 1,
    race: 1,
  },
]

async function fulfillJson(route: Route, payload: unknown, status = 200) {
  await route.fulfill({
    status,
    headers: jsonHeaders,
    body: JSON.stringify(payload),
  })
}

export async function bootstrapSmokeHarness(page: Page, options: SmokeHarnessOptions = {}) {
  const pageErrors: Error[] = []
  const consoleErrors: string[] = []
  page.on('pageerror', (error) => {
    pageErrors.push(error)
  })
  page.on('console', (message) => {
    if (message.type() === 'error') {
      consoleErrors.push(message.text())
    }
  })

  await page.addInitScript(({ tabHoverEnabled }) => {
    class MockWebSocket {
      static CONNECTING = 0
      static OPEN = 1
      static CLOSING = 2
      static CLOSED = 3

      url: string
      readyState = MockWebSocket.OPEN
      onopen: ((event: Event) => void) | null = null
      onclose: ((event: CloseEvent) => void) | null = null
      onerror: ((event: Event) => void) | null = null
      onmessage: ((event: MessageEvent) => void) | null = null
      private listeners: Record<string, Array<(event: Event) => void>> = {}

      constructor(url: string) {
        this.url = url
        setTimeout(() => {
          const event = new Event('open')
          this.onopen?.(event)
          this.dispatch('open', event)
        }, 0)
      }

      addEventListener(type: string, listener: (event: Event) => void) {
        this.listeners[type] = this.listeners[type] || []
        this.listeners[type].push(listener)
      }

      removeEventListener(type: string, listener: (event: Event) => void) {
        this.listeners[type] = (this.listeners[type] || []).filter((entry) => entry !== listener)
      }

      send() {}

      close() {
        this.readyState = MockWebSocket.CLOSED
        const event = new CloseEvent('close')
        this.onclose?.(event)
        this.dispatch('close', event)
      }

      private dispatch(type: string, event: Event) {
        for (const listener of this.listeners[type] || []) {
          listener(event)
        }
      }
    }

    Object.defineProperty(window, 'WebSocket', {
      configurable: true,
      writable: true,
      value: MockWebSocket,
    })

    window.alert = () => {}
    window.confirm = () => true
    window.open = ((url?: string | URL | undefined) => {
      ;(window as typeof window & { __lastOpenedUrl?: string }).__lastOpenedUrl = String(url ?? '')
      return null
    }) as typeof window.open

    if (tabHoverEnabled) {
      localStorage.setItem('tab-hover', 'true')
    }
  }, { tabHoverEnabled: options.tabHoverEnabled ?? false })

  await page.route('**/*', async (route) => {
    const url = new URL(route.request().url())
    const path = url.pathname
    const method = route.request().method()

    if (path === '/api/v1/app/env') {
      return fulfillJson(route, {
        data: {
          os: 'linux',
          env: 'local',
          version: '1.2.4',
          features: {
            github_auth_enabled: false,
          },
          settings: [
            { setting: 'AUTH_ENABLED', value: 'false' },
          ],
          is_spire_initialized: true,
        },
      })
    }

    if (path === '/api/v1/me') {
      return fulfillJson(route, {
        id: 1,
        user_name: 'Codex Tester',
        avatar: 'data:image/gif;base64,R0lGODlhAQABAAAAACw=',
        is_admin: true,
      })
    }

    if (path === '/api/v1/connections') {
      return fulfillJson(route, userConnectionsFixture)
    }

    if (path === '/api/v1/connection-default') {
      return fulfillJson(route, { data: null })
    }

    if (path === '/api/v1/eqemuserver/logs') {
      return fulfillJson(route, fileLogsFixture)
    }

    if (path.startsWith('/api/v1/eqemuserver/log/')) {
      return fulfillJson(route, fileLogContentsFixture)
    }

    if (path.startsWith('/api/v1/eqemuserver/log-search/')) {
      return fulfillJson(route, [])
    }

    if (path.startsWith('/api/v1/connection-check/')) {
      return fulfillJson(route, { data: { message: 'online' } })
    }

    if (path === '/api/v1/admin/serverconfig') {
      return fulfillJson(route, method === 'GET' ? serverConfigFixture : { message: 'saved' })
    }

    if (path === '/api/v1/player_event_log_settings') {
      return fulfillJson(route, playerEventLogSettingsFixture)
    }

    if (path.startsWith('/api/v1/player_event_log_setting/')) {
      return fulfillJson(route, { message: 'saved' })
    }

    if (path === '/api/v1/player_event_logs/count') {
      return fulfillJson(route, playerEventLogsCountFixture)
    }

    if (path === '/api/v1/player_event_logs') {
      return fulfillJson(route, playerEventLogsFixture)
    }

    if (path === '/api/v1/character_data/bulk') {
      return fulfillJson(route, characterDataBulkFixture)
    }

    if (path === '/api/v1/aa_ranks' || path === '/api/v1/aa_abilities' || path === '/api/v1/db_strs') {
      return fulfillJson(route, [])
    }

    if (path === '/api/v1/discord_webhooks') {
      return fulfillJson(route, discordWebhooksFixture)
    }

    if (path === '/api/v1/eqemuserver/player-event-logs/etl-settings') {
      return fulfillJson(route, {
        etl_settings: [{ event_id: 101 }],
      })
    }

    if (path === '/api/v1/eqemuserver/reload/logs') {
      return fulfillJson(route, { message: 'logs reloaded' })
    }

    if (path === '/api/v1/guilds') {
      return fulfillJson(route, [])
    }

    if (path === '/api/v1/eqemuserver/zoneserver-list') {
      return fulfillJson(route, zoneServerListFixture)
    }

    if (path === '/api/v1/zones') {
      return fulfillJson(route, zonesFixture)
    }

    if (path === '/api/v1/eqemuserver/dashboard-stats') {
      return fulfillJson(route, {})
    }

    if (path === '/api/v1/eqemuserver/get-lock-status') {
      return fulfillJson(route, { locked: false })
    }

    if (path === '/api/v1/eqemuserver/server-stats') {
      return fulfillJson(route, {
        boot_time: 'Worldserver Uptime | 1 Day, 2 Hours',
        launcher_connected: true,
      })
    }

    if (path === '/api/v1/eqemuserver/system-all') {
      return fulfillJson(route, [])
    }

    if (path.endsWith('/api/v1/analytics/releases')) {
      return fulfillJson(route, releaseAnalyticsFixture)
    }

    if (path.endsWith('/api/v1/analytics/server-crash-report/counts')) {
      return fulfillJson(route, crashCountsFixture)
    }

    if (path === '/repos/EQEmu/spire/releases/latest') {
      return fulfillJson(route, {
        tag_name: 'v1.2.4',
        body: '* Local smoke test fixture',
      })
    }

    if (path.startsWith('/api/v1/')) {
      return fulfillJson(route, {})
    }

    await route.continue()
  })

  return { consoleErrors, pageErrors }
}
