<script setup>
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import {
  fetchAdminCategories,
  fetchAdminPost,
  fetchAdminTags,
  importMarkdownPost,
  previewMarkdown,
  savePost,
  uploadFile,
} from '../api/admin'

const route = useRoute()
const router = useRouter()

const loading = ref(false)
const saving = ref(false)
const previewHtml = ref('')
const categories = ref([])
const tags = ref([])
const errorMessage = ref('')
const successMessage = ref('')
let previewTimer = null

const form = reactive({
  title: '',
  slug: '',
  summary: '',
  cover_image: '',
  markdown_content: '# 你好，世界\n\n开始写作吧。',
  status: 'draft',
  category_id: null,
  tag_ids: [],
})

const postId = computed(() => route.params.id)
const isEdit = computed(() => Boolean(postId.value))

async function loadReferenceData() {
  const [categoriesResponse, tagsResponse] = await Promise.all([
    fetchAdminCategories(),
    fetchAdminTags(),
  ])
  categories.value = categoriesResponse.data
  tags.value = tagsResponse.data
}

async function loadPost() {
  if (!isEdit.value) {
    previewHtml.value = ''
    await renderPreview()
    return
  }

  loading.value = true
  try {
    const response = await fetchAdminPost(postId.value)
    Object.assign(form, {
      title: response.data.title,
      slug: response.data.slug,
      summary: response.data.summary,
      cover_image: response.data.cover_image,
      markdown_content: response.data.markdown_content,
      status: response.data.status,
      category_id: response.data.category_id,
      tag_ids: (response.data.tags || []).map((tag) => tag.id),
    })
    previewHtml.value = response.data.html_content
  } finally {
    loading.value = false
  }
}

async function renderPreview() {
  if (!form.markdown_content.trim()) {
    previewHtml.value = ''
    return
  }
  const response = await previewMarkdown(form.markdown_content)
  previewHtml.value = response.data.html_content
}

watch(
  () => form.markdown_content,
  () => {
    window.clearTimeout(previewTimer)
    previewTimer = window.setTimeout(() => {
      renderPreview().catch(() => {})
    }, 320)
  },
)

async function onSave() {
  saving.value = true
  errorMessage.value = ''
  successMessage.value = ''
  try {
    const response = await savePost(form, postId.value)
    successMessage.value = '文章已保存。'
    if (!isEdit.value) {
      router.replace(`/admin/posts/${response.data.id}`)
    } else {
      await loadPost()
    }
  } catch (error) {
    errorMessage.value = error.message
  } finally {
    saving.value = false
  }
}

async function onUploadCover(event) {
  const [file] = event.target.files || []
  if (!file) {
    return
  }
  const response = await uploadFile(file)
  form.cover_image = response.data.path
}

async function onInsertImage(event) {
  const [file] = event.target.files || []
  if (!file) {
    return
  }
  const response = await uploadFile(file)
  form.markdown_content += `\n\n![${file.name}](${response.data.path})\n`
  await renderPreview()
}

async function onImportMarkdown(event) {
  const [file] = event.target.files || []
  if (!file) {
    return
  }

  saving.value = true
  errorMessage.value = ''
  successMessage.value = ''
  try {
    const response = await importMarkdownPost(file, {
      status: form.status,
      category_id: form.category_id,
      tag_ids: form.tag_ids,
      cover_image: form.cover_image,
    })
    successMessage.value = 'Markdown 已导入为文章，正在进入编辑页。'
    await router.replace(`/admin/posts/${response.data.id}`)
  } catch (error) {
    errorMessage.value = error.message
  } finally {
    saving.value = false
    event.target.value = ''
  }
}

onMounted(async () => {
  await loadReferenceData()
  await loadPost()
})
</script>

