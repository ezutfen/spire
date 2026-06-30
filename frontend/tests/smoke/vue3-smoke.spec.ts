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
