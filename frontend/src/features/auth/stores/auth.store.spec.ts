import { setActivePinia, createPinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { ApiError } from '@/shared/api'

import { authApi } from '../api/auth.api'
import { tokenStorage } from '../model/token-storage'
import type { Session, User } from '../model/types'
import { useAuthStore } from './auth.store'

vi.mock('../api/auth.api', () => ({
  authApi: { register: vi.fn(), login: vi.fn(), me: vi.fn() },
}))

const user: User = { id: 'user-1', username: 'Somchai', createdAt: '2026-09-21T10:00:00Z' }

const session: Session = {
  accessToken: 'a-token',
  tokenType: 'Bearer',
  expiresAt: '2026-09-21T11:00:00Z',
  user,
}

beforeEach(() => {
  setActivePinia(createPinia())
  vi.mocked(authApi.login).mockReset()
  vi.mocked(authApi.me).mockReset()
  vi.mocked(authApi.register).mockReset()
  window.sessionStorage.clear()
})

describe('useAuthStore', () => {
  it('starts signed out', () => {
    const auth = useAuthStore()

    expect(auth.isAuthenticated).toBe(false)
    expect(auth.user).toBeNull()
  })

  it('keeps the session and persists the token after signing in', async () => {
    vi.mocked(authApi.login).mockResolvedValue(session)
    const auth = useAuthStore()

    await auth.login({ username: 'somchai', password: 'sup3r-secret' })

    expect(auth.isAuthenticated).toBe(true)
    expect(auth.user?.username).toBe('Somchai')
    // Persisted, so a reload does not force another sign in.
    expect(tokenStorage.read()).toBe('a-token')
  })

  // The assignment sends the person back to IT 02-1 to sign in, so creating an
  // account must not start a session.
  it('does not sign the person in when they register', async () => {
    vi.mocked(authApi.register).mockResolvedValue(user)
    const auth = useAuthStore()

    await auth.register({
      username: 'Somchai',
      password: 'sup3r-secret',
      confirmPassword: 'sup3r-secret',
    })

    expect(auth.isAuthenticated).toBe(false)
    expect(tokenStorage.read()).toBeNull()
  })

  it('forgets everything when signing out', async () => {
    vi.mocked(authApi.login).mockResolvedValue(session)
    const auth = useAuthStore()
    await auth.login({ username: 'somchai', password: 'sup3r-secret' })

    auth.signOut()

    expect(auth.isAuthenticated).toBe(false)
    expect(auth.token).toBeNull()
    expect(tokenStorage.read()).toBeNull()
  })

  describe('restore', () => {
    // The token is proof only if the server says so, which is why a reload
    // asks rather than reading the payload.
    it('asks the API who a stored token belongs to', async () => {
      tokenStorage.write('a-token')
      vi.mocked(authApi.me).mockResolvedValue(user)
      const auth = useAuthStore()

      await expect(auth.restore()).resolves.toBe(true)

      expect(authApi.me).toHaveBeenCalledWith('a-token')
      expect(auth.user?.username).toBe('Somchai')
    })

    it('signs out when the stored token is refused', async () => {
      tokenStorage.write('an-expired-token')
      vi.mocked(authApi.me).mockRejectedValue(new ApiError('กรุณาลงชื่อเข้าใช้งาน', 401))
      const auth = useAuthStore()

      await expect(auth.restore()).resolves.toBe(false)

      expect(auth.isAuthenticated).toBe(false)
      // The dead token is cleared, so the next reload does not retry it.
      expect(tokenStorage.read()).toBeNull()
    })

    it('reports no session when there is no token', async () => {
      const auth = useAuthStore()

      await expect(auth.restore()).resolves.toBe(false)
      expect(authApi.me).not.toHaveBeenCalled()
    })

    // Two guarded navigations in a row must not produce two requests.
    it('does not ask again once the session is known', async () => {
      tokenStorage.write('a-token')
      vi.mocked(authApi.me).mockResolvedValue(user)
      const auth = useAuthStore()

      await auth.restore()
      await auth.restore()

      expect(authApi.me).toHaveBeenCalledTimes(1)
    })
  })
})
