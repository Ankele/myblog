import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from './App.vue'
import router from './router'
import { useAuthStore } from './stores/auth'
import { useSiteStore } from './stores/site'
import { useUserAuthStore } from './stores/userAuth'
import './assets/main.css'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.use(router)

const authStore = useAuthStore(pinia)
const siteStore = useSiteStore(pinia)
const userAuthStore = useUserAuthStore(pinia)

router.beforeEach(async (to) => {
  if (!authStore.initialized) {
    await authStore.fetchMe().catch(() => {})
  }
  if (!userAuthStore.initialized) {
    await userAuthStore.fetchMe().catch(() => {})
  }

  if (!siteStore.loaded && !to.meta.adminOnly) {
    await siteStore.ensureSite().catch(() => {})
  }

  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    return {
      name: 'admin-login',
      query: { redirect: to.fullPath },
    }
  }

  if (to.name === 'admin-login' && authStore.isAuthenticated) {
    return { name: 'admin-dashboard' }
  }
  if (to.meta.userGuestOnly && userAuthStore.isAuthenticated) {
    return { name: 'home' }
  }

  return true
})

router.afterEach((to) => {
  const siteName = siteStore.site.site_name || 'MyBlog'
  const title = to.meta.title ? `${to.meta.title} · ${siteName}` : siteName
  document.title = title
})

app.mount('#app')
