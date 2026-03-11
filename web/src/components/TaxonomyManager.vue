<script setup>
import { reactive, watch } from 'vue'

const props = defineProps({
  title: {
    type: String,
    required: true,
  },
  items: {
    type: Array,
    default: () => [],
  },
  busy: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['save', 'remove'])

const form = reactive({
  id: null,
  name: '',
  slug: '',
  description: '',
})

function reset() {
  form.id = null
  form.name = ''
  form.slug = ''
  form.description = ''
}

function edit(item) {
  form.id = item.id
  form.name = item.name
  form.slug = item.slug
  form.description = item.description
}

async function submit() {
  await emit('save', { ...form })
  reset()
}

watch(
  () => props.items,
  () => {
    if (!props.items.some((item) => item.id === form.id)) {
      reset()
    }
  },
)
</script>

<template>
  <div class="taxonomy-manager grid-two">
    <section class="card block">
      <p class="eyebrow">{{ title }}</p>
      <h2 class="serif panel-title">新增或编辑</h2>
      <div class="stack">
        <input v-model="form.name" class="input" placeholder="名称" />
        <input v-model="form.slug" class="input" placeholder="Slug（可留空自动生成）" />
        <textarea v-model="form.description" class="textarea" placeholder="简介"></textarea>
        <div class="actions">
          <button class="button" :disabled="busy" type="button" @click="submit">
            {{ form.id ? '保存修改' : '创建' }}
          </button>
          <button v-if="form.id" class="button secondary" type="button" @click="reset">取消编辑</button>
        </div>
      </div>
    </section>

    <section class="card block">
      <p class="eyebrow">List</p>
      <h2 class="serif panel-title">当前条目</h2>
      <div class="list">
        <article v-for="item in items" :key="item.id" class="item">
          <div>
            <strong>{{ item.name }}</strong>
            <p class="muted">{{ item.slug }}</p>
            <p class="muted">{{ item.description || '暂无描述' }}</p>
          </div>
          <div class="item-actions">
            <button class="button ghost" type="button" @click="edit(item)">编辑</button>
            <button class="button danger" :disabled="busy" type="button" @click="$emit('remove', item)">删除</button>
          </div>
        </article>
      </div>
    </section>
  </div>
</template>

<style scoped>
.block {
  padding: 28px;
}

.panel-title {
  margin: 0 0 22px;
  font-size: 1.9rem;
}

.actions,
.item-actions {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.list {
  display: grid;
  gap: 14px;
}

.item {
  display: flex;
  justify-content: space-between;
  gap: 18px;
  padding: 16px 0;
  border-top: 1px solid var(--line);
}

.item:first-child {
  border-top: none;
  padding-top: 0;
}

.item p {
  margin: 6px 0 0;
}

@media (max-width: 720px) {
  .item {
    flex-direction: column;
  }
}
</style>
