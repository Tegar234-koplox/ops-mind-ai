import { test, expect } from '@playwright/test'

test('shift handover preserves versions, facts and printable review context', async ({ page }) => {
  const errors: string[] = []
  page.on('pageerror', (error) => errors.push(error.message))
  await page.goto('/')
  await page.getByLabel('Nama pengguna').fill(process.env.OPERATOR_USERNAME || 'operator')
  await page.getByLabel('Kata sandi', { exact: true }).fill(process.env.OPERATOR_PASSWORD!)
  await page.getByRole('button', { name: 'Masuk', exact: true }).click()
  await page.getByRole('button', { name: 'Jalankan deteksi', exact: true }).click()
  await page.locator('.case-item').filter({ hasText: 'OT-001' }).click()
  await expect(page.getByRole('button', { name: 'Buka laporan bukti' })).toBeVisible({
    timeout: 30000,
  })
  await page.getByRole('button', { name: 'Serah terima shift', exact: true }).click()
  await expect(page.getByRole('heading', { name: 'Serah terima shift.' })).toBeVisible()
  await page.getByLabel('Lini produksi', { exact: true }).selectOption('A')
  await page.getByLabel('Tanggal operasional', { exact: true }).fill('2026-10-26')
  await expect(page.getByLabel('Shift selesai')).toContainText('2026-10-26')
  await page.getByLabel('Shift selesai').selectOption('A-20261026-S1')
  await page.getByRole('button', { name: 'Susun laporan baru' }).click()
  const report = page.getByRole('article', { name: 'Laporan serah terima tersimpan' })
  await expect(report).toBeVisible()
  await expect(
    report.locator('.handover-fact-section').filter({ hasText: 'Kehadiran dan cakupan operator' }),
  ).toContainText('31')
  await expect(
    report.locator('.handover-fact-section').filter({ hasText: 'Target dan hasil produksi' }),
  ).toContainText('17%')
  await expect(
    report.locator('.handover-fact-section').filter({ hasText: 'Lembur dan pengeluaran' }),
  ).toContainText('38%')
  await expect(
    report
      .locator('.handover-fact-section')
      .filter({ hasText: 'Gangguan mesin dan downtime lini' }),
  ).toContainText('110 menit')
  await expect(report).toContainText('Catatan pending bukan bukti mangkir')
  await expect(report).toContainText('Target produksi')
  await expect(report).toContainText('Output aktual')
  await expect(report).toContainText('SIMULASI · Sintesis investigasi tersimpan')
  const firstID = await page.getByLabel('Versi laporan tersimpan').inputValue()
  let releaseHistory = () => {}
  const historyBlocked = new Promise<void>((resolve) => {
    releaseHistory = resolve
  })
  await page.route(
    (url) => url.pathname === '/api/v1/handovers' && url.searchParams.has('shift_id'),
    async (route) => {
      await historyBlocked
      await route.continue()
    },
  )
  await page.getByRole('button', { name: 'Susun laporan baru' }).click()
  try {
    await expect(page.getByLabel('Versi laporan tersimpan')).not.toHaveValue(firstID)
    const newID = await page.getByLabel('Versi laporan tersimpan').inputValue()
    await expect(report).toContainText(newID)
  } finally {
    releaseHistory()
  }
  await page.unrouteAll({ behavior: 'wait' })
  await page.getByLabel('Versi laporan tersimpan').selectOption(firstID)
  await expect(report).toContainText(firstID)
  await page.reload()
  await page.getByRole('button', { name: 'Serah terima shift', exact: true }).click()
  await page.getByLabel('Lini produksi', { exact: true }).selectOption('A')
  await page.getByLabel('Tanggal operasional', { exact: true }).fill('2026-10-26')
  await page.getByLabel('Shift selesai').selectOption('A-20261026-S1')
  await page.getByLabel('Versi laporan tersimpan').selectOption(firstID)
  await expect(report).toContainText(firstID)
  let release = () => {}
  let signalStarted = () => {}
  const blocked = new Promise<void>((resolve) => {
    release = resolve
  })
  const started = new Promise<void>((resolve) => {
    signalStarted = resolve
  })
  await page.route('**/api/v1/handovers', async (route) => {
    if (route.request().method() === 'POST') {
      signalStarted()
      await blocked
    }
    await route.continue()
  })
  const savedResponse = page.waitForResponse(
    (response) =>
      response.request().method() === 'POST' && response.url().endsWith('/api/v1/handovers'),
  )
  await page.getByRole('button', { name: 'Susun laporan baru' }).click()
  await started
  try {
    await page.getByLabel('Lini produksi', { exact: true }).selectOption('B')
    await expect(page.getByLabel('Shift selesai')).toContainText('Assembly Line B')
    await expect(report).toHaveCount(0)
    await expect(page.getByRole('button', { name: 'Menyusun…', exact: true })).toBeDisabled()
  } finally {
    release()
  }
  await savedResponse
  await expect(page.getByRole('button', { name: 'Susun laporan baru' })).toBeEnabled()
  await expect(report).toHaveCount(0)
  await page.unrouteAll({ behavior: 'wait' })
  await page.getByLabel('Lini produksi', { exact: true }).selectOption('A')
  await page.getByLabel('Shift selesai').selectOption('A-20261026-S1')
  await page.getByLabel('Versi laporan tersimpan').selectOption(firstID)
  await expect(report).toContainText(firstID)
  await page.evaluate(() => {
    window.print = () => {
      document.documentElement.dataset.printRequested = 'true'
    }
  })
  await page.getByRole('button', { name: 'Cetak / Simpan PDF' }).click()
  await expect(page.locator('html')).toHaveAttribute('data-print-requested', 'true')
  await page.emulateMedia({ media: 'print' })
  await expect(page.locator('.handover-controls')).toBeHidden()
  await expect(page.locator('.sidebar')).toBeHidden()
  await expect(report).toBeVisible()
  await page.screenshot({ path: test.info().outputPath('handover-print.png'), fullPage: true })
  await page.emulateMedia({ media: 'screen' })
  await page.evaluate(() => window.scrollTo(0, 0))
  await page.screenshot({ path: test.info().outputPath('handover-desktop.png'), fullPage: true })
  await page.setViewportSize({ width: 390, height: 844 })
  await expect(page.getByRole('button', { name: 'Serah terima shift', exact: true })).toBeVisible()
  await expect
    .poll(() => page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth))
    .toBe(true)
  await page.screenshot({ path: test.info().outputPath('handover-mobile.png'), fullPage: true })
  await page.getByRole('button', { name: 'Telusuri kasus OT-001', exact: true }).click()
  await expect(page.locator('.case-item.active')).toContainText('OT-001')
  expect(errors).toEqual([])
})
