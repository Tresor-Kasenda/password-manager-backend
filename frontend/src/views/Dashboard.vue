<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useVaultStore } from '../store/vaultStore'
import VaultModal from '../components/VaultModal.vue'
import VaultHealth from '../components/VaultHealth.vue'
import SharingManager from '../components/SharingManager.vue'
import AccountSettings from '../components/AccountSettings.vue'
import ShareModal from '../components/ShareModal.vue'

const { items, loading, error, fetchVaults, removeItem } = useVaultStore()

const currentTab = ref('passwords')
const searchQuery = ref('')
const isModalOpen = ref(false)
const isShareModalOpen = ref(false)
const selectedItem = ref<any>(null)

const tabs = [
  { id: 'passwords', label: 'Vault', icon: 'M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z' },
  { id: 'health', label: 'Security Health', icon: 'M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z' },
  { id: 'sharing', label: 'Sharing', icon: 'M8.684 13.342C8.886 12.938 9 12.482 9 12c0-.482-.114-.938-.316-1.342m0 2.684a3 3 0 110-2.684m0 2.684l6.632 3.316m-6.632-6l6.632-3.316m0 0a3 3 0 105.367-2.684 3 3 0 00-5.367 2.684zm0 9.316a3 3 0 105.368 2.684 3 3 0 00-5.368-2.684z' },
  { id: 'settings', label: 'Settings', icon: 'M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.723 1.723 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z' }
]

const filteredItems = computed(() => {
  if (!searchQuery.value) return items.value
  const query = searchQuery.value.toLowerCase()
  return items.value.filter(item => 
    item.title.toLowerCase().includes(query) || 
    item.website?.toLowerCase().includes(query) ||
    item.username?.toLowerCase().includes(query)
  )
})

const openAddModal = () => {
  selectedItem.value = null
  isModalOpen.value = true
}

const openEditModal = (item: any) => {
  selectedItem.value = item
  isModalOpen.value = true
}

const openShareModal = (item: any) => {
  selectedItem.value = item
  isShareModalOpen.value = true
}

const handleDelete = async (id: string) => {
  if (confirm('Are you sure you want to delete this entry?')) {
    try {
      removeItem(id)
    } catch (err) {
      alert('Failed to delete entry')
    }
  }
}

onMounted(() => {
  fetchVaults()
})
</script>

