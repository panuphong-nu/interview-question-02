import { render, screen, waitFor } from '@testing-library/vue'
import userEvent from '@testing-library/user-event'
import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { ApiError } from '@/shared/api'

import { authApi } from '../api/auth.api'
import RegisterPage from './RegisterPage.vue'

vi.mock('../api/auth.api', () => ({
  authApi: { register: vi.fn(), login: vi.fn(), me: vi.fn() },
}))

const replace = vi.fn()
vi.mock('vue-router', () => ({
  useRouter: () => ({ replace }),
  useRoute: () => ({ query: {} }),
  RouterLink: { template: '<a><slot /></a>' },
}))

beforeEach(() => {
  setActivePinia(createPinia())
  vi.mocked(authApi.register).mockReset()
  replace.mockReset()
})

function renderPage() {
  return render(RegisterPage, {
    global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
  })
}

describe('IT 02-2 · RegisterPage', () => {
  it('shows the three fields the form asks for', () => {
    renderPage()

    expect(screen.getByLabelText('User')).toBeInTheDocument()
    expect(screen.getByLabelText('Password')).toBeInTheDocument()
    expect(screen.getByLabelText('Confirm Password')).toBeInTheDocument()
  })

  // The assignment asks for the typed password to be masked.
  it('masks both password fields', () => {
    renderPage()

    expect(screen.getByLabelText('Password')).toHaveAttribute('type', 'password')
    expect(screen.getByLabelText('Confirm Password')).toHaveAttribute('type', 'password')
  })

  it('registers and returns to the sign in screen', async () => {
    vi.mocked(authApi.register).mockResolvedValue({
      id: 'user-1',
      username: 'Somchai',
      createdAt: '2026-09-21T10:00:00Z',
    })
    renderPage()

    await userEvent.type(screen.getByLabelText('User'), 'Somchai')
    await userEvent.type(screen.getByLabelText('Password'), 'sup3r-secret')
    await userEvent.type(screen.getByLabelText('Confirm Password'), 'sup3r-secret')
    await userEvent.click(screen.getByRole('button', { name: 'สมัครสมาชิก' }))

    await waitFor(() => {
      expect(authApi.register).toHaveBeenCalledWith({
        username: 'Somchai',
        password: 'sup3r-secret',
        confirmPassword: 'sup3r-secret',
      })
    })
    // Back to IT 02-1, with the flag that confirms what just happened.
    expect(replace).toHaveBeenCalledWith({ name: 'login', query: { registered: '1' } })
  })

  it('refuses a mismatched confirmation without calling the API', async () => {
    renderPage()

    await userEvent.type(screen.getByLabelText('User'), 'Somchai')
    await userEvent.type(screen.getByLabelText('Password'), 'sup3r-secret')
    await userEvent.type(screen.getByLabelText('Confirm Password'), 'sup3r-secrets')
    await userEvent.click(screen.getByRole('button', { name: 'สมัครสมาชิก' }))

    expect(await screen.findByText('รหัสผ่านและยืนยันรหัสผ่านไม่ตรงกัน')).toBeInTheDocument()
    expect(authApi.register).not.toHaveBeenCalled()
    expect(replace).not.toHaveBeenCalled()
  })

  it('shows the API message when the username is taken', async () => {
    vi.mocked(authApi.register).mockRejectedValue(
      new ApiError('ชื่อผู้ใช้งานนี้ถูกใช้แล้ว', 409),
    )
    renderPage()

    await userEvent.type(screen.getByLabelText('User'), 'Somchai')
    await userEvent.type(screen.getByLabelText('Password'), 'sup3r-secret')
    await userEvent.type(screen.getByLabelText('Confirm Password'), 'sup3r-secret')
    await userEvent.click(screen.getByRole('button', { name: 'สมัครสมาชิก' }))

    expect(await screen.findByRole('alert')).toHaveTextContent('ชื่อผู้ใช้งานนี้ถูกใช้แล้ว')
    expect(replace).not.toHaveBeenCalled()
    // The username survives so it can be edited; the secrets do not.
    expect(screen.getByLabelText('User')).toHaveValue('Somchai')
    expect(screen.getByLabelText('Password')).toHaveValue('')
    expect(screen.getByLabelText('Confirm Password')).toHaveValue('')
  })

  // A double click on a slow connection must not create two accounts.
  it('disables the button while the request is in flight', async () => {
    let settle: (value: { id: string; username: string; createdAt: string }) => void = () => {}
    vi.mocked(authApi.register).mockReturnValue(
      new Promise((resolve) => {
        settle = resolve
      }),
    )
    renderPage()

    await userEvent.type(screen.getByLabelText('User'), 'Somchai')
    await userEvent.type(screen.getByLabelText('Password'), 'sup3r-secret')
    await userEvent.type(screen.getByLabelText('Confirm Password'), 'sup3r-secret')
    await userEvent.click(screen.getByRole('button', { name: 'สมัครสมาชิก' }))

    const button = await screen.findByRole('button', { name: 'กำลังสมัคร…' })
    expect(button).toBeDisabled()

    await userEvent.click(button)
    expect(authApi.register).toHaveBeenCalledTimes(1)

    settle({ id: 'user-1', username: 'Somchai', createdAt: '2026-09-21T10:00:00Z' })
  })
})
