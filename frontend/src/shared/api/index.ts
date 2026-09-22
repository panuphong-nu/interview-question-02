import { HttpClient } from './http-client'

/**
 * The API's address. It is read from the build time environment so the same
 * source produces a local build and a deployed one; the default matches the
 * port the Go server listens on out of the box.
 */
const baseUrl = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:4201'

/** The client every feature's API module uses. */
export const httpClient = new HttpClient(baseUrl)

export { ApiError } from './api-error'
export { HttpClient } from './http-client'
export type { RequestOptions } from './http-client'
