import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { ApiError } from './api-error'
import { HttpClient } from './http-client'

/**
 * fetch is replaced rather than a server started: these tests are about how a
 * response becomes a value or an ApiError, which is all the layers above ever
 * see.
 */
const fetchMock = vi.fn()

/**
 * Runs a request that is expected to fail and returns the ApiError it
 * rejected with, so each test can assert on a typed value rather than the
 * `unknown` a bare catch produces.
 */
async function captureError(request: Promise<unknown>): Promise<ApiError> {
  try {
    await request
  } catch (error) {
    if (error instanceof ApiError) {
      return error
    }
    throw error
  }
  throw new Error('expected the request to fail, but it resolved')
}

function jsonResponse(status: number, body: unknown): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  })
}

beforeEach(() => {
  vi.stubGlobal('fetch', fetchMock)
  fetchMock.mockReset()
})

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('HttpClient', () => {
  const client = new HttpClient('http://api.test')

  it('returns the decoded body of a successful response', async () => {
    fetchMock.mockResolvedValue(jsonResponse(200, { id: 'user-1' }))

    await expect(client.request('/api/auth/me')).resolves.toEqual({ id: 'user-1' })
  })

  it('sends a JSON body and the bearer token', async () => {
    fetchMock.mockResolvedValue(jsonResponse(200, {}))

    await client.request('/api/auth/login', {
      method: 'POST',
      body: { username: 'somchai' },
      token: 'a-token',
    })

    const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit]
    expect(url).toBe('http://api.test/api/auth/login')
    expect(init.method).toBe('POST')
    expect(init.body).toBe('{"username":"somchai"}')

    const headers = init.headers as Headers
    expect(headers.get('Authorization')).toBe('Bearer a-token')
    expect(headers.get('Content-Type')).toBe('application/json')
  })

  it('omits the Authorization header when there is no token', async () => {
    fetchMock.mockResolvedValue(jsonResponse(200, {}))

    await client.request('/api/auth/login', { method: 'POST', body: {} })

    const [, init] = fetchMock.mock.calls[0] as [string, RequestInit]
    expect((init.headers as Headers).has('Authorization')).toBe(false)
  })

  it('keeps the field messages of a validation response', async () => {
    fetchMock.mockResolvedValue(
      jsonResponse(400, {
        title: 'One or more validation errors occurred.',
        status: 400,
        errors: { confirmPassword: ['รหัสผ่านและยืนยันรหัสผ่านไม่ตรงกัน'] },
      }),
    )

    const error = await captureError(client.request('/api/auth/register', { method: 'POST' }))

    expect(error.status).toBe(400)
    expect(error.fieldErrors['confirmPassword']).toEqual(['รหัสผ่านและยืนยันรหัสผ่านไม่ตรงกัน'])
    // The first field message doubles as the summary, so a form with no
    // per-field display still says something useful.
    expect(error.message).toBe('รหัสผ่านและยืนยันรหัสผ่านไม่ตรงกัน')
  })

  it('carries the API wording of a single line error', async () => {
    fetchMock.mockResolvedValue(jsonResponse(401, { message: 'ชื่อผู้ใช้งานหรือรหัสผ่านไม่ถูกต้อง' }))

    const error = await captureError(client.request('/api/auth/login', { method: 'POST' }))

    expect(error.isUnauthorized).toBe(true)
    expect(error.message).toBe('ชื่อผู้ใช้งานหรือรหัสผ่านไม่ถูกต้อง')
  })

  it('keeps the status when the body is not a shape it knows', async () => {
    fetchMock.mockResolvedValue(new Response('<html>502</html>', { status: 502 }))

    const error = await captureError(client.request('/api/auth/me'))

    expect(error.status).toBe(502)
    expect(error.message).not.toBe('')
  })

  // A failed fetch means no response at all: no network, a CORS refusal, or
  // the client's own timeout. It must not surface as a raw TypeError.
  it('turns an unreachable server into an offline ApiError', async () => {
    fetchMock.mockRejectedValue(new TypeError('Failed to fetch'))

    const error = await captureError(client.request('/api/auth/me'))

    expect(error.isOffline).toBe(true)
    expect(error.status).toBe(0)
  })

  it('does not double the slash between the base URL and the path', async () => {
    fetchMock.mockResolvedValue(jsonResponse(200, {}))

    await new HttpClient('http://api.test/').request('/health')

    expect(fetchMock.mock.calls[0]?.[0]).toBe('http://api.test/health')
  })
})
