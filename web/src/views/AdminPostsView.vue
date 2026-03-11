<script setup>
import { onMounted, ref } from 'vue'

import { deletePost, fetchAdminPosts } from '../api/admin'
import EmptyState from '../components/EmptyState.vue'

const loading = ref(true)
const posts = ref([])
const filter = ref('')
const errorMessage = ref('')

async function loadPosts() {
  loading.value = true
  errorMessage.value = ''
  try {
    const response = await fetchAdminPosts({ status: filter.value, page_size: 50 })
    posts.value = response.data
  } catch (error) {
    errorMessage.value = error.message
  } finally {
    loading.value = false
  }
}

async function removePost(id) {
  if (!window.confirm('确认删除这篇文章吗？')) {
    return
  }
  await deletePost(id)
  await loadPosts()
}

onMounted(loadPosts)
</script>

<template>
  <section class="stack fade-in">
    <div class="toolbar">
      <div>
        <p class="eyebrow">Posts</p>
        <h1 class="serif">文章管理</h1>
      </div>
      <div class="toolbar-actions">
        <select v-model="filter" class="select" @change="loadPosts">
          <option value="">全部状态</option>
          <option value="draft">草稿</option>
          <option value="published">已发布</option>
        </select>
        <router-link class="button" to="/admin/posts/new">新建文章</router-link>
      </div>
    </div>

    <p v-if="errorMessage" class="error-text">{{ errorMessage }}</p>

    <EmptyState
      v-if="!loading && posts.length === 0"
      title="还没有文章"
      description="先创建一篇草稿，再逐步补充内容和封面。"
    />

    <section v-else class="card table-card">
      <article v-for="post in posts" :key="post.id" class="post-row">
        <div class="post-main">
          <strong>{{ post.title }}</strong>
          <p class="muted">{{ post.slug }}</p>
          <p class="muted">{{ post.summary || '暂无摘要' }}</p>
        </div>
        <div class="post-side">
          <span class="chip">{{ post.status === 'published' ? '已发布' : '草稿' }}</span>
          <router-link class="button ghost" :to="`/admin/posts/${post.id}`">编辑</router-link>
          <button class="button danger" type="button" @click="removePost(post.id)">删除</button>
        </div>
      </article>
    </section>
  </section>
</template>

<style scoped>
.toolbar {
  display: flex;
  justify-content: space-between;
  gap: 18px;
  flex-wrap: wrap;
}

.toolbar h1 {
  margin: 0;
}

.toolbar-actions {
  display: flex;
  gap: 12px;
  align-items: center;
  flex-wrap: wrap;
}

.table-card {
  padding: 24px;
}

.post-row {
  display: flex;
  justify-content: space-between;
  gap: 22px;
  padding: 18px 0;
  border-top: 1px solid var(--line);
}

.post-row:first-child {
  border-top: none;
  padding-top: 0;
}

.post-main p {
  margin: 6px 0 0;
}

.post-side {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.error-text {
  margin: 0;
  color: var(--danger);
}

@media (max-width: 860px) {
  .post-row {
    flex-direction: column;
  }
}
</style>
