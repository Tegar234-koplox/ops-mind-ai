export class ApiError extends Error {
  status: number
  code: string
  constructor(message: string, status: number, code = '') {
    super(message)
    this.status = status
    this.code = code
  }
}
export async function api<T>(path: string, body?: unknown, signal?: AbortSignal): Promise<T> {
  const res = await fetch('/api/v1' + path, {
    credentials: 'same-origin',
    signal,
    ...(body === undefined
      ? {}
      : {
          method: 'POST',
          headers: { 'Content-Type': 'application/json', 'Idempotency-Key': crypto.randomUUID() },
          body: JSON.stringify(body),
        }),
  }).catch((error: unknown) => {
    if (signal?.aborted) throw error
    throw new ApiError('Kesalahan jaringan. Periksa koneksi dan coba lagi.', 0, 'network_error')
  })
  const data = await res.json().catch(() => null)
  if (!res.ok)
    throw new ApiError(
      data?.error?.message ?? 'Layanan belum siap. Coba lagi setelah beberapa saat.',
      res.status,
      data?.error?.code,
    )
  return data as T
}
export const active = (s?: string) => s === 'queued' || s === 'running'
