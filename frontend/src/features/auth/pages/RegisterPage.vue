<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'

import FormAlert from '@/shared/ui/FormAlert.vue'
import FormField from '@/shared/ui/FormField.vue'
import FormPanel from '@/shared/ui/FormPanel.vue'
import SubmitButton from '@/shared/ui/SubmitButton.vue'
import { ApiError } from '@/shared/api'

import type { FieldErrors } from '../model/credentials'
import { hasErrors, validateRegistration } from '../model/credentials'
import { useAuthStore } from '../stores/auth.store'

/** IT 02-2 — sign up, then back to IT 02-1 to sign in. */

const auth = useAuthStore()
const router = useRouter()

const form = reactive({ username: '', password: '', confirmPassword: '' })
const fieldErrors = ref<FieldErrors>({})
const formError = ref('')
const submitting = ref(false)

async function submit(): Promise<void> {
  formError.value = ''
  // The same rules the API applies, checked here first so a mismatched
  // confirmation is caught without a round trip.
  fieldErrors.value = validateRegistration(form)
  if (hasErrors(fieldErrors.value)) {
    return
  }

  submitting.value = true
  try {
    await auth.register({
      username: form.username.trim(),
      password: form.password,
      confirmPassword: form.confirmPassword,
    })
    // Registering does not sign the person in; the flag tells IT 02-1 to
    // confirm what just happened.
    await router.replace({ name: 'login', query: { registered: '1' } })
  } catch (error) {
    if (!(error instanceof ApiError)) {
      throw error
    }
    formError.value = error.message
    fieldErrors.value = { ...error.fieldErrors }
    // Both secrets are cleared: whatever was wrong, they have to be typed
    // again, and leaving them on screen serves no one.
    form.password = ''
    form.confirmPassword = ''
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <FormPanel title="IT 02-2">
    <form class="form" novalidate @submit.prevent="submit">
      <FormAlert v-if="formError">{{ formError }}</FormAlert>

      <FormField
        v-model="form.username"
        label="User"
        autocomplete="username"
        :errors="fieldErrors['username']"
        :disabled="submitting"
      />

      <FormField
        v-model="form.password"
        label="Password"
        type="password"
        autocomplete="new-password"
        :errors="fieldErrors['password']"
        :disabled="submitting"
      />

      <FormField
        v-model="form.confirmPassword"
        label="Confirm Password"
        type="password"
        autocomplete="new-password"
        :errors="fieldErrors['confirmPassword']"
        :disabled="submitting"
      />

      <div class="form__actions">
        <SubmitButton :busy="submitting" busy-label="กำลังสมัคร…">สมัครสมาชิก</SubmitButton>
        <RouterLink class="form__link" :to="{ name: 'login' }">กลับไปหน้าลงชื่อเข้าใช้งาน</RouterLink>
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
