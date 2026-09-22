import { createRouter, createWebHistory } from 'vue-router'
import type { RouteLocationNormalized, RouteRecordRaw, RouterHistory } from 'vue-router'

import { useAuthStore } from '@/features/auth/stores/auth.store'

/**
 * The three screens of the assignment.
 *
 * Pages are loaded lazily so the first paint of the sign in screen does not
 * carry the code for the ones behind it.
 */
const routes: RouteRecordRaw[] = [
  { path: '/', redirect: { name: 'login' } },
  {
    // IT 02-1
    path: '/login',
    name: 'login',
    component: () => import('@/features/auth/pages/LoginPage.vue'),
    meta: { guestOnly: true, title: 'IT 02-1' },
  },
  {
    // IT 02-2
    path: '/register',
    name: 'register',
    component: () => import('@/features/auth/pages/RegisterPage.vue'),
    meta: { guestOnly: true, title: 'IT 02-2' },
  },
  {
    // IT 02-3
    path: '/welcome',
    name: 'welcome',
    component: () => import('@/features/auth/pages/WelcomePage.vue'),
    meta: { requiresAuth: true, title: 'IT 02-3' },
  },
  // Anything else is not a screen this app has.
  { path: '/:pathMatch(.*)*', redirect: { name: 'login' } },
]

/**
 * Builds the router, taking its history so a test can drive it in memory
 * instead of sharing the one browser history between cases.
 */
export function createAppRouter(history: RouterHistory = createWebHistory()) {
  const router = createRouter({ history, routes })
  router.beforeEach(guard)
  return router
}

/** The instance the running app uses. */
export const router = createAppRouter()

/**
 * The guard decides on session state, not on the presence of a token: after a
 * reload the token is there but unverified, so `restore` asks the API first.
 *
 * This is a convenience, not a protection. The data behind /welcome comes from
 * an endpoint that checks the token itself, so a person who edits the URL sees
 * nothing they could not already see.
 */
async function guard(to: RouteLocationNormalized) {
  const auth = useAuthStore()

  if (to.meta['requiresAuth'] === true) {
    const signedIn = await auth.restore()
    if (!signedIn) {
      // `redirect` lets the sign in screen send them where they were going.
      return { name: 'login', query: { redirect: to.fullPath } }
    }
    return true
  }

  if (to.meta['guestOnly'] === true && (await auth.restore())) {
    return { name: 'welcome' }
  }
  return true
}
