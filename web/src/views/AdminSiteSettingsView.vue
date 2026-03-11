<script setup>
import { onMounted, reactive, ref } from 'vue'

import { fetchAdminSite, updateSite } from '../api/admin'

const loading = ref(true)
const saving = ref(false)
const successMessage = ref('')
const errorMessage = ref('')

const form = reactive({
  site_name: '',
  tagline: '',
  description: '',
  avatar: '',
  footer_text: '',
  seo_title: '',
  seo_description: '',
  github_url: '',
  twitter_url: '',
  linkedin_url: '',
  email: '',
  about_title: '',
  about_content: '',
})

async function loadSite() {
  loading.value = true
  try {
    const response = await fetchAdminSite()
    Object.assign(form, response.data)
  } finally {
    loading.value = false
  }
}

async function onSave() {
  saving.value = true
  successMessage.value = ''
  errorMessage.value = ''
  try {
    await updateSite(form)
    successMessage.value = '站点设置已更新。'
  } catch (error) {
    errorMessage.value = error.message
  } finally {
    saving.value = false
  }
}

onMounted(loadSite)
</script>

<template>
  <section class="stack fade-in">
    <div class="toolbar">
      <div>
        <p class="eyebrow">Site Settings</p>
        <h1 class="serif">站点设置</h1>
      </div>
      <button class="button" :disabled="saving || loading" type="button" @click="onSave">
        {{ saving ? '保存中...' : '保存设置' }}
      </button>
    </div>

    <p v-if="successMessage" class="success-text">{{ successMessage }}</p>
    <p v-if="errorMessage" class="error-text">{{ errorMessage }}</p>

    <section class="card settings-form">
      <div class="grid-two">
        <label class="field">
          <span>站点名称</span>
          <input v-model="form.site_name" class="input" />
        </label>
        <label class="field">
          <span>副标题</span>
          <input v-model="form.tagline" class="input" />
        </label>
      </div>

      <label class="field">
        <span>站点描述</span>
        <textarea v-model="form.description" class="textarea"></textarea>
      </label>

      <div class="grid-two">
        <label class="field">
          <span>头像 URL</span>
          <input v-model="form.avatar" class="input" />
        </label>
        <label class="field">
          <span>页脚文案</span>
          <input v-model="form.footer_text" class="input" />
        </label>
      </div>

      <div class="grid-two">
        <label class="field">
          <span>SEO 标题</span>
          <input v-model="form.seo_title" class="input" />
        </label>
        <label class="field">
          <span>SEO 描述</span>
          <input v-model="form.seo_description" class="input" />
        </label>
      </div>

      <div class="grid-two">
        <label class="field">
          <span>GitHub URL</span>
          <input v-model="form.github_url" class="input" />
        </label>
        <label class="field">
          <span>Twitter URL</span>
          <input v-model="form.twitter_url" class="input" />
        </label>
      </div>

      <div class="grid-two">
        <label class="field">
          <span>LinkedIn URL</span>
          <input v-model="form.linkedin_url" class="input" />
        </label>
        <label class="field">
          <span>Email</span>
          <input v-model="form.email" class="input" />
        </label>
      </div>

      <label class="field">
        <span>关于页标题</span>
        <input v-model="form.about_title" class="input" />
      </label>

      <label class="field">
        <span>关于页内容（支持 HTML）</span>
        <textarea v-model="form.about_content" class="textarea large"></textarea>
      </label>
    </section>
  </section>
</template>

<style scoped>
.toolbar {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  align-items: end;
  flex-wrap: wrap;
}

.toolbar h1 {
  margin: 0;
}

.settings-form {
  display: grid;
  gap: 18px;
  padding: 24px;
}

.field {
  display: grid;
  gap: 10px;
}

.field > span {
  color: var(--ink-soft);
  font-size: 0.92rem;
}

.large {
  min-height: 220px;
}

.success-text {
  margin: 0;
  color: var(--success);
}

.error-text {
  margin: 0;
  color: var(--danger);
}
</style>
