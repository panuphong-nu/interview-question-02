import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

import { ApiError } from '@/shared/api'

import { authApi } from '../api/auth.api'
import { tokenStorage } from '../model/token-storage'
import type { LoginInput, RegisterInput, User } from '../model/types'

/**
 * The session, held in one place.
 *
 * Pages read `user` and call these actions; none of them touches storage or
 * the API directly, which is what keeps "who is signed in" a single answer
 * across every route.
 */
export const useAuthStore = defineStore('auth', () => {
  // Starts empty even when storage holds a token: an unverified token is not
  // a session, and `restore` is the one place that reads storage and asks the
  // API whether what it found is still good.
  const token = ref<string | null>(null)
  const user = ref<User | null>(null)
  /** True while a stored token is being checked against the API. */
  const restoring = ref(false)

  const isAuthenticated = computed(() => token.value !== null && user.value !== null)

  /**
   * Creates an account. No session is started: the assignment sends the person
   * back to the sign in screen, which is also the safer default, as it proves
   * the password works before it is the only way back in.
   */
  async function register(input: RegisterInput): Promise<User> {
    return authApi.register(input)
  }

  async function login(input: LoginInput): Promise<User> {
    const session = await authApi.login(input)

    token.value = session.accessToken
    user.value = session.user
    tokenStorage.write(session.accessToken)
    return session.user
  }

  /**
   * Re-establishes a session from a stored token after a reload, by asking the
   * API who the token belongs to. The answer is what confirms the token is
   * still valid: the app never inspects the token itself, because only the
   * server can check a signature.
   *
   * Returns true when a session is in place afterwards.
   */
  async function restore(): Promise<boolean> {
    if (user.value !== null) {
      return true
    }

    const stored = token.value ?? tokenStorage.read()
    if (stored === null) {
      return false
    }

    restoring.value = true
    try {
      user.value = await authApi.me(stored)
      token.value = stored
      return true
    } catch (error) {
      // A refused token is a finished session. Any other failure, an outage
      // for instance, also leaves this app unable to prove who the person is,
      // so both end the same way: signed out, with a clean slate.
      if (!(error instanceof ApiError)) {
        throw error
      }
      signOut()
      return false
    } finally {
      restoring.value = false
    }
  }

  function signOut(): void {
    token.value = null
    user.value = null
    tokenStorage.clear()
  }

  return { token, user, restoring, isAuthenticated, register, login, restore, signOut }
})
