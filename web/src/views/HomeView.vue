<script setup>
import { onMounted, ref } from 'vue'

import { fetchPosts } from '../api/blog'
import PostCard from '../components/PostCard.vue'
import EmptyState from '../components/EmptyState.vue'
import { useSiteStore } from '../stores/site'

const siteStore = useSiteStore()
const loading = ref(true)
const posts = ref([])
const total = ref(0)

async function loadPosts() {
  loading.value = true
  try {
    const response = await fetchPosts({ page_size: 12 })
    posts.value = response.data
    total.value = response.meta.total
  } finally {
    loading.value = false
  }
}

onMounted(loadPosts)
</script>

<template>
  <section class="page-shell home-hero fade-in">
    <div class="hero-copy">
      <p class="eyebrow">Personal Blog</p>
      <h1 class="section-title serif">{{ siteStore.site.site_name }}</h1>
      <p class="hero-text">
        {{ siteStore.site.description || siteStore.site.tagline }}
      </p>
      <div class="hero-actions">
        <router-link class="button" to="/about">认识作者</router-link>
        <router-link class="button secondary" to="/admin">进入后台</router-link>
      </div>
    </div>

    <aside class="hero-panel card">
      <p class="eyebrow">Browse</p>
      <div class="stack">
        <div>
          <strong>{{ total }}</strong>
          <p class="muted">已发布文章</p>
        </div>
        <div class="chip-group">
          <router-link
            v-for="category in siteStore.categories"
            :key="category.id"
            class="chip"
            :to="`/categories/${category.slug}`"
          >
            {{ category.name }}
          </router-link>
        </div>
      </div>
    </aside>
  </section>

  <section class="page-shell home-list">
    <div class="list-header">
      <div>
        <p class="eyebrow">Latest Writing</p>
        <h2 class="serif list-title">最近更新</h2>
      </div>
      <div class="chip-group">
        <router-link
          v-for="tag in siteStore.tags.slice(0, 6)"
          :key="tag.id"
          class="chip"
          :to="`/tags/${tag.slug}`"
        >
          # {{ tag.name }}
        </router-link>
      </div>
    </div>

    <div v-if="loading" class="loading-grid">
      <div v-for="index in 3" :key="index" class="card skeleton"></div>
    </div>
    <EmptyState
      v-else-if="posts.length === 0"
      title="还没有已发布文章"
      description="登录后台发布第一篇文章后，这里会自动展示。"
    />
    <div v-else class="stack">
      <PostCard v-for="post in posts" :key="post.id" :post="post" />
    </div>
  </section>
</template>

<style scoped>
.home-hero {
  display: grid;
  grid-template-columns: minmax(0, 1.6fr) minmax(280px, 0.8fr);
  gap: 24px;
  padding: 24px 0 40px;
}

.hero-copy {
  padding: 26px 0 10px;
}

.hero-text {
  max-width: 680px;
  margin: 20px 0 0;
  font-size: 1.1rem;
  line-height: 1.9;
  color: var(--ink-soft);
}

.hero-actions {
  display: flex;
  gap: 12px;
  margin-top: 26px;
  flex-wrap: wrap;
}

.hero-panel {
  padding: 28px;
}

.hero-panel strong {
  font-size: 3.4rem;
  line-height: 1;
  font-family: "Iowan Old Style", "Palatino Linotype", Georgia, serif;
}

.hero-panel p {
  margin: 8px 0 0;
}

.chip-group {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.home-list {
  padding-bottom: 40px;
}

.list-header {
  display: flex;
  align-items: end;
  justify-content: space-between;
  gap: 18px;
  margin-bottom: 24px;
  flex-wrap: wrap;
}

.list-title {
  margin: 0;
  font-size: 2rem;
}

.loading-grid {
  display: grid;
  gap: 18px;
}

.skeleton {
  height: 240px;
  background: linear-gradient(90deg, rgba(255,255,255,0.6), rgba(255,255,255,0.18), rgba(255,255,255,0.6));
  background-size: 200% 100%;
  animation: shimmer 1.2s linear infinite;
}

@keyframes shimmer {
  to {
    background-position: -200% 0;
  }
}

@media (max-width: 900px) {
  .home-hero {
    grid-template-columns: 1fr;
  }
}
</style>
