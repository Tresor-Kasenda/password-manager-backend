<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useAuthStore } from './store/authStore'

const { isAuthenticated, user, logout: storeLogout } = useAuthStore()
const router = useRouter()

const logout = () => {
  storeLogout()
  router.push('/login')
}
</script>

<template>
  <div class="min-h-screen bg-slate-50">
    <nav v-if="isAuthenticated" class="bg-white border-b border-slate-200 px-6 py-4 flex justify-between items-center shadow-sm">
      <div class="flex items-center space-x-2">
        <div class="w-8 h-8 bg-indigo-600 rounded-lg flex items-center justify-center">
          <span class="text-white font-bold text-xl">T</span>
        </div>
        <span class="text-xl font-bold bg-gradient-to-r from-indigo-600 to-violet-600 bg-clip-text text-transparent">
          Tresor
        </span>
      </div>
      
      <div class="flex items-center space-x-6">
        <span class="text-sm font-medium text-slate-600">Welcome, {{ user?.email || 'User' }}</span>
        <button 
          @click="logout" 
          class="bg-white hover:bg-slate-50 text-slate-700 px-4 py-2 rounded-lg text-sm font-semibold border border-slate-200 transition shadow-sm"
        >
          Logout
        </button>
      </div>
    </nav>

    <main>
      <router-view></router-view>
    </main>
  </div>
</template>

<style>
/* Global styles if needed */
</style>
