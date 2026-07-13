import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('@/pages/LoginPage.vue'),
      meta: { guest: true },
    },
    {
      path: '/',
      component: () => import('@/pages/MainLayout.vue'),
      meta: { requiresAuth: true },
      children: [
        { path: '', name: 'home', component: () => import('@/pages/HomePage.vue') },
        { path: 'timeline', name: 'timeline', component: () => import('@/pages/TimelinePage.vue') },
        { path: 'profile', name: 'profile', component: () => import('@/pages/ProfilePage.vue') },
      ],
    },
    {
      path: '/baby/new',
      name: 'baby-new',
      component: () => import('@/pages/BabyFormPage.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/baby/:id/edit',
      name: 'baby-edit',
      component: () => import('@/pages/BabyFormPage.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/record/:type',
      name: 'record-new',
      component: () => import('@/pages/RecordFormPage.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/record/:type/:id/edit',
      name: 'record-edit',
      component: () => import('@/pages/RecordFormPage.vue'),
      meta: { requiresAuth: true },
    },
    { path: '/:pathMatch(.*)*', redirect: '/' },
  ],
})

router.beforeEach((to, _from, next) => {
  const auth = useAuthStore()
  if (to.meta.requiresAuth && !auth.isLoggedIn) {
    next('/login')
  } else if (to.meta.guest && auth.isLoggedIn) {
    next('/')
  } else {
    next()
  }
})

export default router
