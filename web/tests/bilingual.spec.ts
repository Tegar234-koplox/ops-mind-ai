import { test, expect } from '@playwright/test'

test('language preference, English investigation and immutable handover work together', async ({
  page,
}) => {
  const errors: string[] = []
  const writes: string[] = []
  page.on('pageerror', (e) => errors.push(e.message))
  page.on('request', (request) => {
    if (request.method() === 'POST') writes.push(request.url())
  })
  await page.goto('/')
  await page.getByTestId('language-picker').selectOption('en')
  await expect(page.locator('html')).toHaveAttribute('lang', 'en')
  await expect(page).toHaveTitle('OpsMind AI — Investigation Workspace')
  await page
    .getByLabel('Username', { exact: true })
    .fill(process.env.OPERATOR_USERNAME || 'operator')
  await page.getByLabel('Password', { exact: true }).fill('intentionally-invalid-password')
  await page.getByRole('button', { name: 'Sign in', exact: true }).click()
  await expect(page.getByRole('alert')).toHaveText('Invalid credentials.')
  await page.getByTestId('language-picker').selectOption('id')
  await expect(page.getByRole('alert')).toHaveText('Kredensial tidak cocok.')
  await expect(page.getByLabel('Nama pengguna')).toHaveValue(
    process.env.OPERATOR_USERNAME || 'operator',
  )
  await page.getByTestId('language-picker').selectOption('en')
  await page.getByLabel('Password', { exact: true }).fill(process.env.OPERATOR_PASSWORD!)
  await page.getByRole('button', { name: 'Sign in', exact: true }).click()
  await expect(page.getByRole('heading', { name: 'Investigation workspace.' })).toBeVisible()
  await expect(page.getByText('SIMULATED MODEL', { exact: true })).toBeVisible()
  const detected = page.waitForResponse(
    (r) => r.request().method() === 'POST' && r.url().endsWith('/api/v1/detections'),
  )
  await page.getByRole('button', { name: 'Run detection', exact: true }).click()
  const detection = await detected
  expect(detection.request().postDataJSON().report_language).toBe('en')
  await page.locator('.case-item').filter({ hasText: 'OT-001' }).click()
  await page.getByTestId('language-picker').selectOption('id')
  await expect(page.getByRole('tab', { name: /Laporan bukti/ })).toHaveAttribute(
    'aria-selected',
    'true',
    { timeout: 30000 },
  )
  const cases = await page.request.get('/api/v1/cases?limit=100').then((r) => r.json())
  const primary = cases.items.find(
    (item: { rule_id: string; shift_id: string }) =>
      item.rule_id === 'OT-001' && item.shift_id === 'A-20261026-S1',
  )
  const detail = await page.request.get('/api/v1/cases/' + primary.id).then((r) => r.json())
  expect(detail.latest_run.report_language).toBe('en')
  await expect(page.locator('.report-intro p')).toContainText('SIMULATION:')
  await expect(page.locator('.report-intro p')).toHaveAttribute('lang', 'en')
  const beforeToggle = writes.length
  await page.getByTestId('language-picker').selectOption('en')
  await expect(
    page.getByRole('heading', { name: 'Observations from source records' }),
  ).toBeVisible()
  const staffing = page.locator('.evidence-card').filter({ hasText: 'Operator availability' })
  await staffing.click()
  await expect(page.getByRole('dialog')).toContainText(
    'Unverified records do not establish unexplained absence',
  )
  await page.getByRole('button', { name: 'Close evidence' }).click()
  await expect(
    page.locator('.evidence-card').filter({ hasText: 'Machine interruptions' }),
  ).toContainText('110 minutes')
  await expect(
    page.locator('.evidence-card').filter({ hasText: 'Overtime cost variance' }),
  ).toContainText('1,380,000')
  expect(writes.length).toBe(beforeToggle)
  await page.getByRole('button', { name: 'Review recommendations' }).click()
  await expect(page.getByRole('button', { name: 'Approve recommendation' }).first()).toBeVisible()
  await expect(
    page.getByText('Business data and worker schedules do not change automatically.', {
      exact: false,
    }),
  ).toBeVisible()
  await page.getByRole('button', { name: 'Shift handover', exact: true }).click()
  await page.getByLabel('Production line', { exact: true }).selectOption('A')
  await page.getByLabel('Operational date', { exact: true }).fill('2026-10-26')
  await page.getByLabel('Completed shift').selectOption('A-20261026-S1')
  await page.getByRole('button', { name: 'Create new report' }).click()
  const report = page.getByRole('article', { name: 'Stored shift handover report' })
  await expect(report).toBeVisible()
  await expect(report).toContainText('Attendance and operator coverage')
  await expect(report.locator('.handover-limitations')).toContainText(
    'The report is a snapshot at creation time.',
  )
  await expect(page.getByRole('button', { name: 'Explore case OT-001', exact: true })).toBeVisible()
  await expect(report).toContainText('Pending records do not establish unexplained absence')
  await expect(report).toContainText('110 minutes')
  await expect(report).toContainText('38%')
  await expect(report).toContainText('17%')
  const reportID = await page.getByLabel('Stored report version').inputValue()
  const original = await page.request.get('/api/v1/handovers/' + reportID).then((r) => r.json())
  await page.screenshot({
    path: test.info().outputPath('bilingual-handover-en.png'),
    fullPage: true,
  })
  const beforeHandoverToggle = writes.length
  await page.getByTestId('language-picker').selectOption('id')
  await expect(page.getByRole('article', { name: 'Laporan serah terima tersimpan' })).toContainText(
    '110 menit',
  )
  const after = await page.request.get('/api/v1/handovers/' + reportID).then((r) => r.json())
  expect(after).toEqual(original)
  expect(writes.length).toBe(beforeHandoverToggle)
  await page.getByTestId('language-picker').selectOption('en')
  await page.emulateMedia({ media: 'print' })
  await expect(page.getByRole('button', { name: 'Print / Save PDF' })).toBeHidden()
  await expect(report).toBeVisible()
  await page.emulateMedia({ media: 'screen' })
  await page.setViewportSize({ width: 390, height: 844 })
  await expect(page.getByTestId('language-picker')).toBeVisible()
  await expect(page.getByRole('article', { name: 'Stored shift handover report' })).toBeVisible()
  expect(
    await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth),
  ).toBeTruthy()
  await page.screenshot({ path: test.info().outputPath('bilingual-mobile-en.png'), fullPage: true })
  await page.reload()
  await expect(page.locator('html')).toHaveAttribute('lang', 'en')
  await expect(page.getByRole('heading', { name: 'Investigation workspace.' })).toBeVisible()
  await page.getByRole('button', { name: 'Sign out', exact: true }).click()
  await expect(page.getByLabel('Username', { exact: true })).toBeVisible()
  await expect(page.getByTestId('language-picker')).toHaveValue('en')
  expect(errors).toEqual([])
})

test('unsupported or blocked browser storage keeps the language control usable', async ({
  page,
}) => {
  await page.addInitScript(() => {
    if (location.search.includes('storage=blocked'))
      Object.defineProperty(window, 'localStorage', {
        get() {
          throw new DOMException('Storage unavailable', 'SecurityError')
        },
      })
    else localStorage.setItem('opsmind.locale.v1', 'unsupported')
  })
  await page.goto('/')
  await expect(page.locator('html')).toHaveAttribute('lang', 'id')
  await expect(page.getByTestId('language-picker')).toHaveValue('id')
  await page.goto('/?storage=blocked')
  await page.getByTestId('language-picker').selectOption('en')
  await expect(page.locator('html')).toHaveAttribute('lang', 'en')
  await expect(page.getByLabel('Username', { exact: true })).toBeVisible()
})
