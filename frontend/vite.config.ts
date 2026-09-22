import { fileURLToPath, URL } from 'node:url'

import vue from '@vitejs/plugin-vue'
// From vitest/config rather than vite, so the `test` section below is typed.
import { defineConfig } from 'vitest/config'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    // '@' points at src so a deep import does not turn into '../../../'.
    alias: { '@': fileURLToPath(new URL('./src', import.meta.url)) },
  },
  server: {
    port: 4200,
    // Fail instead of silently moving to another port: the API's CORS
    // allowlist names this one, and a different port would be refused with a
    // confusing browser error.
    strictPort: true,
  },
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: ['./tests/setup.ts'],
    include: ['src/**/*.spec.ts'],
  },
})
