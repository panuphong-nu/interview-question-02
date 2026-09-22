/**
 * The shapes this feature works in. They mirror the API contract and are the
 * only description of it in the app: a change to the wire format is corrected
 * here and in the api module, never in a component.
 */

export interface User {
  id: string
  username: string
  createdAt: string
}

export interface Session {
  accessToken: string
  tokenType: string
  expiresAt: string
  user: User
}

export interface RegisterInput {
  username: string
  password: string
  confirmPassword: string
}

export interface LoginInput {
  username: string
  password: string
}
