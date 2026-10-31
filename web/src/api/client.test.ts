import { type AxiosAdapter, AxiosError, type InternalAxiosRequestConfig } from 'axios'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { ApiError, apiGet, apiPost, apiPut, configureApiAuth } from './client'

/** Stub transport: the client's interceptors still run around it. */
function respond(
  status: number,
  data: unknown,
  seen?: (config: InternalAxiosRequestConfig) => void,
): AxiosAdapter {
  return async (config) => {
    seen?.(config)
    const response = { data, status, statusText: '', headers: {}, config }
    if (status >= 400) {
      throw new AxiosError('Request failed', 'ERR_BAD_RESPONSE', config, null, response)
    }
    return response
  }
}

describe('api client', () => {
  afterEach(() => {
    configureApiAuth({ getToken: () => null, onUnauthenticated: () => {} })
  })

  it('prefixes the API base URL and resolves to the response body', async () => {
    let url = ''
    const adapter = respond(200, { message: 'ok' }, (c) => {
      url = `${c.baseURL}${c.url}`
    })

    await expect(apiGet('/health', { adapter })).resolves.toEqual({ message: 'ok' })
    expect(url).toBe('/api/health')
  })

  it('sends bodies with apiPost and apiPut', async () => {
    const sent: unknown[] = []
    const record = (c: InternalAxiosRequestConfig) => sent.push([c.method, JSON.parse(c.data)])

    await apiPost('/v1/x', { name: 'A' }, { adapter: respond(201, { message: 'dibuat' }, record) })
    await apiPut('/v1/x/1', { name: 'B' }, { adapter: respond(200, { data: { id: 1 } }, record) })

    expect(sent).toEqual([
      ['post', { name: 'A' }],
      ['put', { name: 'B' }],
    ])
  })

  it('turns API errors into ApiError with code and user-facing message', async () => {
    const adapter = respond(409, { code: 'ALREADY_EXISTS', message: 'user sudah terdaftar' })

    const error = await apiGet('/v1/x', { adapter }).catch((e: unknown) => e)

    expect(error).toBeInstanceOf(ApiError)
    expect(error).toMatchObject({
      status: 409,
      code: 'ALREADY_EXISTS',
      message: 'User sudah terdaftar',
    })
  })

  it('sends the bearer token and reports an expired session', async () => {
    const onUnauthenticated = vi.fn()
    configureApiAuth({ getToken: () => 'abc', onUnauthenticated })
    let auth: unknown
    const adapter = respond(401, { code: 'UNAUTHENTICATED', message: 'sesi tidak valid' }, (c) => {
      auth = c.headers.get('Authorization')
    })

    await apiGet('/v1/users', { adapter }).catch(() => {})

    expect(auth).toBe('Bearer abc')
    expect(onUnauthenticated).toHaveBeenCalledOnce()
  })

  it('reports a network failure in plain words', async () => {
    const adapter: AxiosAdapter = async (config) => {
      throw new AxiosError('Network Error', 'ERR_NETWORK', config)
    }

    const error = await apiGet('/v1/x', { adapter }).catch((e: unknown) => e)

    expect(error).toMatchObject({
      status: 0,
      message: expect.stringMatching(/^Tidak dapat terhubung/),
    })
  })
})