<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <!-- Header -->
    <div class="flex flex-col md:flex-row md:items-center md:justify-between mb-10 gap-6">
      <div>
        <h1 class="text-3xl font-black text-slate-900 tracking-tight">Tresor Vault</h1>
        <p class="text-slate-500 text-sm font-medium mt-1">Protecting what matters most.</p>
      </div>
      
      <div v-if="currentTab === 'passwords'" class="flex items-center gap-3">
        <div class="relative">
          <input 
            v-model="searchQuery"
            type="text" 
            placeholder="Search vault..."
            class="pl-10 pr-4 py-3 bg-white border border-slate-200 rounded-xl text-sm focus:ring-4 focus:ring-indigo-500/10 focus:border-indigo-500 outline-none w-64 transition-all shadow-sm"
          />
          <span class="absolute left-3.5 top-3.5 text-slate-400">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
          </span>
        </div>
        
        <button 
          @click="openAddModal"
          class="bg-indigo-600 hover:bg-indigo-700 text-white px-5 py-3 rounded-xl text-sm font-bold transition shadow-lg shadow-indigo-200 flex items-center gap-2 transform active:scale-95"
        >
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
          </svg>
          Add Item
        </button>
      </div>
    </div>

    <!-- Tabs -->
    <div class="flex space-x-1 bg-slate-200/50 p-1.5 rounded-2xl mb-10 max-w-fit shadow-inner">
      <button 
        v-for="tab in tabs" 
        :key="tab.id"
        @click="currentTab = tab.id"
        class="flex items-center gap-2 px-6 py-2.5 rounded-xl text-sm font-bold transition-all"
        :class="currentTab === tab.id ? 'bg-white text-indigo-600 shadow-sm' : 'text-slate-500 hover:text-slate-800'"
      >
        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" :d="tab.icon" />
        </svg>
        {{ tab.label }}
      </button>
    </div>

    <!-- Content -->
    <div v-if="currentTab === 'passwords'">
      <div v-if="loading" class="flex justify-center items-center py-20">
        <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600"></div>
      </div>

      <div v-else-if="error" class="bg-red-50 text-red-600 p-4 rounded-lg border border-red-100 text-center">
        {{ error }}
        <button @click="fetchVaults" class="block mx-auto mt-2 text-indigo-600 font-semibold underline">Try again</button>
      </div>

      <div v-else-if="filteredItems.length === 0" class="bg-white rounded-3xl border-2 border-dashed border-slate-200 p-24 text-center">
        <div class="w-20 h-20 bg-indigo-50 rounded-full flex items-center justify-center mx-auto mb-6">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-10 w-10 text-indigo-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
          </svg>
        </div>
        <h3 class="text-xl font-bold text-slate-900">Your vault is empty</h3>
        <p class="text-slate-500 max-w-sm mx-auto mt-2">Start adding your digital assets. Everything is encrypted locally before being stored in the cloud.</p>
        <button 
          @click="openAddModal"
          class="mt-8 bg-indigo-600 hover:bg-indigo-700 text-white px-8 py-3 rounded-xl font-bold transition shadow-lg shadow-indigo-200"
        >
          Add Your First Item
        </button>
      </div>

      <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        <div 
          v-for="item in filteredItems" 
          :key="item.id"
          class="bg-white rounded-2xl border border-slate-200 p-6 hover:shadow-xl hover:border-indigo-100 transition-all group shadow-sm flex flex-col justify-between"
        >
          <div>
            <div class="flex justify-between items-start mb-6">
              <div class="flex items-center gap-4">
                <div class="w-12 h-12 bg-indigo-50 rounded-2xl flex items-center justify-center text-indigo-600 group-hover:bg-indigo-600 group-hover:text-white transition-colors duration-300 shadow-sm">
                  <span class="font-black text-xl">{{ item.title[0].toUpperCase() }}</span>
                </div>
                <div>
                  <h3 class="font-bold text-slate-900 leading-tight">{{ item.title }}</h3>
                  <a v-if="item.website" :href="item.website" target="_blank" class="text-xs text-indigo-500 hover:underline font-medium">{{ item.website }}</a>
                  <p v-else class="text-xs text-slate-400 font-medium italic">No URL</p>
                </div>
              </div>
              <div class="flex items-center gap-1 opacity-100 lg:opacity-0 group-hover:opacity-100 transition-opacity duration-300">
                <button @click="openShareModal(item)" class="p-2.5 hover:bg-green-50 rounded-xl text-slate-400 hover:text-green-600 transition-colors" title="Share Item">
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.684 13.342C8.886 12.938 9 12.482 9 12c0-.482-.114-.938-.316-1.342m0 2.684a3 3 0 110-2.684m0 2.684l6.632 3.316m-6.632-6l6.632-3.316m0 0a3 3 0 105.367-2.684 3 3 0 00-5.367 2.684zm0 9.316a3 3 0 105.368 2.684 3 3 0 00-5.368-2.684z" />
                  </svg>
                </button>
                <button @click="openEditModal(item)" class="p-2.5 hover:bg-indigo-50 rounded-xl text-slate-400 hover:text-indigo-600 transition-colors" title="Edit Item">
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
                  </svg>
                </button>
                <button @click="handleDelete(item.id)" class="p-2.5 hover:bg-red-50 rounded-xl text-slate-400 hover:text-red-500 transition-colors">
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                  </svg>
                </button>
              </div>
            </div>
            
            <div class="space-y-3 bg-slate-50/50 p-4 rounded-xl border border-slate-100 mb-2">
              <div class="flex justify-between items-center text-sm">
                <span class="text-slate-500 font-medium">Username</span>
                <span class="text-slate-900 font-bold truncate max-w-[150px]">{{ item.username || '-' }}</span>
              </div>
              <div class="flex justify-between items-center text-sm">
                <span class="text-slate-500 font-medium">Folder</span>
                <span class="px-2 py-0.5 bg-slate-200 text-slate-700 rounded-md text-[10px] uppercase font-black">{{ item.folder || 'ROOT' }}</span>
              </div>
            </div>
          </div>
          
          <div class="mt-4 pt-4 border-t border-slate-100 flex justify-between items-center">
            <span v-if="item.favorite" class="text-yellow-500">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="currentColor" viewBox="0 0 20 20">
                <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z" />
              </svg>
            </span>
            <span class="text-[10px] text-slate-400 font-bold uppercase">Updated {{ new Date(item.updated_at).toLocaleDateString() }}</span>
          </div>
        </div>
      </div>
    </div>

    <div v-else-if="currentTab === 'health'">
      <VaultHealth />
    </div>

    <div v-else-if="currentTab === 'sharing'">
      <SharingManager />
    </div>

    <div v-else-if="currentTab === 'settings'">
      <AccountSettings />
    </div>

    <!-- Modal -->
    <VaultModal 
      v-if="isModalOpen" 
      :item="selectedItem" 
      @close="isModalOpen = false" 
    />

    <ShareModal
      v-if="isShareModalOpen"
      :vault-id="selectedItem.id"
      :item-title="selectedItem.title"
      @close="isShareModalOpen = false"
    />
  </div>
</template>
