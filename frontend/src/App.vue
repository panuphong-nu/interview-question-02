<script setup lang="ts">
import { watchEffect } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()

// The document title follows the screen, so the browser tab and the history
// entries say which of the three forms is open.
watchEffect(() => {
  const screen = route.meta['title']
  document.title = typeof screen === 'string' ? `${screen} — ระบบสมาชิก` : 'ระบบสมาชิก'
})
</script>

<template>
  <!-- A single transitionless outlet: the three screens are full page forms,
       and animating between them would only delay the next input. -->
  <RouterView />
</template>
