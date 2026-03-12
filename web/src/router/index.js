import { createRouter, createWebHistory } from 'vue-router'

import PublicLayout from '../layouts/PublicLayout.vue'
import AdminLayout from '../layouts/AdminLayout.vue'
import HomeView from '../views/HomeView.vue'
import PostDetailView from '../views/PostDetailView.vue'
import CategoryView from '../views/CategoryView.vue'
import TagView from '../views/TagView.vue'
import AboutView from '../views/AboutView.vue'
import UserLoginView from '../views/UserLoginView.vue'
import UserRegisterView from '../views/UserRegisterView.vue'
import AdminLoginView from '../views/AdminLoginView.vue'
import AdminDashboardView from '../views/AdminDashboardView.vue'
import AdminPostsView from '../views/AdminPostsView.vue'
import AdminPostEditorView from '../views/AdminPostEditorView.vue'
import AdminCategoriesView from '../views/AdminCategoriesView.vue'
import AdminTagsView from '../views/AdminTagsView.vue'
import AdminSiteSettingsView from '../views/AdminSiteSettingsView.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      component: PublicLayout,
      children: [
        { path: '', name: 'home', component: HomeView, meta: { title: '首页' } },
        { path: 'posts/:slug', name: 'post-detail', component: PostDetailView, meta: { title: '文章' } },
        { path: 'categories/:slug', name: 'category-detail', component: CategoryView, meta: { title: '分类' } },
        { path: 'tags/:slug', name: 'tag-detail', component: TagView, meta: { title: '标签' } },
        { path: 'about', name: 'about', component: AboutView, meta: { title: '关于' } },
        { path: 'login', name: 'user-login', component: UserLoginView, meta: { title: '用户登录', userGuestOnly: true } },
        { path: 'register', name: 'user-register', component: UserRegisterView, meta: { title: '用户注册', userGuestOnly: true } },
      ],
    },
    {
      path: '/admin/login',
      name: 'admin-login',
      component: AdminLoginView,
      meta: { title: '管理员登录', adminOnly: true },
    },
    {
      path: '/admin',
      component: AdminLayout,
      meta: { requiresAuth: true, adminOnly: true },
      children: [
        { path: '', name: 'admin-dashboard', component: AdminDashboardView, meta: { title: '控制台', requiresAuth: true, adminOnly: true } },
        { path: 'posts', name: 'admin-posts', component: AdminPostsView, meta: { title: '文章管理', requiresAuth: true, adminOnly: true } },
        { path: 'posts/new', name: 'admin-post-new', component: AdminPostEditorView, meta: { title: '新建文章', requiresAuth: true, adminOnly: true } },
        { path: 'posts/:id', name: 'admin-post-edit', component: AdminPostEditorView, meta: { title: '编辑文章', requiresAuth: true, adminOnly: true } },
        { path: 'categories', name: 'admin-categories', component: AdminCategoriesView, meta: { title: '分类管理', requiresAuth: true, adminOnly: true } },
        { path: 'tags', name: 'admin-tags', component: AdminTagsView, meta: { title: '标签管理', requiresAuth: true, adminOnly: true } },
        { path: 'settings', name: 'admin-settings', component: AdminSiteSettingsView, meta: { title: '站点设置', requiresAuth: true, adminOnly: true } },
      ],
    },
  ],
})

export default router
