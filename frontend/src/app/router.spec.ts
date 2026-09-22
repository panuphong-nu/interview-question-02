import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createMemoryHistory } from 'vue-router'
import type { Router } from 'vue-router'

import { ApiError } from '@/shared/api'
import { authApi } from '@/features/auth/api/auth.api'
import { tokenStorage } from '@/features/auth/model/token-storage'

import { createAppRouter } from './router'

vi.mock('@/features/auth/api/auth.api', () => ({
  authApi: { register: vi.fn(), login: vi.fn(), me: vi.fn() },
}))

const user = { id: 'user-1', username: 'Somchai', createdAt: '2026-09-21T10:00:00Z' }

let router: Router

beforeEach(() => {
  setActivePinia(createPinia())
  vi.mocked(authApi.me).mockReset()
  window.sessionStorage.clear()

  // A router of its own, over an in memory history, so no test inherits where
  // the previous one navigated to. The first push in each test is also its
  // initial navigation, which is what an in memory history waits for.
  router = createAppRouter(createMemoryHistory())
})

describe('route guards', () => {
  it('sends a visitor with no session to the sign in screen', async () => {
    await router.push('/welcome')

    expect(router.currentRoute.value.name).toBe('login')
    // Where they were going is remembered, so signing in continues the journey.
    expect(router.currentRoute.value.query['redirect']).toBe('/welcome')
  })

  // After a reload the token is present but unproven, so the guard asks the
  // API rather than trusting what is in storage.
  it('lets a stored token through once the API confirms it', async () => {
    tokenStorage.write('a-token')
    vi.mocked(authApi.me).mockResolvedValue(user)

    await router.push('/welcome')

    expect(authApi.me).toHaveBeenCalledWith('a-token')
    expect(router.currentRoute.value.name).toBe('welcome')
  })

  it('turns a refused token away', async () => {
    tokenStorage.write('an-expired-token')
    vi.mocked(authApi.me).mockRejectedValue(new ApiError('กรุณาลงชื่อเข้าใช้งาน', 401))

    await router.push('/welcome')

    expect(router.currentRoute.value.name).toBe('login')
    expect(tokenStorage.read()).toBeNull()
  })

  // Someone already signed in has no use for the two forms.
  it.each(['/login', '/register'])('redirects a signed in person away from %s', async (path) => {
    tokenStorage.write('a-token')
    vi.mocked(authApi.me).mockResolvedValue(user)

    await router.push(path)

    expect(router.currentRoute.value.name).toBe('welcome')
  })

  it('opens on the sign in screen', async () => {
    await router.push('/')

    expect(router.currentRoute.value.name).toBe('login')
  })

  it('sends an unknown address to the sign in screen', async () => {
    await router.push('/no-such-page')

    expect(router.currentRoute.value.name).toBe('login')
  })
})
