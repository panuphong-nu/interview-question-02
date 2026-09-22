<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import FormAlert from '@/shared/ui/FormAlert.vue'
import FormField from '@/shared/ui/FormField.vue'
import FormPanel from '@/shared/ui/FormPanel.vue'
import SubmitButton from '@/shared/ui/SubmitButton.vue'
import { ApiError } from '@/shared/api'

import type { FieldErrors } from '../model/credentials'
import { hasErrors, validateLogin } from '../model/credentials'
import { useAuthStore } from '../stores/auth.store'

/** IT 02-1 — sign in, and the way through to the sign up form. */

const auth = useAuthStore()
const router = useRouter()
const route = useRoute()

const form = reactive({ username: '', password: '' })
const fieldErrors = ref<FieldErrors>({})
const formError = ref('')
const submitting = ref(false)

// Shown after registering: RegisterPage sends the person here with this flag
// rather than signing them in, as the assignment's flow requires.
const justRegistered = computed(() => route.query['registered'] === '1')

async function submit(): Promise<void> {
  formError.value = ''
  fieldErrors.value = validateLogin(form)
  if (hasErrors(fieldErrors.value)) {
    return
  }

  submitting.value = true
  try {
    await auth.login({ username: form.username.trim(), password: form.password })
    // Back to where the guard interrupted, when it did.
    const redirect = route.query['redirect']
    await router.replace(typeof redirect === 'string' ? redirect : { name: 'welcome' })
  } catch (error) {
    if (!(error instanceof ApiError)) {
      throw error
    }
    // The API answers a wrong password and an unknown username identically,
    // and so does this screen.
    formError.value = error.message
    fieldErrors.value = { ...error.fieldErrors }
    // The password is cleared but the username is kept, so a mistyped password
    // costs one field rather than the whole form.
    form.password = ''
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <FormPanel title="IT 02-1">
    <form class="form" novalidate @submit.prevent="submit">
      <FormAlert v-if="justRegistered && !formError" tone="success">
        สมัครสมาชิกเรียบร้อยแล้ว กรุณาลงชื่อเข้าใช้งาน
      </FormAlert>
      <FormAlert v-if="formError">{{ formError }}</FormAlert>

      <FormField
        v-model="form.username"
        label="User"
        autocomplete="username"
        :errors="fieldErrors['username']"
        :disabled="submitting"
      />

      <!-- type="password" is what masks the characters as * in the browser. -->
      <FormField
        v-model="form.password"
        label="Password"
        type="password"
        autocomplete="current-password"
        :errors="fieldErrors['password']"
        :disabled="submitting"
      />

      <div class="form__actions">
        <SubmitButton :busy="submitting" busy-label="กำลังเข้าสู่ระบบ…">
          ลงชื่อเข้าใช้งาน
        </SubmitButton>
        <RouterLink class="form__link" :to="{ name: 'register' }">สมัครสมาชิก</RouterLink>
      </div>
    </form>
  </FormPanel>
</template>

<style scoped>
.form {
  display: flex;
  flex-direction: column;
  gap: 1.15rem;
}

.form__actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  margin-top: var(--space-2);
}

.form__link {
  color: var(--color-brand);
  font-size: 0.9rem;
  font-weight: 650;
  text-decoration-color: var(--color-gold);
  text-decoration-thickness: 2px;
  text-underline-offset: 0.2rem;
}

@media (max-width: 34rem) {
  .form__actions {
    flex-direction: column;
  }

  .form__link {
    width: 100%;
    text-align: center;
  }
}
</style>
