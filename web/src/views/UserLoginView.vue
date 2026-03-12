<script setup>
import { reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { useUserAuthStore } from '../stores/userAuth'

const route = useRoute()
const router = useRouter()
const userAuthStore = useUserAuthStore()

const form = reactive({
  account: '',
  password: '',
})
const errorMessage = ref('')

async function onSubmit() {
  errorMessage.value = ''
  try {
    await userAuthStore.login(form)
    router.push(route.query.redirect || '/')
  } catch (error) {
    errorMessage.value = error.message
  }
}
</script>

<template>
  <section class="auth-page">
    <div class="auth-card card fade-in">
      <p class="eyebrow">User Login</p>
      <h1 class="serif">普通用户登录</h1>
      <p class="muted">登录后可参与评论与点赞互动。</p>

      <div class="stack">
        <input v-model="form.account" class="input" placeholder="用户名或邮箱" />
        <input v-model="form.password" class="input" type="password" placeholder="密码" />
        <p v-if="errorMessage" class="error-text">{{ errorMessage }}</p>
        <button class="button" :disabled="userAuthStore.pending" type="button" @click="onSubmit">
          {{ userAuthStore.pending ? '登录中...' : '登录' }}
        </button>
        <p class="muted">
          没有账号？
          <router-link class="inline-link" to="/register">去注册</router-link>
        </p>
      </div>
    </div>
  </section>
</template>

<style scoped>
.auth-page {
  min-height: calc(100vh - 180px);
  display: grid;
  place-items: center;
  padding: 24px 0;
}

.auth-card {
  width: min(500px, 100%);
  padding: 34px;
}

.auth-card h1 {
  margin: 0 0 14px;
  font-size: 2.2rem;
  line-height: 1.02;
}

.error-text {
  margin: 0;
  color: var(--danger);
}

.inline-link {
  color: var(--accent);
}
</style>
