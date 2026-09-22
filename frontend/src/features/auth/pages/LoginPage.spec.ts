import { render, screen, waitFor } from '@testing-library/vue'
import userEvent from '@testing-library/user-event'
import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { ApiError } from '@/shared/api'

import { authApi } from '../api/auth.api'
import type { Session } from '../model/types'
import LoginPage from './LoginPage.vue'

vi.mock('../api/auth.api', () => ({
  authApi: { register: vi.fn(), login: vi.fn(), me: vi.fn() },
}))

const replace = vi.fn()
let query: Record<string, string> = {}

vi.mock('vue-router', () => ({
  useRouter: () => ({ replace }),
  useRoute: () => ({ query }),
  RouterLink: { template: '<a><slot /></a>' },
}))

const session: Session = {
  accessToken: 'a-token',
  tokenType: 'Bearer',
  expiresAt: '2026-09-21T11:00:00Z',
  user: { id: 'user-1', username: 'Somchai', createdAt: '2026-09-21T10:00:00Z' },
}

beforeEach(() => {
  setActivePinia(createPinia())
  vi.mocked(authApi.login).mockReset()
  replace.mockReset()
  query = {}
  window.sessionStorage.clear()
})

function renderPage() {
  return render(LoginPage, {
    global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
  })
}

describe('IT 02-1 · LoginPage', () => {
  it('shows the two fields, the button and the way to sign up', () => {
    renderPage()

    expect(screen.getByLabelText('User')).toBeInTheDocument()
    expect(screen.getByLabelText('Password')).toHaveAttribute('type', 'password')
    expect(screen.getByRole('button', { name: 'ลงชื่อเข้าใช้งาน' })).toBeInTheDocument()
    expect(screen.getByText('สมัครสมาชิก')).toBeInTheDocument()
  })

  it('signs in and moves on to the welcome screen', async () => {
    vi.mocked(authApi.login).mockResolvedValue(session)
    renderPage()

    await userEvent.type(screen.getByLabelText('User'), 'Somchai')
    await userEvent.type(screen.getByLabelText('Password'), 'sup3r-secret')
    await userEvent.click(screen.getByRole('button', { name: 'ลงชื่อเข้าใช้งาน' }))

    await waitFor(() => {
      expect(authApi.login).toHaveBeenCalledWith({
        username: 'Somchai',
        password: 'sup3r-secret',
      })
    })
    expect(replace).toHaveBeenCalledWith({ name: 'welcome' })
  })

  // The guard records where the person was going, and sign in has to honour it.
  it('returns to the page the guard interrupted', async () => {
    query = { redirect: '/welcome' }
    vi.mocked(authApi.login).mockResolvedValue(session)
    renderPage()

    await userEvent.type(screen.getByLabelText('User'), 'Somchai')
    await userEvent.type(screen.getByLabelText('Password'), 'sup3r-secret')
    await userEvent.click(screen.getByRole('button', { name: 'ลงชื่อเข้าใช้งาน' }))

    await waitFor(() => expect(replace).toHaveBeenCalledWith('/welcome'))
  })

  it('reports refused credentials and clears only the password', async () => {
    vi.mocked(authApi.login).mockRejectedValue(
      new ApiError('ชื่อผู้ใช้งานหรือรหัสผ่านไม่ถูกต้อง', 401),
    )
    renderPage()

    await userEvent.type(screen.getByLabelText('User'), 'Somchai')
    await userEvent.type(screen.getByLabelText('Password'), 'wrong-password')
    await userEvent.click(screen.getByRole('button', { name: 'ลงชื่อเข้าใช้งาน' }))

    expect(await screen.findByRole('alert')).toHaveTextContent('ชื่อผู้ใช้งานหรือรหัสผ่านไม่ถูกต้อง')
    expect(replace).not.toHaveBeenCalled()
    expect(screen.getByLabelText('User')).toHaveValue('Somchai')
    expect(screen.getByLabelText('Password')).toHaveValue('')
  })

  it('does not call the API with an empty form', async () => {
    renderPage()

    await userEvent.click(screen.getByRole('button', { name: 'ลงชื่อเข้าใช้งาน' }))

    expect(await screen.findByText('กรุณากรอกชื่อผู้ใช้งาน')).toBeInTheDocument()
    expect(screen.getByText('กรุณากรอกรหัสผ่าน')).toBeInTheDocument()
    expect(authApi.login).not.toHaveBeenCalled()
  })

  it('tells the person a server it could not reach is the problem', async () => {
    vi.mocked(authApi.login).mockRejectedValue(
      new ApiError('ไม่สามารถเชื่อมต่อเซิร์ฟเวอร์ได้ กรุณาลองใหม่', 0),
    )
    renderPage()

    await userEvent.type(screen.getByLabelText('User'), 'Somchai')
    await userEvent.type(screen.getByLabelText('Password'), 'sup3r-secret')
    await userEvent.click(screen.getByRole('button', { name: 'ลงชื่อเข้าใช้งาน' }))

    expect(await screen.findByRole('alert')).toHaveTextContent('ไม่สามารถเชื่อมต่อเซิร์ฟเวอร์ได้')
  })

  // RegisterPage arrives here with this flag instead of a started session.
  it('confirms a completed registration', () => {
    query = { registered: '1' }
    renderPage()

    expect(screen.getByRole('alert')).toHaveTextContent('สมัครสมาชิกเรียบร้อยแล้ว')
  })
})
