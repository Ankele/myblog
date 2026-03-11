import { request, toQuery } from './http'

export function fetchPosts(params = {}) {
  return request(`/api/posts${toQuery(params)}`)
}

export function fetchPost(slug) {
  return request(`/api/posts/${slug}`)
}

export function fetchCategories() {
  return request('/api/categories')
}

export function fetchTags() {
  return request('/api/tags')
}

export function fetchSite() {
  return request('/api/site')
}
