import { config } from '../config/env'
import { getCookie } from '../lib/cookies'

/** Matches backend platform/auth CSRFCookieName. */
export const CSRF_COOKIE = 'csrf_token'
/** Matches backend platform/auth CSRFHeaderName. */
export const CSRF_HEADER = 'X-CSRF-Token'

const SAFE_METHODS = new Set(['GET', 'HEAD', 'OPTIONS'])

/** Paths relative to apiBaseUrl that bootstrap auth and stay CSRF-exempt. */
const CSRF_EXEMPT_SUFFIXES = ['/auth/register', '/auth/login']

export type ApiErrorBody = {
  code: string
  message: string
  status: number
  details?: Record<string, string>
}

export class ApiError extends Error {
  readonly code: string
  readonly status: number
  readonly details?: Record<string, string>

  constructor(body: ApiErrorBody) {
    super(body.message)
    this.name = 'ApiError'
    this.code = body.code
    this.status = body.status
    this.details = body.details
  }
}

export type ApiRequestInit = Omit<RequestInit, 'body'> & {
  body?: unknown
  /** Skip CSRF header even on mutating methods (rarely needed). */
  skipCsrf?: boolean
}

function joinUrl(base: string, path: string): string {
  const b = base.replace(/\/+$/, '')
  const p = path.startsWith('/') ? path : `/${path}`
  return `${b}${p}`
}

function isCsrfExempt(path: string): boolean {
  const normalized = path.startsWith('/') ? path : `/${path}`
  return CSRF_EXEMPT_SUFFIXES.some(
    (suffix) => normalized === suffix || normalized.endsWith(suffix),
  )
}

function shouldAttachCsrf(method: string, path: string, skipCsrf?: boolean): boolean {
  if (skipCsrf) return false
  if (SAFE_METHODS.has(method.toUpperCase())) return false
  if (isCsrfExempt(path)) return false
  return true
}

/**
 * Fetch wrapper: credentials (cookies) always on; reads csrf_token and sends
 * X-CSRF-Token on mutating requests except register/login.
 */
export async function api<T>(path: string, init: ApiRequestInit = {}): Promise<T> {
  const method = (init.method ?? 'GET').toUpperCase()
  const headers = new Headers(init.headers)

  if (init.body !== undefined && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json')
  }

  if (shouldAttachCsrf(method, path, init.skipCsrf)) {
    const csrf = getCookie(CSRF_COOKIE)
    if (csrf) {
      headers.set(CSRF_HEADER, csrf)
    }
  }

  const { body: rawBody, skipCsrf: _skip, ...rest } = init
  const res = await fetch(joinUrl(config.apiBaseUrl, path), {
    ...rest,
    method,
    headers,
    credentials: 'include',
    body: rawBody === undefined ? undefined : JSON.stringify(rawBody),
  })

  if (res.status === 204) {
    return undefined as T
  }

  const text = await res.text()
  const data = text ? (JSON.parse(text) as unknown) : null

  if (!res.ok) {
    const err = data as ApiErrorBody | null
    throw new ApiError({
      code: err?.code ?? 'HTTP_ERROR',
      message: err?.message ?? (res.statusText || 'Request failed'),
      status: err?.status ?? res.status,
      details: err?.details,
    })
  }

  return data as T
}

/** Current csrf_token cookie value (empty if not set yet). */
export function readCsrfToken(): string {
  return getCookie(CSRF_COOKIE)
}
