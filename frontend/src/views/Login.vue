<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { authApi } from '../api/auth'
import { useAuthStore } from '../store/authStore'

const router = useRouter()
const { login: successLogin } = useAuthStore()

const email = ref('')
const masterPassword = ref('')
const error = ref('')
const loading = ref(false)

const handleLogin = async () => {
  loading.value = true
  error.value = ''
  try {
    const response = await authApi.login({
      email: email.value,
      master_password: masterPassword.value,
    }) as any
    successLogin(response.user, response.access_token)
    router.push('/')
  } catch (err: any) {
    error.value = err.message || 'Login failed'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen bg-slate-50 flex items-center justify-center p-4">
    <div class="max-w-md w-full bg-white rounded-2xl shadow-xl p-8 border border-slate-100">
      <div class="text-center mb-8">
        <h1 class="text-3xl font-bold text-slate-900 mb-2">Welcome Back</h1>
        <p class="text-slate-500">Enter your credentials to access your vault</p>
      </div>

      <form @submit.prevent="handleLogin" class="space-y-6">
        <div>
          <label class="block text-sm font-medium text-slate-700 mb-2">Email Address</label>
          <input 
            v-model="email"
            type="email" 
            required
            class="w-full px-4 py-3 rounded-lg border border-slate-200 focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 outline-none transition"
            placeholder="john@example.com"
          />
        </div>

        <div>
          <label class="block text-sm font-medium text-slate-700 mb-2">Master Password</label>
          <input 
            v-model="masterPassword"
            type="password" 
            required
            class="w-full px-4 py-3 rounded-lg border border-slate-200 focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 outline-none transition"
            placeholder="••••••••"
          />
        </div>

        <div v-if="error" class="bg-red-50 text-red-600 p-3 rounded-lg text-sm border border-red-100">
          {{ error }}
        </div>

        <button 
          type="submit"
          :disabled="loading"
          class="w-full bg-indigo-600 hover:bg-indigo-700 text-white font-semibold py-3 rounded-lg transition transform active:scale-[0.98] disabled:opacity-70 disabled:cursor-not-allowed"
        >
          <span v-if="loading">Signing in...</span>
          <span v-else>Sign In</span>
        </button>
      </form>

      <div class="mt-8 text-center">
        <p class="text-slate-500 text-sm">
          Don't have an account? 
          <router-link to="/register" class="text-indigo-600 font-semibold hover:underline">Create one</router-link>
        </p>
      </div>
    </div>
  </div>
</template>
