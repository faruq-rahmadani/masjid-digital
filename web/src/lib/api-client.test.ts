import { afterEach, describe, expect, it, vi } from 'vitest'
import { api, ApiError } from './api-client'

function mockFetch(response: Response) {
  const fetchMock = vi.fn().mockResolvedValue(response)
  vi.stubGlobal('fetch', fetchMock)
  return fetchMock
}

const json = (body: unknown, status = 200) =>
  new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } })

describe('api client', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('prefixes the base URL and parses JSON', async () => {
    const fetchMock = mockFetch(json({ message: 'ok' }))

    await expect(api.get('/health')).resolves.toEqual({ message: 'ok' })
    expect(fetchMock).toHaveBeenCalledWith(
      '/api/health',
      expect.objectContaining({ method: 'GET' }),
    )
  })

  it('serializes the body and sets Content-Type', async () => {
    const fetchMock = mockFetch(json({ id: '1' }, 201))

    await api.post('/announcements', { title: 'Kajian' })

    const [, init] = fetchMock.mock.calls[0] as [string, RequestInit]
    expect(init.body).toBe('{"title":"Kajian"}')
    expect(init.headers).toMatchObject({ 'Content-Type': 'application/json' })
  })

  it('throws ApiError with status and body on non-2xx', async () => {
    mockFetch(json({ message: 'not found' }, 404))

    const error = await api.get('/missing').catch((e: unknown) => e)
    expect(error).toBeInstanceOf(ApiError)
    expect(error).toMatchObject({ status: 404, body: { message: 'not found' } })
  })
})
