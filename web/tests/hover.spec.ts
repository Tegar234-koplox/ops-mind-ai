import { test, expect, type Locator, type Page } from '@playwright/test'

async function expectHoverFeedback(page: Page, target: Locator) {
  await page.mouse.move(1400, 30)
  await page.waitForTimeout(200) // Biarkan transisi keluar selesai sebelum membandingkan hover.
  const initial = await target.evaluate((element) => getComputedStyle(element).backgroundColor)
  await target.hover()
  await expect
    .poll(() => target.evaluate((element) => getComputedStyle(element).backgroundColor))
    .not.toBe(initial)
}

test('hover, keyboard focus and navigation state remain consistent', async ({ page }) => {
  await page.goto('/')
  await page.getByLabel('Nama pengguna').fill(process.env.OPERATOR_USERNAME || 'operator')
  await page.getByLabel('Kata sandi', { exact: true }).fill(process.env.OPERATOR_PASSWORD!)
  await page.getByRole('button', { name: 'Masuk', exact: true }).click()
  await page.getByRole('button', { name: 'Jalankan deteksi', exact: true }).click()
  await page.locator('.case-item').filter({ hasText: 'OT-001' }).click()
  await expect(page.getByRole('button', { name: 'Buka laporan bukti' })).toBeVisible({
    timeout: 30000,
  })

  const reportNav = page.getByRole('navigation').getByRole('button', { name: 'Laporan bukti' })
  await expectHoverFeedback(page, reportNav)
  await expectHoverFeedback(page, page.getByRole('tab', { name: /Laporan bukti/ }))
  await expectHoverFeedback(page, page.locator('.case-item.active'))
  await expectHoverFeedback(page, page.locator('.case-item:not(.active)').first())
  await expectHoverFeedback(
    page,
    page.getByRole('button', { name: 'Jalankan deteksi', exact: true }),
  )

  await reportNav.click()
  await expect(reportNav).toHaveAttribute('aria-current', 'page')
  await expect(page.getByRole('navigation').locator('[aria-current="page"]')).toHaveCount(1)
  await expect(page.getByRole('tab', { name: /Laporan bukti/ })).toHaveAttribute(
    'aria-selected',
    'true',
  )
  await expectHoverFeedback(page, reportNav)
  await expectHoverFeedback(page, page.getByRole('tab', { name: /Laporan bukti/ }))
  await expectHoverFeedback(page, page.locator('.evidence-card').first())
  const evidenceLink = page.locator('.evidence-links button').first()
  await expectHoverFeedback(page, evidenceLink)
  await evidenceLink.click()
  await expect(page.getByRole('dialog')).toBeVisible()
  await page.getByRole('button', { name: 'Tutup bukti' }).click()

  await page.getByRole('button', { name: 'Keluar', exact: true }).focus()
  await page.keyboard.press('Tab')
  const detect = page.getByRole('button', { name: 'Jalankan deteksi', exact: true })
  await expect(detect).toBeFocused()
  await expect(detect).toHaveCSS('outline-style', 'solid')

  await page.emulateMedia({ reducedMotion: 'reduce' })
  await expect(reportNav).toHaveCSS('transition-duration', '0s')
  await page.evaluate(() => window.scrollTo(0, 0))
  await reportNav.hover()
  await page.screenshot({ path: test.info().outputPath('hover-report-desktop.png') })

  await page.setViewportSize({ width: 390, height: 844 })
  await page.getByRole('tab', { name: /Review manusia/ }).click()
  await expect(page.getByRole('tab', { name: /Review manusia/ })).toHaveAttribute(
    'aria-selected',
    'true',
  )
  await expect(page.getByText('Anda memegang keputusan akhir.', { exact: true })).toBeVisible()
  await expect
    .poll(() => page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth))
    .toBe(true)
  await page.evaluate(() => window.scrollTo(0, 0))
  await page.screenshot({ path: test.info().outputPath('navigation-mobile.png') })
})
