import { describe, expect, it } from 'vitest'

import {
  hasErrors,
  passwordMaxLength,
  passwordMinLength,
  usernameMaxLength,
  usernameMinLength,
  validateLogin,
  validateRegistration,
} from './credentials'

describe('validateRegistration', () => {
  it('accepts a valid sign up', () => {
    const errors = validateRegistration({
      username: 'Somchai_01',
      password: 'sup3r-secret',
      confirmPassword: 'sup3r-secret',
    })

    expect(hasErrors(errors)).toBe(false)
  })

  it('rejects a confirmation that does not match, on the field the form asks to fix', () => {
    const errors = validateRegistration({
      username: 'Somchai',
      password: 'sup3r-secret',
      confirmPassword: 'sup3r-secrets',
    })

    expect(errors['confirmPassword']).toHaveLength(1)
    expect(errors['password']).toBeUndefined()
  })

  it.each([
    ['empty', '', 'กรุณากรอกชื่อผู้ใช้งาน'],
    ['too short', 'ab', `อย่างน้อย ${usernameMinLength}`],
    ['too long', 'a'.repeat(usernameMaxLength + 1), `ไม่เกิน ${usernameMaxLength}`],
    ['containing a space', 'som chai', '. _ -'],
  ])('rejects a username that is %s', (_case, username, fragment) => {
    const errors = validateRegistration({
      username,
      password: 'sup3r-secret',
      confirmPassword: 'sup3r-secret',
    })

    expect(errors['username']?.[0]).toContain(fragment)
  })

  it('rejects a password shorter than the policy', () => {
    const errors = validateRegistration({
      username: 'Somchai',
      password: 'short',
      confirmPassword: 'short',
    })

    expect(errors['password']?.[0]).toContain(`อย่างน้อย ${passwordMinLength}`)
  })

  // The API measures the password in bytes, because that is the unit its
  // hashing algorithm limits. A Thai character is three of them.
  it('measures the password in bytes, as the API does', () => {
    const errors = validateRegistration({
      username: 'Somchai',
      // 25 characters, 75 bytes in UTF-8: within the character limit and past
      // the byte one.
      password: 'ก'.repeat(25),
      confirmPassword: 'ก'.repeat(25),
    })

    expect(errors['password']?.[0]).toContain(`ไม่เกิน ${passwordMaxLength}`)
  })

  it('reports every broken rule at once', () => {
    const errors = validateRegistration({ username: '', password: 'short', confirmPassword: '' })

    expect(Object.keys(errors).sort()).toEqual(['confirmPassword', 'password', 'username'])
  })
})

describe('validateLogin', () => {
  // Applying the sign up policy here would lock out an account created before
  // a rule was tightened.
  it('checks presence only', () => {
    expect(hasErrors(validateLogin({ username: 'ab', password: 'old' }))).toBe(false)
  })

  it('rejects empty fields', () => {
    const errors = validateLogin({ username: '   ', password: '' })

    expect(errors['username']).toHaveLength(1)
    expect(errors['password']).toHaveLength(1)
  })
})
