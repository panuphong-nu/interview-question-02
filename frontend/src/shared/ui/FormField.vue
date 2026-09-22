<script setup lang="ts">
import { computed, useId } from 'vue'

/**
 * One labelled input with its error message.
 *
 * The label, the input, the error and the aria-* wiring that ties them
 * together are produced in one place, so no screen can accidentally ship an
 * input a screen reader cannot name.
 */
// The optional props are written as `| undefined` because
// exactOptionalPropertyTypes distinguishes "absent" from "present and
// undefined", and a template binding always passes the property.
const props = defineProps<{
  label: string
  type?: 'text' | 'password' | undefined
  autocomplete?: string | undefined
  /** Messages for this field, from the browser's checks or the API's. */
  errors?: string[] | undefined
  disabled?: boolean | undefined
}>()

const model = defineModel<string>({ required: true })

// useId gives a value that is stable across a server render and the client,
// and unique even when two fields share a label.
const inputId = useId()
const errorId = computed(() => `${inputId}-error`)
const hasError = computed(() => (props.errors?.length ?? 0) > 0)
</script>

<template>
  <div class="field">
    <label class="field__label" :for="inputId">{{ label }}</label>

    <div class="field__control">
      <input
        :id="inputId"
        v-model="model"
        class="field__input"
        :class="{ 'field__input--invalid': hasError }"
        :type="type ?? 'text'"
        :autocomplete="autocomplete"
        :disabled="disabled"
        :aria-invalid="hasError"
        :aria-describedby="hasError ? errorId : undefined"
      />

      <!-- Every message for the field, not just the first: the API reports
           each broken rule, and showing one at a time turns fixing a form
           into a guessing game. -->
      <p v-if="hasError" :id="errorId" class="field__error">
        <span v-for="message in errors" :key="message">{{ message }}</span>
      </p>
    </div>
  </div>
</template>

<style scoped>
.field {
  display: flex;
  flex-direction: column;
  gap: 0.45rem;
}

.field__label {
  color: #2e3f35;
  font-size: 0.9rem;
  font-weight: 650;
}

.field__control {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
}

.field__input {
  width: 100%;
  min-height: 2.9rem;
  padding: 0.65rem 0.8rem;
  border: 1px solid var(--color-field-border);
  border-radius: var(--radius-control);
  font: inherit;
  background: var(--color-surface);
  color: inherit;
  transition: border-color 140ms ease, box-shadow 140ms ease;
}

.field__input:focus {
  border-color: var(--color-brand);
  outline: none;
  box-shadow: 0 0 0 3px rgb(0 99 61 / 12%);
}

.field__input:disabled {
  background: var(--color-page);
  color: var(--color-muted);
}

.field__input--invalid {
  border-color: var(--color-danger);
}

.field__error {
  margin: 0;
  display: flex;
  flex-direction: column;
  color: var(--color-danger);
  font-size: 0.875rem;
}
</style>
