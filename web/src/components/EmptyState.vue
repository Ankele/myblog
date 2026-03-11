<script setup>
defineProps({
  title: {
    type: String,
    required: true,
  },
  description: {
    type: String,
    required: true,
  },
})
</script>

<template>
  <div class="empty-state card">
    <p class="eyebrow">Empty</p>
    <h3 class="serif">{{ title }}</h3>
    <p class="muted">{{ description }}</p>
  </div>
</template>

<style scoped>
.empty-state {
  padding: 32px;
}

.empty-state h3 {
  margin: 0 0 10px;
  font-size: 1.4rem;
}

.empty-state p {
  margin: 0;
}
</style>
