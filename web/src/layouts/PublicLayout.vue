<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'

import { useSiteStore } from '../stores/site'
import { useUserAuthStore } from '../stores/userAuth'

const route = useRoute()
const siteStore = useSiteStore()
const userAuthStore = useUserAuthStore()

const activePath = computed(() => route.path)

async function onUserLogout() {
  await userAuthStore.logout()
}
</script>

<template>
  <div class="public-layout fade-in">
    <header class="page-shell site-header">
      <router-link class="brand serif" to="/">
        <span class="brand-mark"></span>
        <div>
          <strong>{{ siteStore.site.site_name || 'MyBlog' }}</strong>
          <p>{{ siteStore.site.tagline || 'A personal notebook of code and craft.' }}</p>
        </div>
      </router-link>

      <nav class="site-nav">
        <router-link :class="{ active: activePath === '/' }" to="/">首页</router-link>
        <router-link :class="{ active: activePath.startsWith('/about') }" to="/about">关于</router-link>
        <template v-if="!userAuthStore.isAuthenticated">
          <router-link :class="{ active: activePath.startsWith('/login') }" to="/login">登录</router-link>
          <router-link :class="{ active: activePath.startsWith('/register') }" to="/register">注册</router-link>
        </template>
        <template v-else>
          <span class="chip user-chip">{{ userAuthStore.user?.username }}</span>
          <button class="logout-btn" type="button" @click="onUserLogout">退出登录</button>
        </template>
        <router-link class="admin-link" to="/admin">后台</router-link>
      </nav>
    </header>

    <main>
      <router-view />
    </main>

    <footer class="page-shell site-footer">
      <p>{{ siteStore.site.footer_text || 'Built with care.' }}</p>
      <div class="footer-links">
        <a v-if="siteStore.site.github_url" :href="siteStore.site.github_url" target="_blank" rel="noreferrer">GitHub</a>
        <a v-if="siteStore.site.twitter_url" :href="siteStore.site.twitter_url" target="_blank" rel="noreferrer">Twitter</a>
        <a v-if="siteStore.site.linkedin_url" :href="siteStore.site.linkedin_url" target="_blank" rel="noreferrer">LinkedIn</a>
      </div>
    </footer>
  </div>
</template>

<style scoped>
.public-layout {
  min-height: 100vh;
  padding: 18px 0 40px;
}

.site-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
  padding: 22px 0 34px;
}

.brand {
  display: flex;
  align-items: center;
  gap: 14px;
}

.brand-mark {
  width: 18px;
  height: 18px;
  border-radius: 50%;
  background: linear-gradient(135deg, var(--accent), #d7a56a);
  box-shadow: 0 0 0 8px rgba(154, 93, 47, 0.12);
}

.brand strong {
  display: block;
  font-size: 1.35rem;
  letter-spacing: -0.04em;
}

.brand p {
  margin: 4px 0 0;
  color: var(--ink-soft);
  font-size: 0.92rem;
}

.site-nav {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.site-nav a {
  padding: 10px 14px;
  border-radius: 999px;
  color: var(--ink-soft);
  transition: background 180ms ease, color 180ms ease;
}

.site-nav a.active,
.site-nav a:hover {
  background: rgba(255, 255, 255, 0.45);
  color: var(--ink);
}

.admin-link {
  border: 1px solid var(--line);
}

.user-chip {
  pointer-events: none;
}

.logout-btn {
  border: 1px solid var(--line);
  background: transparent;
  color: var(--ink-soft);
  padding: 10px 14px;
  border-radius: 999px;
  cursor: pointer;
  transition: background 180ms ease, color 180ms ease;
}

.logout-btn:hover {
  background: rgba(255, 255, 255, 0.45);
  color: var(--ink);
}

.site-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding-top: 36px;
  color: var(--ink-soft);
  font-size: 0.92rem;
}

.footer-links {
  display: flex;
  gap: 14px;
}

@media (max-width: 760px) {
  .site-header,
  .site-footer {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>
