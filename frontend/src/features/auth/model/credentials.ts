/**
 * The same rules the API enforces, applied in the browser so a person is told
 * about a mistake as they type instead of after a round trip. The server
 * repeats every one of them: this copy is a convenience, never the guarantee.
 */

export const usernameMinLength = 3
export const usernameMaxLength = 32
export const passwordMinLength = 8
export const passwordMaxLength = 72

/** Letters, digits and the separators . _ - only. */
const usernamePattern = /^[\p{L}\p{N}._-]+$/u

/** Messages keyed by field, in the same shape the API's validation body uses. */
export type FieldErrors = Record<string, string[]>

export function validateRegistration(input: {
  username: string
  password: string
  confirmPassword: string
}): FieldErrors {
  const errors: FieldErrors = {}

  const username = validateUsername(input.username)
  if (username) {
    errors['username'] = [username]
  }

  const password = validatePassword(input.password)
  if (password) {
    errors['password'] = [password]
  }

  if (input.confirmPassword === '') {
    errors['confirmPassword'] = ['กรุณากรอกยืนยันรหัสผ่าน']
  } else if (input.password !== input.confirmPassword) {
    errors['confirmPassword'] = ['รหัสผ่านและยืนยันรหัสผ่านไม่ตรงกัน']
  }

  return errors
}

/**
 * Sign in checks presence only, matching the API: applying the current policy
 * to an existing password would lock out an account created under an older one.
 */
export function validateLogin(input: { username: string; password: string }): FieldErrors {
  const errors: FieldErrors = {}

  if (input.username.trim() === '') {
    errors['username'] = ['กรุณากรอกชื่อผู้ใช้งาน']
  }
  if (input.password === '') {
    errors['password'] = ['กรุณากรอกรหัสผ่าน']
  }
  return errors
}

export function hasErrors(errors: FieldErrors): boolean {
  return Object.keys(errors).length > 0
}

function validateUsername(raw: string): string | null {
  const username = raw.trim()

  if (username === '') {
    return 'กรุณากรอกชื่อผู้ใช้งาน'
  }
  if ([...username].length < usernameMinLength) {
    return `ชื่อผู้ใช้งานต้องมีอย่างน้อย ${usernameMinLength} ตัวอักษร`
  }
  if ([...username].length > usernameMaxLength) {
    return `ชื่อผู้ใช้งานต้องไม่เกิน ${usernameMaxLength} ตัวอักษร`
  }
  if (!usernamePattern.test(username)) {
    return 'ชื่อผู้ใช้งานใช้ได้เฉพาะตัวอักษร ตัวเลข และ . _ - เท่านั้น'
  }
  return null
}

function validatePassword(password: string): string | null {
  if (password === '') {
    return 'กรุณากรอกรหัสผ่าน'
  }
  // Measured in bytes, as the API does: its limit comes from what the hashing
  // algorithm accepts, and a Thai character is three bytes in UTF-8.
  const length = new TextEncoder().encode(password).length

  if (length < passwordMinLength) {
    return `รหัสผ่านต้องมีอย่างน้อย ${passwordMinLength} ตัวอักษร`
  }
  if (length > passwordMaxLength) {
    return `รหัสผ่านต้องไม่เกิน ${passwordMaxLength} ตัวอักษร`
  }
  return null
}
