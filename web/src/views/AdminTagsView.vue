<script setup>
import { onMounted, ref } from 'vue'

import { deleteTag, fetchAdminTags, saveTag } from '../api/admin'
import TaxonomyManager from '../components/TaxonomyManager.vue'

const items = ref([])
const busy = ref(false)

async function loadItems() {
  const response = await fetchAdminTags()
  items.value = response.data
}

async function handleSave(payload) {
  busy.value = true
  try {
    await saveTag(payload, payload.id)
    await loadItems()
  } finally {
    busy.value = false
  }
}

async function handleRemove(item) {
  if (!window.confirm(`确认删除标签“${item.name}”吗？`)) {
    return
  }
  busy.value = true
  try {
    await deleteTag(item.id)
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
      <p class="eyebrow">Tags</p>
      <h1 class="serif">标签管理</h1>
    </div>
    <TaxonomyManager title="标签" :items="items" :busy="busy" @save="handleSave" @remove="handleRemove" />
  </section>
</template>

<style scoped>
h1 {
  margin: 0;
}
</style>
