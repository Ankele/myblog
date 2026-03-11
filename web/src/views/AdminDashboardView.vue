<script setup>
import { onMounted, ref } from 'vue'

import { fetchAdminCategories, fetchAdminPosts, fetchAdminTags } from '../api/admin'

const loading = ref(true)
const stats = ref({
  posts: 0,
  draftPosts: 0,
  categories: 0,
  tags: 0,
})
const recentPosts = ref([])

async function loadDashboard() {
  loading.value = true
  try {
    const [postsResponse, categoriesResponse, tagsResponse] = await Promise.all([
      fetchAdminPosts({ page_size: 6 }),
      fetchAdminCategories(),
      fetchAdminTags(),
    ])

    recentPosts.value = postsResponse.data
    stats.value = {
      posts: postsResponse.meta.total,
      draftPosts: postsResponse.data.filter((post) => post.status === 'draft').length,
      categories: categoriesResponse.data.length,
      tags: tagsResponse.data.length,
    }
  } finally {
    loading.value = false
  }
}

onMounted(loadDashboard)
</script>

<template>
  <section class="dashboard fade-in">
    <div class="dashboard-header">
      <div>
        <p class="eyebrow">Overview</p>
        <h1 class="serif">控制台</h1>
      </div>
      <router-link class="button" to="/admin/posts/new">写新文章</router-link>
    </div>

    <div class="grid-three">
      <article class="card stat-card">
        <span>文章总数</span>
        <strong>{{ loading ? '...' : stats.posts }}</strong>
      </article>
      <article class="card stat-card">
        <span>草稿数量</span>
        <strong>{{ loading ? '...' : stats.draftPosts }}</strong>
      </article>
      <article class="card stat-card">
        <span>分类 / 标签</span>
        <strong>{{ loading ? '...' : `${stats.categories} / ${stats.tags}` }}</strong>
      </article>
    </div>

    <section class="card recent-card">
      <div class="section-row">
        <div>
          <p class="eyebrow">Recent</p>
          <h2 class="serif">最近编辑</h2>
        </div>
        <router-link class="button secondary" to="/admin/posts">查看全部</router-link>
      </div>

      <div class="recent-list">
        <article v-for="post in recentPosts" :key="post.id" class="recent-item">
          <div>
            <strong>{{ post.title }}</strong>
            <p class="muted">{{ post.status === 'published' ? '已发布' : '草稿' }} · {{ post.slug }}</p>
          </div>
          <router-link class="button ghost" :to="`/admin/posts/${post.id}`">编辑</router-link>
        </article>
      </div>
    </section>
  </section>
</template>

<style scoped>
.dashboard {
  display: grid;
  gap: 18px;
}

.dashboard-header,
.section-row {
  display: flex;
  align-items: end;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
}

.dashboard-header h1,
.section-row h2 {
  margin: 0;
}

.stat-card,
.recent-card {
  padding: 24px;
}

.stat-card span {
  color: var(--ink-soft);
}

.stat-card strong {
  display: block;
  margin-top: 16px;
  font-size: 2.5rem;
  line-height: 1;
  font-family: "Iowan Old Style", Georgia, serif;
}

.recent-list {
  display: grid;
  gap: 14px;
}

.recent-item {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  padding-top: 14px;
  border-top: 1px solid var(--line);
}

.recent-item:first-child {
  border-top: none;
  padding-top: 0;
}
</style>
