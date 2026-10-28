import { test, expect } from '@playwright/test'
test('P1 cases expose confirmed sources without changing golden', async ({ page }) => {
  test.skip(process.env.ENABLE_P1_RULES !== 'true', 'P1 dipilih eksplisit.')
  const errors: string[] = []
  page.on('pageerror', (e) => errors.push(e.message))
  await page.goto('/')
  await page.getByLabel('Nama pengguna').fill(process.env.OPERATOR_USERNAME || 'operator')
  await page.getByLabel('Kata sandi', { exact: true }).fill(process.env.OPERATOR_PASSWORD!)
  await page.getByRole('button', { name: 'Masuk', exact: true }).click()
  await expect(page.getByText('8 aturan aktif', { exact: true })).toBeVisible()
  await page.getByRole('button', { name: 'Jalankan deteksi', exact: true }).click()
  for (const rule of ['ATT-001', 'OT-002', 'COST-001', 'COST-002'])
    await expect(page.locator('.case-item').filter({ hasText: rule })).toBeVisible()
  await page.locator('.case-item').filter({ hasText: 'ATT-001' }).click()
  await expect(
    page.getByRole('heading', { name: 'Rangkaian absen terkonfirmasi', level: 2, exact: true }),
  ).toBeVisible()
  const start = page.getByRole('button', { name: 'Mulai investigasi', exact: true })
  if (await start.isVisible()) await start.click()
  await expect(page.getByRole('button', { name: 'Buka laporan bukti' })).toBeVisible({
    timeout: 30000,
  })
  await page.getByRole('button', { name: 'Buka laporan bukti' }).click()
  const attendance = page.locator('.evidence-card').filter({ hasText: 'Rekonsiliasi kehadiran' })
  await expect(attendance).toContainText('Pekerja dengan rangkaian absen')
  await expect(attendance).toContainText('Shift berturut-turut')
  await attendance.click()
  await expect(page.getByRole('dialog')).toContainText('SHA-256 snapshot')
  await page.getByRole('button', { name: 'Tutup bukti' }).click()
  await page.screenshot({ path: test.info().outputPath('p1-report-desktop.png'), fullPage: true })
  await page.locator('.case-item').filter({ hasText: 'OT-001' }).click()
  await expect(page.locator('.case-numbers')).toContainText('38%')
  expect(errors).toEqual([])
})
