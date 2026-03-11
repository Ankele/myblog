<script setup>
import { reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { useUserAuthStore } from '../stores/userAuth'

const route = useRoute()
const router = useRouter()
const userAuthStore = useUserAuthStore()

const form = reactive({
  username: '',
  email: '',
  password: '',
  confirmPassword: '',
})
const errorMessage = ref('')

async function onSubmit() {
  errorMessage.value = ''
  if (form.password !== form.confirmPassword) {
    errorMessage.value = '两次密码输入不一致'
    return
  }

  try {
    await userAuthStore.register(form)
    router.push(route.query.redirect || '/')
  } catch (error) {
    errorMessage.value = error.message
  }
}
</script>

<template>
  <section class="auth-page">
    <div class="auth-card card fade-in">
      <p class="eyebrow">User Register</p>
      <h1 class="serif">普通用户注册</h1>
      <p class="muted">注册后可以在文章下评论和点赞。</p>

      <div class="stack">
        <input v-model="form.username" class="input" placeholder="用户名" />
        <input v-model="form.email" class="input" type="email" placeholder="邮箱" />
        <input v-model="form.password" class="input" type="password" placeholder="密码（至少8位，含字母和数字）" />
        <input v-model="form.confirmPassword" class="input" type="password" placeholder="确认密码" />
        <p v-if="errorMessage" class="error-text">{{ errorMessage }}</p>
        <button class="button" :disabled="userAuthStore.pending" type="button" @click="onSubmit">
          {{ userAuthStore.pending ? '注册中...' : '注册并登录' }}
        </button>
        <p class="muted">
          已有账号？
          <router-link class="inline-link" to="/login">去登录</router-link>
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
