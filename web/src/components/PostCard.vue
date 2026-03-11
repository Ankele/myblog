<script setup>
defineProps({
  post: {
    type: Object,
    required: true,
  },
})
</script>

<template>
  <article class="post-card card">
    <img v-if="post.cover_image" class="cover" :src="post.cover_image" :alt="post.title" />

    <div class="content">
      <div class="meta">
        <span>{{ post.published_at ? new Date(post.published_at).toLocaleDateString('zh-CN') : '未发布' }}</span>
        <router-link v-if="post.category?.slug" class="chip" :to="`/categories/${post.category.slug}`">{{ post.category.name }}</router-link>
      </div>

      <router-link class="title serif" :to="`/posts/${post.slug}`">{{ post.title }}</router-link>
      <p class="summary">{{ post.summary || '这篇文章还没有摘要。' }}</p>

      <div class="footer">
        <div class="tags">
          <router-link v-for="tag in post.tags || []" :key="tag.id" class="chip" :to="`/tags/${tag.slug}`">
            # {{ tag.name }}
          </router-link>
        </div>
        <router-link class="readmore" :to="`/posts/${post.slug}`">继续阅读</router-link>
      </div>
    </div>
  </article>
</template>

<style scoped>
.post-card {
  overflow: hidden;
}

.cover {
  width: 100%;
  aspect-ratio: 16 / 8;
  object-fit: cover;
}

.content {
  padding: 24px;
  display: grid;
  gap: 16px;
}

.meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  color: var(--ink-soft);
  font-size: 0.9rem;
}

.title {
  font-size: clamp(1.6rem, 3vw, 2.4rem);
  line-height: 1.02;
  letter-spacing: -0.04em;
}

.summary {
  margin: 0;
  color: var(--ink-soft);
  line-height: 1.75;
}

.footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
}

.tags {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.readmore {
  color: var(--accent);
  font-weight: 600;
}
</style>
