// The single place that owns all fetch calls and error-envelope parsing.
// Components and pages never build URLs or call fetch directly.

export interface ErrorResponse {
  error: {
    code: string
    message?: string
    fields?: Record<string, string>
    sku?: string
    available?: number
  }
}

export class ApiError extends Error {
  readonly status: number
  readonly code: string
  readonly fields?: Record<string, string>
  readonly sku?: string
  readonly available?: number

  constructor(status: number, detail: ErrorResponse['error']) {
    super(detail.message ?? detail.code)
    this.name = 'ApiError'
    this.status = status
    this.code = detail.code
    this.fields = detail.fields
    this.sku = detail.sku
    this.available = detail.available
  }
}

export async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(path, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...(init?.headers ?? {}),
    },
  })

  if (res.status === 204) {
    return undefined as T
  }

  const text = await res.text()
  let body: unknown = null
  if (text) {
    try {
      body = JSON.parse(text)
    } catch {
      body = text
    }
  }

  if (!res.ok) {
    const detail = (body as ErrorResponse | null)?.error
    throw new ApiError(res.status, detail ?? { code: 'UNKNOWN', message: res.statusText })
  }

  return body as T
}
