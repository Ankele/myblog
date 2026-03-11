<script setup>
import { onMounted, ref } from 'vue'

import { deleteCategory, fetchAdminCategories, saveCategory } from '../api/admin'
import TaxonomyManager from '../components/TaxonomyManager.vue'

const items = ref([])
const busy = ref(false)

async function loadItems() {
  const response = await fetchAdminCategories()
  items.value = response.data
}

async function handleSave(payload) {
  busy.value = true
  try {
    await saveCategory(payload, payload.id)
    await loadItems()
  } finally {
    busy.value = false
  }
}

async function handleRemove(item) {
  if (!window.confirm(`确认删除分类“${item.name}”吗？文章会变成未分类。`)) {
    return
  }
  busy.value = true
  try {
    await deleteCategory(item.id)
    await loadItems()
  } finally {
    busy.value = false
  }
}

onMounted(loadItems)
</script>

<template>
  <section class="stack fade-in">
    <div>
      <p class="eyebrow">Categories</p>
      <h1 class="serif">分类管理</h1>
    </div>
    <TaxonomyManager title="分类" :items="items" :busy="busy" @save="handleSave" @remove="handleRemove" />
  </section>
</template>

<style scoped>
h1 {
  margin: 0;
}
</style>
