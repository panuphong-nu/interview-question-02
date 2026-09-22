<script setup lang="ts">
import { useRouter } from 'vue-router'

import FormPanel from '@/shared/ui/FormPanel.vue'

import { useAuthStore } from '../stores/auth.store'

/**
 * IT 02-3 — the screen behind the token.
 *
 * The name comes from the store, which got it from the API's /me endpoint
 * rather than from the token's payload: a JWT is signed but not secret, and
 * only the server can tell whether the signature is good.
 */

const auth = useAuthStore()
const router = useRouter()

async function signOut(): Promise<void> {
  auth.signOut()
  await router.replace({ name: 'login' })
}
</script>

<template>
  <FormPanel title="IT 02-3">
    <div class="welcome">
      <p class="welcome__greeting">Welcome User : {{ auth.user?.username }}</p>
      <button class="welcome__sign-out" type="button" @click="signOut">ออกจากระบบ</button>
    </div>
  </FormPanel>
</template>

<style scoped>
.welcome {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1.75rem;
}

.welcome__greeting {
  margin: 0;
  width: 100%;
  padding: 1.4rem;
  border-left: 3px solid var(--color-gold);
  background: var(--color-brand-soft);
  color: var(--color-brand-dark);
  font-size: clamp(1.1rem, 2.5vw, 1.35rem);
  font-weight: 700;
  text-align: center;
}

.welcome__sign-out {
  min-width: 10rem;
  min-height: 2.8rem;
  padding: 0.6rem 1.5rem;
  border: 1px solid var(--color-brand);
  border-radius: var(--radius-control);
  background: var(--color-brand);
  color: #ffffff;
  font: inherit;
  font-weight: 700;
  cursor: pointer;
  transition: background-color 140ms ease, border-color 140ms ease, transform 140ms ease;
}

.welcome__sign-out:hover {
  border-color: var(--color-brand-dark);
  background: var(--color-brand-dark);
  transform: translateY(-1px);
}
</style>
