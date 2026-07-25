import { api } from '../../../shared/api'

export type PublicUser = {
  id: string
  email: string
  username: string
  display_name: string
  created_at: string
}

export type AuthResponse = {
  user: PublicUser
}

export type RegisterInput = {
  email: string
  username: string
  password: string
  display_name?: string
}

export type LoginInput = {
  email: string
  password: string
}

/** POST /auth/register — CSRF-exempt; sets access + refresh + csrf cookies. */
export function register(input: RegisterInput): Promise<AuthResponse> {
  return api<AuthResponse>('/auth/register', {
    method: 'POST',
    body: input,
  })
}

/** POST /auth/login — CSRF-exempt; sets access + refresh + csrf cookies. */
export function login(input: LoginInput): Promise<AuthResponse> {
  return api<AuthResponse>('/auth/login', {
    method: 'POST',
    body: input,
  })
}

/** POST /auth/refresh — sends X-CSRF-Token from csrf_token cookie. */
export function refresh(): Promise<AuthResponse> {
  return api<AuthResponse>('/auth/refresh', {
    method: 'POST',
  })
}

/** POST /auth/logout — sends X-CSRF-Token from csrf_token cookie. */
export function logout(): Promise<void> {
  return api<void>('/auth/logout', {
    method: 'POST',
  })
}

/** GET /users/me — cookie session; safe method (no CSRF). */
export function me(): Promise<{ user: PublicUser }> {
  return api<{ user: PublicUser }>('/users/me')
}
