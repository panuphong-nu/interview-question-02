<script setup lang="ts">
/**
 * A message about the form as a whole: a refused sign in, a taken username, a
 * server that could not be reached, or the confirmation after registering.
 *
 * role="alert" is what makes a screen reader announce it, which matters here
 * because the message appears far from the control that caused it.
 */
defineProps<{
  tone?: 'error' | 'success' | undefined
}>()
</script>

<template>
  <p class="alert" :class="`alert--${tone ?? 'error'}`" role="alert">
    <slot />
  </p>
</template>

<style scoped>
.alert {
  margin: 0;
  padding: 0.75rem 0.85rem;
  border-left: 3px solid currentColor;
  border-radius: var(--radius-control);
  font-size: 0.9375rem;
}

.alert--error {
  background: var(--color-danger-surface);
  color: var(--color-danger);
}

.alert--success {
  background: var(--color-success-surface);
  color: var(--color-success-text);
}
</style>
