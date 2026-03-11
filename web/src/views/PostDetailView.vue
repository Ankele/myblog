<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'

import { fetchPost } from '../api/blog'
import EmptyState from '../components/EmptyState.vue'
import { useUserAuthStore } from '../stores/userAuth'

const route = useRoute()
const userAuthStore = useUserAuthStore()
const loading = ref(true)
const error = ref('')
const post = ref(null)

const formattedDate = computed(() => {
  if (!post.value?.published_at) {
    return '未发布'
  }
  return new Date(post.value.published_at).toLocaleDateString('zh-CN')
})

async function loadPost() {
  loading.value = true
  error.value = ''
  try {
    const response = await fetchPost(route.params.slug)
    post.value = response.data
  } catch (err) {
    error.value = err.message
    post.value = null
  } finally {
    loading.value = false
  }
}

watch(() => route.params.slug, loadPost)
onMounted(loadPost)
</script>

<template>
  <section class="page-shell post-detail">
    <EmptyState
      v-if="error"
      title="文章不存在"
      :description="error"
    />

    <template v-else-if="post">
      <article class="post-hero">
        <p class="eyebrow">{{ formattedDate }}</p>
        <h1 class="section-title serif">{{ post.title }}</h1>
        <p class="deck muted">{{ post.summary }}</p>
        <div class="post-taxonomy">
          <router-link v-if="post.category?.slug" class="chip" :to="`/categories/${post.category.slug}`">
            {{ post.category.name }}
          </router-link>
          <router-link v-for="tag in post.tags" :key="tag.id" class="chip" :to="`/tags/${tag.slug}`">
            # {{ tag.name }}
          </router-link>
        </div>
      </article>

      <img v-if="post.cover_image" class="hero-cover card" :src="post.cover_image" :alt="post.title" />

      <article class="content-wrap card">
        <div class="article-html" v-html="post.html_content"></div>
      </article>

      <section class="interaction-panel card">
        <p v-if="userAuthStore.isAuthenticated" class="muted">
          已登录为 {{ userAuthStore.user?.username }}，可参与评论和点赞。
        </p>
        <p v-else class="muted">
          <router-link class="inline-link" :to="{ name: 'user-login', query: { redirect: route.fullPath } }">登录</router-link>
          后可参与评论和点赞。
        </p>
      </section>
    </template>

    <div v-else-if="loading" class="card loading-panel"></div>
  </section>
</template>

<style scoped>
.post-detail {
  padding-bottom: 50px;
}

.post-hero {
  max-width: 820px;
  margin: 24px auto 26px;
}

.deck {
  margin: 18px 0 0;
  font-size: 1.08rem;
  line-height: 1.9;
}

.post-taxonomy {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
  margin-top: 26px;
}

.hero-cover {
  width: min(980px, 100%);
  margin: 0 auto 28px;
  aspect-ratio: 16 / 8;
  object-fit: cover;
  overflow: hidden;
}

.content-wrap {
  width: min(820px, 100%);
  margin: 0 auto;
  padding: 32px;
}

.loading-panel {
  width: min(820px, 100%);
  height: 320px;
  margin: 30px auto 0;
}

.interaction-panel {
  width: min(820px, 100%);
  margin: 16px auto 0;
  padding: 18px 22px;
}

.interaction-panel p {
  margin: 0;
}

.inline-link {
  color: var(--accent);
}
</style>
