<script setup lang="ts">
/**
 * The form's submit control. It owns the "busy" state so no screen can leave
 * a button clickable while its request is still in flight, which is what
 * produces duplicate registrations.
 */
defineProps<{
  busy?: boolean | undefined
  /** Replaces the label while the request is in flight. */
  busyLabel?: string | undefined
}>()
</script>

<template>
  <button class="submit" type="submit" :disabled="busy" :aria-busy="busy">
    <span v-if="busy">{{ busyLabel ?? 'กำลังดำเนินการ…' }}</span>
    <slot v-else />
  </button>
</template>

<style scoped>
.submit {
  min-width: 11rem;
  min-height: 2.8rem;
  padding: 0.6rem 1.5rem;
  border: 1px solid var(--color-brand);
  border-radius: var(--radius-control);
  background: var(--color-brand);
  color: #ffffff;
  font: inherit;
  font-weight: 700;
  box-shadow: inset 0 -2px 0 rgb(0 0 0 / 10%);
  cursor: pointer;
  transition: background-color 140ms ease, border-color 140ms ease, transform 140ms ease;
}

.submit:hover:not(:disabled) {
  border-color: var(--color-brand-dark);
  background: var(--color-brand-dark);
  transform: translateY(-1px);
}

.submit:disabled {
  cursor: progress;
  opacity: 0.65;
}

@media (max-width: 34rem) {
  .submit {
    width: 100%;
  }
}
</style>
