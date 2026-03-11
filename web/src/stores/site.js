import { defineStore } from 'pinia'

import { fetchCategories, fetchSite, fetchTags } from '../api/blog'

export const useSiteStore = defineStore('site', {
  state: () => ({
    site: {
      site_name: 'MyBlog',
      tagline: '',
      description: '',
      footer_text: '',
      about_title: '',
      about_content: '',
    },
    categories: [],
    tags: [],
    loaded: false,
  }),
  actions: {
    async ensureSite() {
      if (this.loaded) {
        return
      }
      const [siteResponse, categoriesResponse, tagsResponse] = await Promise.all([
        fetchSite(),
        fetchCategories(),
        fetchTags(),
      ])
      this.site = siteResponse.data
      this.categories = categoriesResponse.data
      this.tags = tagsResponse.data
      this.loaded = true
    },
    async refreshTaxonomies() {
      const [categoriesResponse, tagsResponse] = await Promise.all([
        fetchCategories(),
        fetchTags(),
      ])
      this.categories = categoriesResponse.data
      this.tags = tagsResponse.data
    },
  },
})
