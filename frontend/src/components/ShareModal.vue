<script setup lang="ts">
import { ref } from 'vue'
import { sharingApi } from '../api/healthAndSharing'

const props = defineProps<{
  vaultId: string
  itemTitle: string
}>()

const emit = defineEmits(['close', 'shared'])

const email = ref('')
const expiresIn = ref(24)
const requirePassword = ref(false)
const sharePassword = ref('')
const loading = ref(false)
const error = ref('')
const shareResult = ref<any>(null)

const handleShare = async () => {
  if (!email.value) return
  loading.value = true
  error.value = ''
  try {
    const res = await sharingApi.share(props.vaultId, email.value, {
      expires_in_hours: expiresIn.value,
      require_password: requirePassword.value,
      share_password: requirePassword.value ? sharePassword.value : undefined
    }) as any
    shareResult.value = res
    emit('shared', res)
  } catch (err: any) {
    error.value = err.message || 'Failed to share item'
  } finally {
    loading.value = false
  }
}

const copyToClipboard = (text: string) => {
  navigator.clipboard.writeText(text)
  alert('Link copied to clipboard!')
}
</script>

<template>
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-sm animate-in fade-in duration-300">
    <div class="bg-white rounded-[2rem] w-full max-w-lg shadow-2xl border border-white/20 overflow-hidden transform animate-in zoom-in-95 slide-in-from-bottom-4 duration-300">
      
      <!-- Header -->
      <div class="px-8 py-6 border-b border-slate-100 flex justify-between items-center bg-slate-50/50">
        <div>
          <h2 class="text-xl font-black text-slate-900">Share Item</h2>
          <p class="text-sm text-slate-500 font-medium">Sharing: {{ itemTitle }}</p>
        </div>
        <button @click="$emit('close')" class="p-2 hover:bg-white rounded-xl transition-colors text-slate-400 hover:text-slate-600">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <div class="p-8">
        <div v-if="!shareResult" class="space-y-6">
          <div v-if="error" class="p-4 bg-red-50 border border-red-100 text-red-600 rounded-2xl text-sm font-bold flex items-center gap-3 animate-in shake duration-500">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            {{ error }}
          </div>

          <div class="space-y-4">
            <div>
              <label class="block text-xs font-black text-slate-400 uppercase tracking-widest mb-2 ml-1">Recipient Email</label>
              <input 
                v-model="email" 
                type="email" 
                placeholder="colleague@company.com"
                class="w-full px-5 py-3.5 bg-slate-50 border border-slate-200 rounded-2xl focus:bg-white focus:ring-4 focus:ring-indigo-500/10 focus:border-indigo-500 outline-none transition-all font-medium"
              />
            </div>

            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-xs font-black text-slate-400 uppercase tracking-widest mb-2 ml-1">Expires In</label>
                <select v-model="expiresIn" class="w-full px-5 py-3.5 bg-slate-50 border border-slate-200 rounded-2xl focus:bg-white focus:ring-4 focus:ring-indigo-500/10 focus:border-indigo-500 outline-none transition-all font-medium">
                  <option :value="1">1 Hour</option>
                  <option :value="24">24 Hours</option>
                  <option :value="72">3 Days</option>
                  <option :value="168">7 Days</option>
                </select>
              </div>
              <div class="flex flex-col justify-end">
                  <label class="flex items-center gap-3 px-5 py-3.5 bg-slate-50 border border-slate-200 rounded-2xl cursor-pointer hover:bg-slate-100 transition">
                      <input type="checkbox" v-model="requirePassword" class="w-4 h-4 text-indigo-600 rounded">
                      <span class="text-sm font-bold text-slate-700">Protect with Password</span>
                  </label>
              </div>
            </div>

            <div v-if="requirePassword" class="animate-in slide-in-from-top-2 duration-300">
              <label class="block text-xs font-black text-slate-400 uppercase tracking-widest mb-2 ml-1">Share Password</label>
              <input 
                v-model="sharePassword" 
                type="text" 
                placeholder="Secure access code"
                class="w-full px-5 py-3.5 bg-indigo-50/30 border border-indigo-100 rounded-2xl focus:bg-white focus:ring-4 focus:ring-indigo-500/10 focus:border-indigo-500 outline-none transition-all font-medium"
              />
            </div>
          </div>

          <button 
            @click="handleShare" 
            :disabled="loading || !email"
            class="w-full bg-indigo-600 hover:bg-indigo-700 disabled:opacity-50 text-white py-4 rounded-2xl font-black text-lg transition-all shadow-xl shadow-indigo-200 flex items-center justify-center gap-3 transform active:scale-95"
          >
            <span v-if="loading" class="animate-spin border-2 border-white/20 border-t-white rounded-full h-5 w-5"></span>
            {{ loading ? 'Generating Link...' : 'Create Secure Link' }}
          </button>
        </div>

        <div v-else class="space-y-6 animate-in zoom-in-95 duration-300">
            <div class="text-center pb-4">
                <div class="w-16 h-16 bg-green-50 text-green-600 rounded-full flex items-center justify-center mx-auto mb-4">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                    </svg>
                </div>
                <h3 class="text-2xl font-black text-slate-900">Successfully Shared!</h3>
                <p class="text-slate-500 text-sm mt-1">Send this link to the recipient securely.</p>
            </div>

            <div class="p-4 bg-slate-50 border border-slate-200 rounded-2xl flex items-center justify-between gap-4">
                <code class="text-xs font-bold text-indigo-600 truncate">{{ shareResult.share_url }}</code>
                <button 
                  @click="copyToClipboard(shareResult.share_url)"
                  class="bg-white text-slate-900 border border-slate-200 px-4 py-2 rounded-xl text-xs font-bold hover:bg-slate-50 transition shadow-sm whitespace-nowrap"
                >
                    Copy Link
                </button>
            </div>

            <button 
              @click="$emit('close')"
              class="w-full bg-slate-900 text-white py-4 rounded-2xl font-black transition-all shadow-xl"
            >
                Done
            </button>
        </div>
      </div>
    </div>
  </div>
</template>
