export interface ErrorResponse {
  error?: string
  message?: string
}

const API_BASE_URL = trimTrailingSlash(import.meta.env.VITE_API_BASE_URL || '')

function trimTrailingSlash(value: string): string {
  return value.replace(/\/+$/, '')
}

function withBaseUrl(path: string): string {
  if (!API_BASE_URL || /^https?:\/\//.test(path)) {
    return path
  }

  if (path.startsWith('/')) {
    return `${API_BASE_URL}${path}`
  }

  return `${API_BASE_URL}/${path}`
}

async function readErrorMessage(response: Response): Promise<string> {
  try {
    const payload = await response.json() as ErrorResponse
    return payload.message || payload.error || `Request failed (${response.status})`
  } catch {
    return `Request failed (${response.status})`
  }
}

function buildUrl(path: string, params?: Record<string, string | number | undefined>): string {
  const urlPath = withBaseUrl(path)

  if (!params) {
    return urlPath
  }

  const query = new URLSearchParams()

  for (const [key, value] of Object.entries(params)) {
    if (value !== undefined) {
      query.set(key, String(value))
    }
  }

  const serialized = query.toString()
  if (!serialized) {
    return urlPath
  }

  return `${urlPath}?${serialized}`
}

function getCsrfToken(): string | null {
  const cookies = document.cookie.split('; ')
  const csrfCookie = cookies.find(row => row.startsWith('sa_csrf='))
  return csrfCookie ? csrfCookie.split('=')[1] : null
}

function requiresCsrfToken(method: string): boolean {
  return ['POST', 'PATCH', 'PUT', 'DELETE'].includes(method.toUpperCase())
}

export async function requestJson<T>(
  path: string,
  init?: RequestInit,
  params?: Record<string, string | number | undefined>,
): Promise<T> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...init?.headers,
  }

  // Add CSRF token for state-changing requests
  const method = init?.method || 'GET'
  if (requiresCsrfToken(method)) {
    const csrfToken = getCsrfToken()
    if (csrfToken) {
      headers['X-CSRF-Token'] = csrfToken
    }
  }

  const response = await fetch(buildUrl(path, params), {
    ...init,
    credentials: 'include',
    headers,
  })

  if (!response.ok) {
    throw new Error(await readErrorMessage(response))
  }

  return response.json() as Promise<T>
}

export async function requestNoContent(path: string, init?: RequestInit): Promise<void> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...init?.headers,
  }

  // Add CSRF token for state-changing requests
  const method = init?.method || 'GET'
  if (requiresCsrfToken(method)) {
    const csrfToken = getCsrfToken()
    if (csrfToken) {
      headers['X-CSRF-Token'] = csrfToken
    }
  }

  const response = await fetch(path, {
    ...init,
    credentials: 'include',
    headers,
  })

  if (!response.ok) {
    throw new Error(await readErrorMessage(response))
  }
}
