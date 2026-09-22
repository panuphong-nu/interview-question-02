import { httpClient } from '@/shared/api'

import type { LoginInput, RegisterInput, Session, User } from '../model/types'

/**
 * One function per endpoint. This is the only module that knows the API's
 * paths and verbs, so the store above it reads as use cases rather than URLs.
 */
export const authApi = {
  register(input: RegisterInput): Promise<User> {
    return httpClient.request<User>('/api/auth/register', { method: 'POST', body: input })
  },

  login(input: LoginInput): Promise<Session> {
    return httpClient.request<Session>('/api/auth/login', { method: 'POST', body: input })
  },

  /** Resolves the account behind a token, which is also how the token is checked. */
  me(token: string): Promise<User> {
    return httpClient.request<User>('/api/auth/me', { token })
  },
}

export type AuthApi = typeof authApi
