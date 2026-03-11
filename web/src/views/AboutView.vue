<script setup>
import { computed } from 'vue'

import { useSiteStore } from '../stores/site'

const siteStore = useSiteStore()

const socialLinks = computed(() =>
  [
    { label: 'GitHub', value: siteStore.site.github_url },
    { label: 'Twitter', value: siteStore.site.twitter_url },
    { label: 'LinkedIn', value: siteStore.site.linkedin_url },
    { label: 'Email', value: siteStore.site.email ? `mailto:${siteStore.site.email}` : '' },
  ].filter((item) => item.value),
)
</script>

<template>
  <section class="page-shell about-view">
    <div class="about-grid">
      <article class="about-copy">
        <p class="eyebrow">About</p>
        <h1 class="section-title serif">{{ siteStore.site.about_title || '关于我' }}</h1>
        <p class="muted intro">{{ siteStore.site.description }}</p>
        <div class="article-html" v-html="siteStore.site.about_content"></div>
      </article>

      <aside class="card about-panel">
        <img v-if="siteStore.site.avatar" class="avatar" :src="siteStore.site.avatar" alt="avatar" />
        <div class="stack">
          <div>
            <p class="eyebrow">Elsewhere</p>
            <div class="social-links">
              <a v-for="item in socialLinks" :key="item.label" class="chip" :href="item.value" target="_blank" rel="noreferrer">
                {{ item.label }}
              </a>
            </div>
          </div>
          <router-link class="button secondary" to="/">回到文章列表</router-link>
        </div>
      </aside>
    </div>
  </section>
</template>

<style scoped>
.about-view {
  padding-bottom: 48px;
}

.about-grid {
  display: grid;
  grid-template-columns: minmax(0, 1.4fr) minmax(280px, 0.8fr);
  gap: 24px;
}

.about-copy {
  padding-top: 24px;
}

.intro {
  margin-top: 18px;
}

.about-panel {
  align-self: start;
  padding: 28px;
}

.avatar {
  width: 100%;
  aspect-ratio: 1;
  object-fit: cover;
  border-radius: 22px;
  margin-bottom: 24px;
}

.social-links {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

@media (max-width: 900px) {
  .about-grid {
    grid-template-columns: 1fr;
  }
}
</style>
