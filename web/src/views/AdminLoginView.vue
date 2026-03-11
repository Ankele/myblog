<script setup>
import { reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { useAuthStore } from '../stores/auth'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const form = reactive({
  username: 'admin',
  password: 'admin123456',
})
const errorMessage = ref('')

async function onSubmit() {
  errorMessage.value = ''
  try {
    await authStore.login(form)
    router.push(route.query.redirect || '/admin')
  } catch (error) {
    errorMessage.value = error.message
  }
}
</script>

<template>
  <section class="login-page">
    <div class="login-card card fade-in">
      <p class="eyebrow">Admin Login</p>
      <h1 class="serif">进入写作后台</h1>
      <p class="muted">
        默认管理员账号已在迁移阶段初始化。首次启动后可以直接登录并开始发布文章。
      </p>

      <div class="stack">
        <input v-model="form.username" class="input" placeholder="用户名" />
        <input v-model="form.password" class="input" type="password" placeholder="密码" />
        <p v-if="errorMessage" class="error-text">{{ errorMessage }}</p>
        <button class="button" :disabled="authStore.pending" type="button" @click="onSubmit">
          {{ authStore.pending ? '登录中...' : '登录后台' }}
        </button>
      </div>
    </div>
  </section>
</template>

<style scoped>
.login-page {
  min-height: 100vh;
  display: grid;
  place-items: center;
  padding: 24px;
}

.login-card {
  width: min(460px, 100%);
  padding: 34px;
}

.login-card h1 {
  margin: 0 0 14px;
  font-size: 2.4rem;
  line-height: 0.98;
}

.error-text {
  margin: 0;
  color: var(--danger);
}
</style>
