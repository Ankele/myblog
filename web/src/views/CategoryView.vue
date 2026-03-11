<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'

import { fetchPosts } from '../api/blog'
import EmptyState from '../components/EmptyState.vue'
import PostCard from '../components/PostCard.vue'
import { useSiteStore } from '../stores/site'

const route = useRoute()
const siteStore = useSiteStore()
const loading = ref(true)
const posts = ref([])

const currentCategory = computed(() => siteStore.categories.find((item) => item.slug === route.params.slug))

async function loadCategoryPosts() {
  loading.value = true
  try {
    const response = await fetchPosts({ category: route.params.slug, page_size: 20 })
    posts.value = response.data
  } finally {
    loading.value = false
  }
}

watch(() => route.params.slug, loadCategoryPosts)
onMounted(loadCategoryPosts)
</script>

<template>
  <section class="page-shell taxonomy-view">
    <div class="taxonomy-header">
      <p class="eyebrow">Category</p>
      <h1 class="section-title serif">{{ currentCategory?.name || route.params.slug }}</h1>
      <p class="muted">{{ currentCategory?.description || '按分类筛选的文章列表。' }}</p>
    </div>

    <EmptyState
      v-if="!loading && posts.length === 0"
      title="这个分类下还没有文章"
      description="发布并绑定文章后，这里会自动更新。"
    />
    <div v-else class="stack">
      <PostCard v-for="post in posts" :key="post.id" :post="post" />
    </div>
  </section>
</template>

<style scoped>
.taxonomy-view {
  padding-bottom: 44px;
}

.taxonomy-header {
  max-width: 720px;
  margin: 24px 0 30px;
}

.taxonomy-header p:last-child {
  margin-top: 16px;
}
</style>
