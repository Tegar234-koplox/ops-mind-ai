import { test as base, expect } from '@playwright/test'

// Non-authentication scenarios reuse one session per worker. Login behavior is
// exercised through the UI in bilingual.spec.ts and workflow.spec.ts.
export const test = base.extend<{}, { operatorSession: string }>({
  operatorSession: [
    async ({ playwright }, use) => {
      const request = await playwright.request.newContext({
        baseURL: process.env.APP_ORIGIN || 'http://localhost:8080',
        extraHTTPHeaders: { Origin: process.env.APP_ORIGIN || 'http://localhost:8080' },
      })
      try {
        const response = await request.post('/api/v1/auth/login', {
          data: {
            username: process.env.OPERATOR_USERNAME || 'operator',
            password: process.env.OPERATOR_PASSWORD,
          },
        })
        expect(response.ok(), 'operator session must authenticate').toBeTruthy()
        await use(JSON.stringify(await request.storageState()))
      } finally {
        await request.dispose()
      }
    },
    { scope: 'worker' },
  ],
  storageState: async ({ operatorSession }, use) => {
    await use(JSON.parse(operatorSession))
  },
})

export { expect }
