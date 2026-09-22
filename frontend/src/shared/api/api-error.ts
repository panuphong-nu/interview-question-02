/**
 * The error every API call rejects with, in one shape, so a component never
 * has to tell a network failure apart from an HTTP status by inspection.
 */
export class ApiError extends Error {
  /** HTTP status, or 0 when the request never reached the server. */
  readonly status: number

  /**
   * Messages keyed by the form field that caused them, as returned by the
   * API's validation responses. Empty for every other kind of failure.
   */
  readonly fieldErrors: Readonly<Record<string, string[]>>

  constructor(message: string, status: number, fieldErrors: Record<string, string[]> = {}) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.fieldErrors = fieldErrors
  }

  /** True when the credentials were refused or the token is no longer valid. */
  get isUnauthorized(): boolean {
    return this.status === 401
  }

  /** True when the request could not be sent or no response came back. */
  get isOffline(): boolean {
    return this.status === 0
  }
}
