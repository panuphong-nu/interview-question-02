import { ApiError } from './api-error'

/**
 * The single place that talks to the network.
 *
 * Everything above it works with plain values and ApiError, so the feature
 * layer holds no knowledge of fetch, status codes or response shapes, and can
 * be tested without a server.
 */

/** Wording for the failures that never reach the server, so they read the same as the API's own. */
const messageOffline = 'ไม่สามารถเชื่อมต่อเซิร์ฟเวอร์ได้ กรุณาลองใหม่'
const messageUnexpected = 'เกิดข้อผิดพลาด กรุณาลองใหม่'

/** How long a request may take before it is abandoned. */
const requestTimeoutMs = 15_000

/** Shape of the API's single line error body. */
interface MessageBody {
  message?: string
}

/** Shape of the API's validation error body. */
interface ValidationBody {
  title?: string
  errors?: Record<string, string[]>
}

export interface RequestOptions {
  method?: 'GET' | 'POST'
  body?: unknown
  /** Bearer token to send, when the endpoint requires one. */
  token?: string | undefined
  signal?: AbortSignal | undefined
}

export class HttpClient {
  private readonly baseUrl: string

  constructor(baseUrl: string) {
    // A trailing slash would produce '//api/...' once a path is appended.
    this.baseUrl = baseUrl.replace(/\/+$/, '')
  }

  async request<TResponse>(path: string, options: RequestOptions = {}): Promise<TResponse> {
    const { method = 'GET', body, token, signal } = options

    const headers = new Headers({ Accept: 'application/json' })
    if (body !== undefined) {
      headers.set('Content-Type', 'application/json')
    }
    if (token) {
      headers.set('Authorization', `Bearer ${token}`)
    }

    // A request that hangs would leave a form disabled with no way out, so
    // every call carries its own deadline alongside any caller's signal.
    const timeout = AbortSignal.timeout(requestTimeoutMs)
    const abort = signal ? AbortSignal.any([signal, timeout]) : timeout

    let response: Response
    try {
      response = await fetch(`${this.baseUrl}${path}`, {
        method,
        headers,
        signal: abort,
        ...(body === undefined ? {} : { body: JSON.stringify(body) }),
      })
    } catch (cause) {
      // fetch rejects only when the request could not be completed at all:
      // no network, DNS failure, CORS refusal, or the timeout above.
      throw new ApiError(messageOffline, 0, {})
    }

    if (!response.ok) {
      throw await this.toApiError(response)
    }
    return (await this.readJson(response)) as TResponse
  }

  /** Reads the body of a failed response into the one error type the app uses. */
  private async toApiError(response: Response): Promise<ApiError> {
    const payload = await this.readJson(response)

    if (isValidationBody(payload)) {
      const messages = Object.values(payload.errors).flat()
      return new ApiError(messages[0] ?? messageUnexpected, response.status, payload.errors)
    }
    if (isMessageBody(payload)) {
      return new ApiError(payload.message, response.status, {})
    }
    // A body that is not JSON, or not a shape this app knows: the status is
    // still meaningful, so it is kept and the wording falls back.
    return new ApiError(messageUnexpected, response.status, {})
  }

  /** Parses a JSON body, returning undefined rather than throwing on an empty or invalid one. */
  private async readJson(response: Response): Promise<unknown> {
    const text = await response.text()
    if (text === '') {
      return undefined
    }
    try {
      return JSON.parse(text) as unknown
    } catch {
      return undefined
    }
  }
}

function isMessageBody(payload: unknown): payload is Required<MessageBody> {
  return (
    typeof payload === 'object' &&
    payload !== null &&
    typeof (payload as MessageBody).message === 'string'
  )
}

function isValidationBody(payload: unknown): payload is ValidationBody & {
  errors: Record<string, string[]>
} {
  if (typeof payload !== 'object' || payload === null) {
    return false
  }
  const errors = (payload as ValidationBody).errors
  return typeof errors === 'object' && errors !== null && Object.keys(errors).length > 0
}
