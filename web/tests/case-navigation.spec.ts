import { test, expect } from './auth'

test('mobile case selection opens focused details and preserves the filtered inbox', async ({
  page,
}) => {
  await page.setViewportSize({ width: 390, height: 844 })
  await page.goto('/')
  const detect = page.getByRole('button', { name: 'Jalankan deteksi', exact: true })
  await expect(detect).toHaveCSS('font-size', '14px')
  await detect.click()
  const search = page.getByRole('textbox', { name: 'Cari kasus' })
  await search.fill('OT-001')
  await expect(page.locator('.case-item')).toHaveCount(1)
  await page.locator('.case-item').click()

  const panel = page.getByRole('region', { name: 'Detail kasus' })
  await expect(panel).toBeFocused()
  await expect(page.locator('.inbox')).toBeHidden()
  const back = page.getByRole('button', { name: 'Kembali ke daftar kasus' })
  await expect(back).toBeInViewport()
  await expect(page.getByRole('tab', { name: /Laporan bukti/ })).toHaveAttribute(
    'aria-selected',
    'true',
    { timeout: 30000 },
  )
  await expect(page.getByRole('button', { name: 'Tinjau rekomendasi' })).toBeVisible()

  await page.getByRole('tab', { name: /Review manusia/ }).click()
  const supportingEvidence = page.locator('.recommendation .evidence-links button').first()
  await supportingEvidence.click()
  await expect(page.getByRole('dialog')).toBeVisible()
  await expect(page.getByRole('dialog')).toContainText('SHA-256 snapshot')
  await page.getByRole('button', { name: 'Tutup bukti' }).click()

  await back.click()
  await expect(search).toHaveValue('OT-001')
  await expect(page.locator('.case-item.active')).toBeFocused()
  await expect(panel).toBeHidden()
  await expect
    .poll(() => page.evaluate(() => document.documentElement.scrollWidth <= innerWidth))
    .toBe(true)

  await page.getByTestId('language-picker').selectOption('en')
  await page.locator('.case-item').click()
  await expect(page.getByRole('button', { name: 'Back to cases' })).toBeInViewport()
  await expect(page.getByRole('button', { name: 'Review recommendations' })).toBeVisible()
  await page.screenshot({ path: test.info().outputPath('case-detail-mobile.png'), fullPage: true })
})

test('an incomplete report stays flagged and its process is available without starting another run', async ({
  page,
}) => {
  const writes: string[] = []
  page.on('request', (request) => {
    if (request.method() === 'POST') writes.push(request.url())
  })
  await page.route(/\/api\/v1\/cases\/[^/?]+$/, async (route) => {
    const response = await route.fetch()
    const detail = await response.json()
    if (detail.latest_run?.report) {
      detail.latest_run.status = 'insufficient_evidence'
      detail.recommendations = []
    }
    await route.fulfill({ response, json: detail })
  })
  await page.route(/\/api\/v1\/investigations\/[^/?]+$/, async (route) => {
    const response = await route.fetch()
    const detail = await response.json()
    detail.run.status = 'insufficient_evidence'
    await route.fulfill({ response, json: detail })
  })
  await page.goto('/')
  await expect(page.locator('.case-next-step')).toContainText('Bukti belum cukup')
  await expect(page.getByRole('button', { name: 'Tinjau rekomendasi' })).toHaveCount(0)
  await page.getByRole('button', { name: 'Tinjau proses investigasi' }).click()
  await expect(page.locator('.process-details')).not.toHaveAttribute('open')
  await page.locator('.process-details summary').click()
  await expect(page.locator('.timeline')).toBeVisible()
  await expect(page.locator('.timeline')).toContainText('Bukti tersimpan')
  expect(writes).toEqual([])
})