<template>
  <section class="stack fade-in">
    <div class="editor-header">
      <div>
        <p class="eyebrow">Editor</p>
        <h1 class="serif">{{ isEdit ? '编辑文章' : '新建文章' }}</h1>
      </div>
      <div class="toolbar-actions">
        <label class="button ghost upload-button">
          导入 Markdown
          <input type="file" accept=".md,.markdown,.txt,text/markdown,text/plain" @change="onImportMarkdown" />
        </label>
        <button class="button secondary" type="button" @click="renderPreview">刷新预览</button>
        <button class="button" :disabled="saving || loading" type="button" @click="onSave">
          {{ saving ? '保存中...' : '保存文章' }}
        </button>
      </div>
    </div>

    <p v-if="errorMessage" class="error-text">{{ errorMessage }}</p>
    <p v-if="successMessage" class="success-text">{{ successMessage }}</p>

    <section class="editor-grid">
      <article class="card editor-form">
        <div class="stack">
          <input v-model="form.title" class="input" placeholder="文章标题" />
          <input v-model="form.slug" class="input" placeholder="slug（可留空自动生成）" />
          <textarea v-model="form.summary" class="textarea" placeholder="摘要"></textarea>

          <div class="grid-two">
            <label class="field">
              <span>状态</span>
              <select v-model="form.status" class="select">
                <option value="draft">草稿</option>
                <option value="published">已发布</option>
              </select>
            </label>
            <label class="field">
              <span>分类</span>
              <select v-model="form.category_id" class="select">
                <option :value="null">未分类</option>
                <option v-for="category in categories" :key="category.id" :value="category.id">
                  {{ category.name }}
                </option>
              </select>
            </label>
          </div>

          <div class="field">
            <span>封面图 URL</span>
            <div class="upload-row">
              <input v-model="form.cover_image" class="input" placeholder="/uploads/cover.jpg" />
              <label class="button secondary upload-button">
                上传封面
                <input type="file" accept="image/*" @change="onUploadCover" />
              </label>
            </div>
          </div>

          <div class="field">
            <span>标签</span>
            <div class="tag-picker">
              <label v-for="tag in tags" :key="tag.id" class="tag-option">
                <input v-model="form.tag_ids" type="checkbox" :value="tag.id" />
                <span>{{ tag.name }}</span>
              </label>
            </div>
          </div>

          <div class="field">
            <span>Markdown 内容</span>
            <div class="upload-row inline-end">
              <label class="button ghost upload-button">
                插入图片
                <input type="file" accept="image/*" @change="onInsertImage" />
              </label>
            </div>
            <textarea v-model="form.markdown_content" class="textarea editor-textarea"></textarea>
          </div>
        </div>
      </article>

      <article class="card editor-preview">
        <p class="eyebrow">Preview</p>
        <div v-if="form.cover_image" class="preview-cover-wrap">
          <img class="preview-cover" :src="form.cover_image" alt="cover" />
        </div>
        <h2 class="serif preview-title">{{ form.title || '未命名文章' }}</h2>
        <p class="muted">{{ form.summary || '摘要会显示在这里。' }}</p>
        <div class="article-html" v-html="previewHtml"></div>
      </article>
    </section>
  </section>
</template>

<style scoped>
.editor-header {
  display: flex;
  justify-content: space-between;
  gap: 18px;
  align-items: end;
  flex-wrap: wrap;
}

.editor-header h1 {
  margin: 0;
}

.editor-grid {
  display: grid;
  grid-template-columns: minmax(0, 1.1fr) minmax(0, 0.9fr);
  gap: 18px;
}

.editor-form,
.editor-preview {
  padding: 24px;
}

.field {
  display: grid;
  gap: 10px;
}

.field > span {
  font-size: 0.92rem;
  color: var(--ink-soft);
}

.editor-textarea {
  min-height: 520px;
  font-family: "SFMono-Regular", Consolas, Menlo, monospace;
}

.tag-picker {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.tag-option {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  border: 1px solid var(--line);
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.45);
}

.upload-row {
  display: flex;
  gap: 10px;
  align-items: center;
  flex-wrap: wrap;
}

.inline-end {
  justify-content: flex-end;
}

.upload-button {
  position: relative;
  overflow: hidden;
}

.upload-button input {
  position: absolute;
  inset: 0;
  opacity: 0;
  cursor: pointer;
}

.preview-title {
  margin: 0 0 8px;
  font-size: 2rem;
  line-height: 1;
}

.preview-cover-wrap {
  margin-bottom: 18px;
}

.preview-cover {
  width: 100%;
  border-radius: 18px;
  aspect-ratio: 16 / 8;
  object-fit: cover;
}

.error-text {
  margin: 0;
  color: var(--danger);
}

.success-text {
  margin: 0;
  color: var(--success);
}

@media (max-width: 1120px) {
  .editor-grid {
    grid-template-columns: 1fr;
  }
}
</style>
