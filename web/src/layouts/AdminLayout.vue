<script setup>
import { useRouter } from 'vue-router'

import { useAuthStore } from '../stores/auth'

const router = useRouter()
const authStore = useAuthStore()

async function onLogout() {
  await authStore.logout()
  router.push({ name: 'admin-login' })
}
</script>

<template>
  <div class="admin-layout">
    <aside class="admin-sidebar card">
      <div class="sidebar-top">
        <p class="eyebrow">Admin</p>
        <h1 class="serif">Writing Room</h1>
        <p class="muted">单管理员工作台，负责发文、整理分类和站点信息。</p>
      </div>

      <nav class="sidebar-nav">
        <router-link to="/admin">控制台</router-link>
        <router-link to="/admin/posts">文章</router-link>
        <router-link to="/admin/categories">分类</router-link>
        <router-link to="/admin/tags">标签</router-link>
        <router-link to="/admin/settings">设置</router-link>
      </nav>

      <div class="sidebar-bottom">
        <p class="muted">已登录：{{ authStore.user?.username }}</p>
        <button class="button secondary" type="button" @click="onLogout">退出登录</button>
      </div>
    </aside>

    <section class="admin-content">
      <router-view />
    </section>
  </div>
</template>

<style scoped>
.admin-layout {
  min-height: 100vh;
  padding: 18px;
  display: grid;
  grid-template-columns: 320px 1fr;
  gap: 18px;
}

.admin-sidebar {
  position: sticky;
  top: 18px;
  align-self: start;
  padding: 28px;
}

.admin-sidebar h1 {
  margin: 0 0 12px;
  font-size: 2rem;
  line-height: 0.98;
}

.sidebar-nav {
  display: grid;
  gap: 8px;
  margin: 28px 0;
}

.sidebar-nav a {
  border: 1px solid transparent;
  border-radius: 14px;
  padding: 12px 14px;
  color: var(--ink-soft);
  transition: background 180ms ease, border-color 180ms ease;
}

.sidebar-nav a.router-link-active {
  background: rgba(255, 255, 255, 0.56);
  border-color: var(--line-strong);
  color: var(--ink);
}

.sidebar-bottom {
  display: grid;
  gap: 10px;
}

.admin-content {
  min-width: 0;
}

@media (max-width: 980px) {
  .admin-layout {
    grid-template-columns: 1fr;
  }

  .admin-sidebar {
    position: static;
  }
}
</style>
