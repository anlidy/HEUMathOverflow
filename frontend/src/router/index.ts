import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
    history: createWebHistory(import.meta.env.BASE_URL),
    routes: [
        { path: '/', component: () => import('@/views/HomeView.vue') },
        { path: '/login', component: () => import('@/views/LoginView.vue') },
        { path: '/register', component: () => import('@/views/RegisterView.vue') },
        { path: '/profile/me', component: () => import('@/views/ProfileView.vue') },
        { path: '/settings', component: () => import('@/views/SettingsView.vue') },
        { path: '/posts/:id', component: () => import('@/views/PostDetailView.vue') },
        { path: '/editor/create', component: () => import('@/views/CreatePostView.vue') },
    ],
})

export default router
