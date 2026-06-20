<template>
  <el-container class="main-container">
    <el-aside width="220px" class="sidebar">
      <div class="logo">
        <el-icon :size="28" color="#fff"><OfficeBuilding /></el-icon>
        <span>粮库管理系统</span>
      </div>
      <el-menu
        :default-active="activeMenu"
        router
        background-color="#1f2937"
        text-color="#cbd5e1"
        active-text-color="#fff"
        class="menu"
      >
        <template v-for="item in menuItems" :key="item.path">
          <el-menu-item
            v-if="!item.roles || item.roles.includes(userStore.userRole as any)"
            :index="item.path"
          >
            <el-icon><component :is="item.icon" /></el-icon>
            <span>{{ item.title }}</span>
            <el-badge
              v-if="item.badgeKey && badgeCounts[item.badgeKey] > 0"
              :value="badgeCounts[item.badgeKey]"
              :max="99"
              class="menu-badge"
            />
          </el-menu-item>
        </template>
      </el-menu>
    </el-aside>

    <el-container>
      <el-header class="header">
        <div class="breadcrumb">
          <el-breadcrumb separator="/">
            <el-breadcrumb-item :to="{ path: '/dashboard' }">首页</el-breadcrumb-item>
            <el-breadcrumb-item v-if="currentTitle">{{ currentTitle }}</el-breadcrumb-item>
          </el-breadcrumb>
        </div>
        <div class="user-info">
          <el-tag :type="roleTagType" effect="dark" size="small">{{ roleLabel }}</el-tag>
          <el-dropdown>
            <span class="user-dropdown">
              <el-avatar :size="32" class="avatar">
                {{ userStore.userName?.charAt(0) || 'U' }}
              </el-avatar>
              <span class="username">{{ userStore.userName }}</span>
              <el-icon><ArrowDown /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item disabled>
                  <el-icon><User /></el-icon>
                  {{ userStore.user?.username }}
                </el-dropdown-item>
                <el-dropdown-item divided @click="handleLogout">
                  <el-icon><SwitchButton /></el-icon>
                  退出登录
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>

      <el-main class="main-content">
        <router-view v-slot="{ Component }">
          <transition name="fade" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessageBox } from 'element-plus'
import {
  DataAnalysis,
  OfficeBuilding,
  Files,
  TrendCharts,
  Warning,
  RefreshRight,
  Reading,
  Bell,
  ArrowDown,
  User,
  SwitchButton
} from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import { RoleLabels } from '@/types'
import type { UserRole } from '@/types'
import { suggestionApi, fumigationApi } from '@/api'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const menuItems = [
  { path: '/dashboard', title: '数据看板', icon: 'DataAnalysis' },
  { path: '/granaries', title: '仓房管理', icon: 'OfficeBuilding', roles: ['admin', 'keeper'] as UserRole[], badgeKey: 'abnormal' },
  { path: '/grain-conditions', title: '粮情录入', icon: 'Files', roles: ['admin', 'keeper'] as UserRole[] },
  { path: '/temperature-monitor', title: '粮温监控', icon: 'TrendCharts' },
  { path: '/fumigation', title: '熏蒸管理', icon: 'Warning', badgeKey: 'pendingFumigation' },
  { path: '/unseal', title: '通风解封', icon: 'RefreshRight', roles: ['admin', 'duty_officer', 'keeper'] as UserRole[] },
  { path: '/gas-detections', title: '气体检测', icon: 'Reading', roles: ['admin', 'duty_officer'] as UserRole[] },
  { path: '/turnover-suggestions', title: '翻仓建议', icon: 'Bell', badgeKey: 'suggestions' }
]

const activeMenu = computed(() => {
  const path = route.path
  for (const item of menuItems) {
    if (path.startsWith(item.path) || (item.path === '/dashboard' && path === '/')) {
      return item.path
    }
  }
  return '/dashboard'
})

const currentTitle = computed(() => {
  const title = route.meta.title as string
  return title || ''
})

const roleLabel = computed(() => {
  if (!userStore.user?.role) return ''
  return RoleLabels[userStore.user.role]
})

const roleTagType = computed(() => {
  const map: Record<UserRole, any> = {
    admin: 'danger',
    keeper: 'success',
    safety_officer: 'warning',
    duty_officer: 'info'
  }
  return userStore.user?.role ? map[userStore.user.role] : 'info'
})

const badgeCounts = ref<Record<string, number>>({
  suggestions: 0,
  pendingFumigation: 0,
  abnormal: 0
})

const loadBadgeCounts = async () => {
  try {
    const [suggestions, fumigation] = await Promise.all([
      suggestionApi.list({ status: 'pending' }).catch(() => [] as any),
      fumigationApi.listPlans({ status: 'pending_approval' }).catch(() => [] as any)
    ])
    badgeCounts.value.suggestions = Array.isArray(suggestions) ? suggestions.length : 0
    badgeCounts.value.pendingFumigation = Array.isArray(fumigation) ? fumigation.length : 0
  } catch {}
}

const handleLogout = async () => {
  try {
    await ElMessageBox.confirm('确定要退出登录吗？', '提示', {
      type: 'warning'
    })
    userStore.logout()
    router.push('/login')
  } catch {}
}

onMounted(() => {
  loadBadgeCounts()
})
</script>

<style scoped>
.main-container {
  height: 100vh;
}

.sidebar {
  background: #1f2937;
  transition: all 0.3s;
  overflow-y: auto;
}

.logo {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  color: #fff;
  font-size: 16px;
  font-weight: 600;
  border-bottom: 1px solid #374151;
}

.menu {
  border-right: none;
}

.menu :deep(.el-menu-item) {
  height: 50px;
  line-height: 50px;
}

.menu-badge {
  margin-left: auto;
  margin-right: 8px;
}

.header {
  background: #fff;
  border-bottom: 1px solid #e4e7ed;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  height: 60px;
}

.breadcrumb {
  font-size: 14px;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 16px;
}

.user-dropdown {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 6px;
  transition: background 0.2s;
}

.user-dropdown:hover {
  background: #f5f7fa;
}

.username {
  font-size: 14px;
  color: #303133;
}

.avatar {
  background: linear-gradient(135deg, #667eea, #764ba2);
  font-weight: 600;
}

.main-content {
  background: #f5f7fa;
  padding: 20px;
  overflow-y: auto;
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
