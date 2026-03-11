import { request, toQuery } from './http'

export function login(payload) {
  return request('/api/admin/login', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}

export function logout() {
  return request('/api/admin/logout', {
    method: 'POST',
  })
}

export function fetchMe() {
  return request('/api/admin/me')
}

export function fetchAdminPosts(params = {}) {
  return request(`/api/admin/posts${toQuery(params)}`)
}

export function fetchAdminPost(id) {
  return request(`/api/admin/posts/${id}`)
}

export function savePost(payload, id) {
  return request(id ? `/api/admin/posts/${id}` : '/api/admin/posts', {
    method: id ? 'PUT' : 'POST',
    body: JSON.stringify(payload),
  })
}

export function deletePost(id) {
  return request(`/api/admin/posts/${id}`, {
    method: 'DELETE',
  })
}

export function previewMarkdown(markdownContent) {
  return request('/api/admin/posts/preview', {
    method: 'POST',
    body: JSON.stringify({ markdown_content: markdownContent }),
  })
}

export function fetchAdminCategories() {
  return request('/api/admin/categories')
}

export function saveCategory(payload, id) {
  return request(id ? `/api/admin/categories/${id}` : '/api/admin/categories', {
    method: id ? 'PUT' : 'POST',
    body: JSON.stringify(payload),
  })
}

export function deleteCategory(id) {
  return request(`/api/admin/categories/${id}`, {
    method: 'DELETE',
  })
}

export function fetchAdminTags() {
  return request('/api/admin/tags')
}

export function saveTag(payload, id) {
  return request(id ? `/api/admin/tags/${id}` : '/api/admin/tags', {
    method: id ? 'PUT' : 'POST',
    body: JSON.stringify(payload),
  })
}

export function deleteTag(id) {
  return request(`/api/admin/tags/${id}`, {
    method: 'DELETE',
  })
}

export function fetchAdminSite() {
  return request('/api/admin/site')
}

export function updateSite(payload) {
  return request('/api/admin/site', {
    method: 'PUT',
    body: JSON.stringify(payload),
  })
}

export function uploadFile(file) {
  const form = new FormData()
  form.append('file', file)
  return request('/api/admin/upload', {
    method: 'POST',
    body: form,
  })
}

export function importMarkdownPost(file, options = {}) {
  const form = new FormData()
  form.append('file', file)
  if (options.status) {
    form.append('status', options.status)
  }
  if (options.category_id !== undefined && options.category_id !== null && options.category_id !== '') {
    form.append('category_id', options.category_id)
  }
  if (Array.isArray(options.tag_ids) && options.tag_ids.length > 0) {
    form.append('tag_ids', options.tag_ids.join(','))
  }
  if (options.cover_image) {
    form.append('cover_image', options.cover_image)
  }

  return request('/api/admin/posts/import-markdown', {
    method: 'POST',
    body: form,
  })
}
