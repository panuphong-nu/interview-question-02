import { render, screen } from '@testing-library/vue'
import userEvent from '@testing-library/user-event'
import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { authApi } from '../api/auth.api'
import { tokenStorage } from '../model/token-storage'
import { useAuthStore } from '../stores/auth.store'
import WelcomePage from './WelcomePage.vue'

vi.mock('../api/auth.api', () => ({
  authApi: { register: vi.fn(), login: vi.fn(), me: vi.fn() },
}))

const replace = vi.fn()
vi.mock('vue-router', () => ({
  useRouter: () => ({ replace }),
  useRoute: () => ({ query: {} }),
}))

beforeEach(() => {
  setActivePinia(createPinia())
  replace.mockReset()
  window.sessionStorage.clear()
})

async function signIn() {
  vi.mocked(authApi.login).mockResolvedValue({
    accessToken: 'a-token',
    tokenType: 'Bearer',
    expiresAt: '2026-09-21T11:00:00Z',
    user: { id: 'user-1', username: 'Somchai', createdAt: '2026-09-21T10:00:00Z' },
  })
  const auth = useAuthStore()
  await auth.login({ username: 'Somchai', password: 'sup3r-secret' })
  return auth
}

describe('IT 02-3 · WelcomePage', () => {
  it('greets the signed in user by name', async () => {
    await signIn()

    render(WelcomePage)

    expect(screen.getByText('Welcome User : Somchai')).toBeInTheDocument()
  })

  it('signs out and returns to the sign in screen', async () => {
    const auth = await signIn()
    render(WelcomePage)

    await userEvent.click(screen.getByRole('button', { name: 'ออกจากระบบ' }))

    expect(auth.isAuthenticated).toBe(false)
    expect(tokenStorage.read()).toBeNull()
    expect(replace).toHaveBeenCalledWith({ name: 'login' })
  })
})
