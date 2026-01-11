<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { authApi } from '../api/auth'

const router = useRouter()

const email = ref('')
const masterPassword = ref('')
const confirmPassword = ref('')
const error = ref('')
const loading = ref(false)

const handleRegister = async () => {
  if (masterPassword.value !== confirmPassword.value) {
    error.value = 'Passwords do not match'
    return
  }

  loading.value = true
  error.value = ''
  try {
    await authApi.register({
      email: email.value,
      master_password: masterPassword.value,
    })
    router.push('/login')
  } catch (err: any) {
    error.value = err.message || 'Registration failed'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen bg-slate-50 flex items-center justify-center p-4">
    <div class="max-w-md w-full bg-white rounded-2xl shadow-xl p-8 border border-slate-100">
      <div class="text-center mb-8">
        <h1 class="text-3xl font-bold text-slate-900 mb-2">Create Account</h1>
        <p class="text-slate-500">Secure your digital life with Tresor</p>
      </div>

      <form @submit.prevent="handleRegister" class="space-y-6">
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
            minlength="8"
            class="w-full px-4 py-3 rounded-lg border border-slate-200 focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 outline-none transition"
            placeholder="Min 8 characters"
          />
        </div>

        <div>
          <label class="block text-sm font-medium text-slate-700 mb-2">Confirm Password</label>
          <input 
            v-model="confirmPassword"
            type="password" 
            required
            class="w-full px-4 py-3 rounded-lg border border-slate-200 focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 outline-none transition"
            placeholder="Repeat master password"
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
          <span v-if="loading">Creating account...</span>
          <span v-else>Register</span>
        </button>
      </form>

      <div class="mt-8 text-center">
        <p class="text-slate-500 text-sm">
          Already have an account? 
          <router-link to="/login" class="text-indigo-600 font-semibold hover:underline">Sign In</router-link>
        </p>
      </div>
    </div>
  </div>
</template>
