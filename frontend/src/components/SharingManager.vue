<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { sharingApi } from '../api/healthAndSharing'

const sentShares = ref<any[]>([])
const receivedShares = ref<any[]>([])
const loading = ref(false)
const error = ref('')

const fetchShares = async () => {
  loading.value = true
  error.value = ''
  try {
    const response = await sharingApi.listShares() as any
    sentShares.value = response.sent || []
    receivedShares.value = response.received || []
  } catch (err: any) {
    error.value = err.message || 'Failed to fetch shares'
  } finally {
    loading.value = false
  }
}

const handleRevoke = async (token: string) => {
  if (confirm('Are you sure you want to revoke this share?')) {
    try {
      await sharingApi.revokeShare(token)
      sentShares.value = sentShares.value.filter(s => s.share_token !== token)
    } catch (err) {
      alert('Failed to revoke share')
    }
  }
}

onMounted(() => {
  fetchShares()
})
</script>

<template>
  <div class="space-y-8">
    <div v-if="loading" class="flex justify-center py-12">
      <div class="animate-spin rounded-full h-10 w-10 border-b-2 border-indigo-600"></div>
    </div>

    <div v-else-if="error" class="bg-red-50 text-red-600 p-4 rounded-xl border border-red-100 text-center">
      {{ error }}
      <button @click="fetchShares" class="block mx-auto mt-2 font-bold underline">Try again</button>
    </div>

    <div v-else class="grid grid-cols-1 lg:grid-cols-2 gap-8">
      <!-- Sent Shares -->
      <div class="bg-white rounded-2xl border border-slate-100 shadow-sm overflow-hidden">
        <div class="px-6 py-4 border-b border-slate-100 bg-slate-50/50 flex justify-between items-center">
          <h3 class="font-bold text-slate-900">Sent Shares</h3>
          <span class="text-xs font-bold bg-slate-200 text-slate-600 px-2 py-1 rounded-md">{{ sentShares.length }}</span>
        </div>
        <div class="divide-y divide-slate-100">
          <div v-if="sentShares.length === 0" class="p-12 text-center text-slate-400 text-sm italic">
            You haven't shared any passwords yet.
          </div>
          <div 
            v-for="share in sentShares" 
            :key="share.id"
            class="p-4 flex items-center justify-between hover:bg-slate-50 transition"
          >
            <div>
              <p class="font-bold text-slate-900 text-sm">{{ share.recipient_email }}</p>
              <p class="text-xs text-slate-500">Shared on {{ new Date(share.created_at).toLocaleDateString() }}</p>
            </div>
            <button 
              @click="handleRevoke(share.share_token)"
              class="text-xs font-bold text-red-600 hover:text-red-700 p-2 rounded-lg hover:bg-red-50 transition"
            >
              Revoke
            </button>
          </div>
        </div>
      </div>

      <!-- Received Shares -->
      <div class="bg-white rounded-2xl border border-slate-100 shadow-sm overflow-hidden">
        <div class="px-6 py-4 border-b border-slate-100 bg-slate-50/50 flex justify-between items-center">
          <h3 class="font-bold text-slate-900">Received Shares</h3>
          <span class="text-xs font-bold bg-slate-200 text-slate-600 px-2 py-1 rounded-md">{{ receivedShares.length }}</span>
        </div>
        <div class="divide-y divide-slate-100">
          <div v-if="receivedShares.length === 0" class="p-12 text-center text-slate-400 text-sm italic">
            No one has shared a password with you yet.
          </div>
          <div 
            v-for="share in receivedShares" 
            :key="share.id"
            class="p-4 flex items-center justify-between hover:bg-slate-50 transition"
          >
            <div class="flex items-center gap-3">
              <div class="w-8 h-8 bg-indigo-100 rounded-lg flex items-center justify-center text-indigo-600 font-bold text-sm">
                S
              </div>
              <div>
                <p class="font-bold text-slate-900 text-sm">{{ share.vault_title || 'Shared Vault Item' }}</p>
                <p class="text-xs text-slate-500">From User ID: {{ share.owner_id.substring(0,8) }}...</p>
              </div>
            </div>
            <router-link 
              :to="`/shared/${share.share_token}`"
              class="text-xs font-bold text-indigo-600 hover:text-indigo-700 p-2 rounded-lg hover:bg-indigo-50 transition"
            >
              View
            </router-link>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
