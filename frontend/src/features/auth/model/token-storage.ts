/**
 * Where the access token is kept between page loads.
 *
 * sessionStorage rather than localStorage: the token is cleared when the tab
 * closes, which keeps a session from outliving its use on a shared computer.
 * Neither is immune to a script injected into this origin; the durable defence
 * against that is the token's short lifetime and the API refusing an expired
 * one, not the storage choice.
 */

const storageKey = 'it02.accessToken'

export const tokenStorage = {
  read(): string | null {
    // Storage throws in a browser configured to block it, and returning null
    // simply means the person signs in again.
    try {
      return window.sessionStorage.getItem(storageKey)
    } catch {
      return null
    }
  },

  write(token: string): void {
    try {
      window.sessionStorage.setItem(storageKey, token)
    } catch {
      // A session that lives only in memory still works for this tab.
    }
  },

  clear(): void {
    try {
      window.sessionStorage.removeItem(storageKey)
    } catch {
      // Nothing to do: there is no stored token to remove.
    }
  },
}

export type TokenStorage = typeof tokenStorage
