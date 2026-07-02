import { expect, test } from '@playwright/test'

import { bootstrapSmokeHarness } from './helpers'

test('server config restores tab state and supports hover tab switching', async ({ page }) => {
  const { consoleErrors, pageErrors } = await bootstrapSmokeHarness(page, { tabHoverEnabled: true })

  await page.goto('/admin/configuration/server?s=Zone+Server')

  await expect(page.getByText('Default Player Account Status')).toBeVisible()
  await expect(page.getByText('Zone Port Range Start')).toBeVisible()

  await page.locator('.eq-tab-box-fancy').first().locator('li').filter({ hasText: /^World Server$/ }).hover()

  await expect(page.getByText('Server Long Name')).toBeVisible()
  await expect(page.getByText('Server Short Name')).toBeVisible()
  expect(consoleErrors).toEqual([])
  expect(pageErrors).toEqual([])
})

test('release analytics renders and opens release notes modal', async ({ page }) => {
  const { consoleErrors, pageErrors } = await bootstrapSmokeHarness(page)

  await page.goto('/dev/releases')

  await expect(page.getByText('Release Version Analytics')).toBeVisible()
  await expect(page.getByText('v1.2.4')).toBeVisible()

  await page.getByTitle('View Release Notes').first().click()

  await expect(page.getByText('Release Notes')).toBeVisible()
  await expect(page.getByText('Added Vue 3 smoke coverage')).toBeVisible()
  expect(consoleErrors).toEqual([])
  expect(pageErrors).toEqual([])
})

test('player event log settings save path shows reload notification', async ({ page }) => {
  const { consoleErrors, pageErrors } = await bootstrapSmokeHarness(page)

  await page.goto('/admin/player-event-logs/settings?search=login')

  await expect(page.getByPlaceholder('Search log settings...')).toHaveValue('login')
  await expect(page.getByText('(101) Player Login')).toBeVisible()

  const row = page.locator('tr', { hasText: 'Player Login' })
  const updateRequest = page.waitForRequest((request) =>
    request.method() === 'PATCH' && request.url().includes('/api/v1/player_event_log_setting/101'),
  )

  await row.locator('select').first().selectOption('30')
  await updateRequest

  await expect(page.getByText('Server logs settings reloaded in-game!')).toBeVisible()
  expect(consoleErrors).toEqual([])
  expect(pageErrors).toEqual([])
})

test('zone servers restore query state and show filtered players', async ({ page }) => {
  const { consoleErrors, pageErrors } = await bootstrapSmokeHarness(page)

  await page.goto('/admin/zones?search=qeynos&showPlayers=true')

  await expect(page.getByPlaceholder('Search zone servers by zone name or player name...')).toHaveValue('qeynos')
  await expect(page.getByRole('cell', { name: 'qeynos' })).toBeVisible()
  await expect(page.getByText('Alyra')).toBeVisible()
  await expect(page.getByText('West Freeport')).toHaveCount(0)
  expect(consoleErrors).toEqual([])
  expect(pageErrors).toEqual([])
})

test('file logs list renders, filters, and streams a watched log file', async ({ page }) => {
  const { consoleErrors, pageErrors } = await bootstrapSmokeHarness(page)

  await page.goto('/admin/file-logs')

  await expect(page.getByText('Files (3)')).toBeVisible()
  await expect(page.getByText('zone-eqemu.log')).toBeVisible()
  await expect(page.getByText('world-eqemu.log')).toBeVisible()

  await page.getByRole('button', { name: 'Zone' }).click()

  await expect(page.getByText('zone-eqemu.log')).toBeVisible()
  await expect(page.getByText('world-eqemu.log')).toHaveCount(0)

  await page.getByTitle('View and watch log file').first().click()

  await expect(page.getByText('Line Buffer')).toBeVisible()
  await expect(page.locator('#file-contents')).toContainText('Zone booted')
  expect(consoleErrors).toEqual([])
  expect(pageErrors).toEqual([])
})

test('user connections render and the manage developer modal opens', async ({ page }) => {
  const { consoleErrors, pageErrors } = await bootstrapSmokeHarness(page)

  await page.goto('/connections')

  await expect(page.getByText('User Database Connections (1)')).toBeVisible()
  await expect(page.getByRole('heading', { name: 'Test Production DB' })).toBeVisible()

  await page.getByTitle('DevAlice').click()

  await expect(page.getByText('Manage Developer [DevAlice] for connection [Test Production DB]')).toBeVisible()
  expect(consoleErrors).toEqual([])
  expect(pageErrors).toEqual([])
})
