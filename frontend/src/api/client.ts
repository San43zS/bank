import type { ApiError } from './types'

const BASE_URL = (import.meta as any).env?.VITE_API_BASE_URL ?? 'http://localhost:8080'

export class HttpError extends Error {
  status: number
  body?: unknown

  constructor(message: string, status: number, body?: unknown) {
    super(message)
    this.name = 'HttpError'
    this.status = status
    this.body = body
  }
}

async function readJsonSafe(res: Response): Promise<unknown | undefined> {
  const text = await res.text()
  if (!text) return undefined
  try {
    return JSON.parse(text)
  } catch {
    return { error: text }
  }
}

export async function apiFetch<T>(
  path: string,
  opts: {
    method?: string
    token?: string | null
    body?: unknown
    query?: Record<string, string | number | undefined | null>
  } = {},
): Promise<T> {
  const url = new URL(path, BASE_URL)
  if (opts.query) {
    for (const [k, v] of Object.entries(opts.query)) {
      if (v === undefined || v === null || v === '') continue
      url.searchParams.set(k, String(v))
    }
  }

  const res = await fetch(url.toString(), {
    method: opts.method ?? (opts.body ? 'POST' : 'GET'),
    headers: {
      ...(opts.body ? { 'Content-Type': 'application/json' } : {}),
      ...(opts.token ? { Authorization: `Bearer ${opts.token}` } : {}),
    },
    body: opts.body ? JSON.stringify(opts.body) : undefined,
  })

  if (!res.ok) {
    const body = (await readJsonSafe(res)) as ApiError | undefined
    const msg =
      (body && typeof body === 'object' && 'error' in body && typeof (body as any).error === 'string'
        ? (body as any).error
        : `http_${res.status}`) || `http_${res.status}`
    throw new HttpError(msg, res.status, body)
  }

  const json = (await readJsonSafe(res)) as T | undefined
  return (json as T) ?? (undefined as T)
}

