import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { User, UserRole } from '@/types'
import { authApi } from '@/api'

export const useUserStore = defineStore('user', () => {
  const token = ref<string>(localStorage.getItem('token') || '')
  const user = ref<User | null>(null)

  const isLoggedIn = computed(() => !!token.value)
  const userRole = computed<UserRole | ''>(() => user.value?.role || '')
  const userName = computed(() => user.value?.full_name || user.value?.username || '')

  const hasRole = (...roles: UserRole[]) => {
    if (!user.value) return false
    return roles.includes(user.value.role)
  }

  const login = async (username: string, password: string) => {
    const res = await authApi.login({ username, password })
    token.value = res.token
    user.value = res.user
    localStorage.setItem('token', res.token)
    localStorage.setItem('user', JSON.stringify(res.user))
    return res
  }

  const loadStoredUser = () => {
    const stored = localStorage.getItem('user')
    if (stored) {
      try {
        user.value = JSON.parse(stored)
      } catch {}
    }
  }

  const logout = () => {
    token.value = ''
    user.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('user')
  }

  const fetchCurrentUser = async () => {
    try {
      user.value = await authApi.getCurrentUser()
      localStorage.setItem('user', JSON.stringify(user.value))
    } catch {}
  }

  return {
    token,
    user,
    isLoggedIn,
    userRole,
    userName,
    hasRole,
    login,
    logout,
    loadStoredUser,
    fetchCurrentUser
  }
})
