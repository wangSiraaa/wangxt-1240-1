import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router'
import { useUserStore } from '@/stores/user'
import type { UserRole } from '@/types'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/',
    component: () => import('@/layouts/MainLayout.vue'),
    redirect: '/dashboard',
    meta: { requiresAuth: true },
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/Dashboard.vue'),
        meta: { title: '数据看板', icon: 'DataAnalysis' }
      },
      {
        path: 'granaries',
        name: 'Granaries',
        component: () => import('@/views/Granaries.vue'),
        meta: { title: '仓房管理', icon: 'OfficeBuilding', roles: ['admin', 'keeper'] as UserRole[] }
      },
      {
        path: 'granaries/:id',
        name: 'GranaryDetail',
        component: () => import('@/views/GranaryDetail.vue'),
        meta: { title: '仓房详情', hidden: true }
      },
      {
        path: 'grain-conditions',
        name: 'GrainConditions',
        component: () => import('@/views/GrainConditions.vue'),
        meta: { title: '粮情录入', icon: 'Files', roles: ['admin', 'keeper'] as UserRole[] }
      },
      {
        path: 'temperature-monitor',
        name: 'TemperatureMonitor',
        component: () => import('@/views/TemperatureMonitor.vue'),
        meta: { title: '粮温监控', icon: 'TrendCharts' }
      },
      {
        path: 'fumigation',
        name: 'Fumigation',
        component: () => import('@/views/Fumigation.vue'),
        meta: { title: '熏蒸管理', icon: 'Warning' }
      },
      {
        path: 'fumigation/:id',
        name: 'FumigationDetail',
        component: () => import('@/views/FumigationDetail.vue'),
        meta: { title: '熏蒸方案详情', hidden: true }
      },
      {
        path: 'unseal',
        name: 'Unseal',
        component: () => import('@/views/Unseal.vue'),
        meta: { title: '通风解封', icon: 'Wind', roles: ['admin', 'duty_officer', 'keeper'] as UserRole[] }
      },
      {
        path: 'gas-detections',
        name: 'GasDetections',
        component: () => import('@/views/GasDetections.vue'),
        meta: { title: '气体检测', icon: 'Reading', roles: ['admin', 'duty_officer'] as UserRole[] }
      },
      {
        path: 'turnover-suggestions',
        name: 'TurnoverSuggestions',
        component: () => import('@/views/TurnoverSuggestions.vue'),
        meta: { title: '翻仓建议', icon: 'Bell' }
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, from, next) => {
  const userStore = useUserStore()

  if (to.meta.requiresAuth !== false && !userStore.isLoggedIn) {
    next({ path: '/login', query: { redirect: to.fullPath } })
    return
  }

  if (to.path === '/login' && userStore.isLoggedIn) {
    next('/')
    return
  }

  if (to.meta.roles) {
    const roles = to.meta.roles as UserRole[]
    if (!userStore.hasRole(...roles)) {
      next('/dashboard')
      return
    }
  }

  next()
})

export default router
