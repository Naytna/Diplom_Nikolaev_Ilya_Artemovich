export const API_URL = 'http://localhost:28080/api'

export type UserRole = 'expert' | 'student'
export type DemoRole = 'guest' | 'learner' | 'expert'

export type AuthUser = {
  id: number
  role: UserRole
  full_name: string
}

type ApiOptions = RequestInit & {
  authToken?: string | null
  onUnauthorized?: () => void
}

let currentDemoRole: DemoRole = 'guest'

type ApiErrorPayload = {
  error?: string
}

export class ApiError extends Error {
  status: number

  constructor(status: number, message: string) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

export function setDemoRole(role: DemoRole) {
  currentDemoRole = role
}

export function getDemoRole() {
  return currentDemoRole
}

export async function api<T>(path: string, options: ApiOptions = {}): Promise<T> {
  const headers = new Headers(options.headers ?? {})

  if (!headers.has('Content-Type') && options.body && !(options.body instanceof FormData)) {
    headers.set('Content-Type', 'application/json')
  }

  if (options.authToken) {
    headers.set('Authorization', `Bearer ${options.authToken}`)
  }

  headers.set('X-Demo-Role', currentDemoRole)

  const response = await fetch(`${API_URL}${path}`, {
    ...options,
    headers,
  })

  const text = await response.text()
  const parsed = tryParsePayload(text)

  if (!response.ok) {
    if (response.status === 401) {
      options.onUnauthorized?.()
    }

    const message = parsed?.error || text || `Ошибка запроса: ${response.status}`
    throw new ApiError(response.status, message)
  }

  if (!text) {
    return null as T
  }

  return (parsed ?? JSON.parse(text)) as T
}

export async function publicApi<T>(path: string, options: RequestInit = {}): Promise<T> {
  const headers = new Headers(options.headers ?? {})

  if (!headers.has('Content-Type') && options.body && !(options.body instanceof FormData)) {
    headers.set('Content-Type', 'application/json')
  }

  headers.set('X-Demo-Role', currentDemoRole)

  const response = await fetch(`${API_URL}${path}`, {
    ...options,
    headers,
  })

  const text = await response.text()
  const parsed = tryParsePayload(text)

  if (!response.ok) {
    const message = parsed?.error || text || `Ошибка запроса: ${response.status}`
    throw new ApiError(response.status, message)
  }

  if (!text) {
    return null as T
  }

  return (parsed ?? JSON.parse(text)) as T
}

export function tryParsePayload(value: string): ApiErrorPayload | null {
  if (!value) {
    return null
  }

  try {
    return JSON.parse(value) as ApiErrorPayload
  } catch {
    return null
  }
}
