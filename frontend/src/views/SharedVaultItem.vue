<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { sharingApi } from '../api/healthAndSharing'

const route = useRoute()
const token = route.params.token as string
const sharePassword = ref('')
const loading = ref(false)
const error = ref('')
const data = ref<any>(null)
const permissions = ref<any>(null)

const fetchShared = async () => {
  loading.value = true
  error.value = ''
  try {
    // We need to use a custom apiClient call because this might be accessed without login
    // But the backend router.go shows 'shared/:token' is PROTECTED.
    // Wait, let's check router.go again.
    // ...
    // shared := protected.Group("/shared")
    // {
    //   shared.GET("/:token", r.sharingHandler.GetSharedPassword)
    // }
    // Yes, it is protected. So the user MUST be logged in to view a shared password?
    // That's fine for now, usually sharing is within the team.
    
    // apiClient in my case adds the token automatically.
    const res = await sharingApi.getSharedPassword(token, sharePassword.value) as any
    data.value = res.data
    permissions.value = res.permissions
  } catch (err: any) {
    if (err.message?.includes('password required')) {
      error.value = 'Password required'
    } else {
      error.value = err.message || 'Failed to access share'
    }
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchShared()
})
</script>

<template>
  <div class="min-h-[80vh] flex items-center justify-center p-4">
    <div class="max-w-md w-full bg-white rounded-3xl border border-slate-100 shadow-2xl p-8">
      <div v-if="loading" class="text-center py-12">
        <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600 mx-auto"></div>
        <p class="mt-4 text-slate-500 font-bold">Accessing Secure Share...</p>
      </div>

      <div v-else-if="error === 'Password required'" class="space-y-6">
          <div class="text-center">
              <div class="w-16 h-16 bg-indigo-50 text-indigo-600 rounded-full flex items-center justify-center mx-auto mb-4">
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                  </svg>
              </div>
              <h2 class="text-2xl font-black text-slate-900">Protected Share</h2>
              <p class="text-sm text-slate-500 mt-1">This entry requires a password to view.</p>
          </div>
          
          <div class="space-y-4">
              <input 
                v-model="sharePassword" 
                type="password" 
                class="w-full px-6 py-4 bg-slate-50 border border-slate-200 rounded-2xl focus:ring-4 focus:ring-indigo-500/10 outline-none transition" 
                placeholder="Enter share password"
              />
              <button 
                @click="fetchShared"
                class="w-full bg-indigo-600 hover:bg-indigo-700 text-white py-4 rounded-2xl font-black transition shadow-xl shadow-indigo-200"
              >
                Unlock Entry
              </button>
          </div>
      </div>

      <div v-else-if="error" class="text-center py-8">
          <div class="w-16 h-16 bg-red-50 text-red-600 rounded-full flex items-center justify-center mx-auto mb-4">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
              </svg>
          </div>
          <h2 class="text-xl font-bold text-slate-900">Access Error</h2>
          <p class="text-sm text-slate-500 mt-2">{{ error }}</p>
          <router-link to="/" class="mt-6 inline-block text-indigo-600 font-bold hover:underline">Back to Safety</router-link>
      </div>

      <div v-else-if="data" class="space-y-8 animate-in zoom-in-95 duration-300">
          <div class="text-center">
              <div class="w-20 h-20 bg-indigo-600 text-white rounded-3xl flex items-center justify-center mx-auto mb-6 shadow-xl shadow-indigo-200">
                  <span class="text-3xl font-black">{{ data.title[0].toUpperCase() }}</span>
              </div>
              <h2 class="text-2xl font-black text-slate-900">{{ data.title }}</h2>
              <p class="text-sm text-slate-500 mt-1">{{ data.website || 'No website associated' }}</p>
          </div>

          <div class="space-y-4">
              <div class="p-4 bg-slate-50 rounded-2xl border border-slate-100">
                  <p class="text-[10px] font-black text-slate-400 uppercase tracking-widest mb-1">Username</p>
                  <p class="font-bold text-slate-900">{{ data.username || 'Not set' }}</p>
              </div>
              <div class="p-4 bg-indigo-50 rounded-2xl border border-indigo-100 flex justify-between items-center group">
                  <div>
                    <p class="text-[10px] font-black text-indigo-400 uppercase tracking-widest mb-1">Password</p>
                    <p class="font-black text-indigo-900 text-xl tracking-tight">{{ data.password }}</p>
                  </div>
                  <button 
                    v-if="permissions.can_copy"
                    class="p-3 bg-white text-indigo-600 rounded-xl hover:bg-slate-50 transition shadow-sm opacity-0 group-hover:opacity-100"
                    title="Copy to clipboard"
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 5H6a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2v-1M8 5a2 2 0 002 2h2a2 2 0 002-2M8 5a2 2 0 012-2h2a2 2 0 012 2m0 0h2a2 2 0 012 2v3m2 4H10m0 0l3-3m-3 3l3 3" />
                    </svg>
                  </button>
              </div>
          </div>

          <div class="pt-6 border-t border-slate-100">
              <div class="flex items-center gap-2 text-xs font-bold text-slate-400">
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                  Share managed by Tresor Security
              </div>
          </div>
      </div>
    </div>
  </div>
</template>
