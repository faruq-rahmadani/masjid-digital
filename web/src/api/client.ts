import axios, { type AxiosRequestConfig } from 'axios'
import { env } from '@/utils/env'

/** Error codes the Go API returns (pkg/apperror). */
export type ApiErrorCode =
  | 'INVALID_INPUT'
  | 'UNAUTHENTICATED'
  | 'FORBIDDEN'
  | 'NOT_FOUND'
  | 'ALREADY_EXISTS'
  | 'BUSINESS_RULE'
  | 'INTERNAL'

/** Successful payloads from the API are wrapped as { data }. */
export type DataResponse<T> = { data: T }
export type MessageResponse = { message: string }

export class ApiError extends Error {
  readonly status: number
  readonly code: ApiErrorCode | undefined
  readonly body: unknown

  constructor(status: number, body: unknown) {
    const parsed = parseErrorBody(body)
    // The API's messages are written for users (Bahasa Indonesia), so they can be shown as is.
    super(parsed.message ?? `Permintaan gagal (${status})`)
    this.name = 'ApiError'
    this.status = status
    this.code = parsed.code
    this.body = body
  }
}

function parseErrorBody(body: unknown): { code?: ApiErrorCode; message?: string } {
  if (typeof body !== 'object' || body === null) return {}
  const { code, message } = body as Record<string, unknown>
  return {
    code: typeof code === 'string' ? (code as ApiErrorCode) : undefined,
    message: typeof message === 'string' ? sentenceCase(message) : undefined,
  }
}

function sentenceCase(text: string) {
  return text.charAt(0).toUpperCase() + text.slice(1)
}

// The session store registers these, so this module doesn't import the store (no cycle).
let getToken: () => string | null = () => null
let onUnauthenticated: () => void = () => {}

export function configureApiAuth(options: {
  getToken: () => string | null
  onUnauthenticated: () => void
}) {
  getToken = options.getToken
  onUnauthenticated = options.onUnauthenticated
}

// Private: requests go through apiGet / apiPost / apiPut so they all share the interceptors.
const http = axios.create({
  baseURL: env.apiBaseUrl,
  headers: { Accept: 'application/json' },
})

http.interceptors.request.use((config) => {
  const token = getToken()
  if (token) config.headers.set('Authorization', `Bearer ${token}`)
  return config
})

// Every failure becomes an ApiError, so callers handle one error type.
http.interceptors.response.use(undefined, (error: unknown) => {
  if (axios.isAxiosError(error) && error.response) {
    const apiError = new ApiError(error.response.status, error.response.data)
    // An expired or revoked session anywhere sends the user back to login.
    if (apiError.status === 401 && error.config?.headers.has('Authorization')) onUnauthenticated()
    return Promise.reject(apiError)
  }
  return Promise.reject(
    new ApiError(0, { message: 'tidak dapat terhubung ke server, periksa koneksi internet' }),
  )
})

/*
 * Request helpers. Each resolves to the response body (no axios wrapper) and rejects with ApiError.
 * Type the body as the API returns it, e.g. apiGet<DataResponse<User[]>>('/v1/users').
 */

export async function apiGet<T>(url: string, config?: AxiosRequestConfig) {
  return (await http.get<T>(url, config)).data
}

export async function apiPost<T, B = unknown>(url: string, body?: B, config?: AxiosRequestConfig) {
  return (await http.post<T>(url, body, config)).data
}

export async function apiPut<T, B = unknown>(url: string, body?: B, config?: AxiosRequestConfig) {
  return (await http.put<T>(url, body, config)).data
}
