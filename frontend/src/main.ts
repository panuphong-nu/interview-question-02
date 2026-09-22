import { createPinia } from 'pinia'
import { createApp } from 'vue'

import App from './App.vue'
import { router } from './app/router'
import './shared/ui/styles.css'

/**
 * The composition root of the frontend: it creates the app, installs the
 * plugins and mounts. Nothing else reaches for a global.
 */
createApp(App).use(createPinia()).use(router).mount('#app')
