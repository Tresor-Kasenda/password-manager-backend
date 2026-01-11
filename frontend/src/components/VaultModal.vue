<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { vaultApi } from '../api/vault'
import { useVaultStore } from '../store/vaultStore'
import PasswordGenerator from './PasswordGenerator.vue'

const props = defineProps<{
  item?: any
}>()

const emit = defineEmits(['close'])
const { addItem, updateItem } = useVaultStore()

const title = ref('')
const website = ref('')
const username = ref('')
const password = ref('')
const notes = ref('')
const folder = ref('')
const masterPassword = ref('')

const loading = ref(false)
const error = ref('')
const isEdit = ref(false)

const decrypting = ref(false)

const handleDecrypt = async () => {
  if (!masterPassword.value) {
    error.value = 'Master password required to decrypt'
    return
  }
  
  decrypting.value = true
  error.value = ''
  try {
    const data = await vaultApi.get(props.item.id, masterPassword.value) as any
    password.value = data.password
    notes.value = data.notes || ''
  } catch (err: any) {
    error.value = err.message || 'Decryption failed'
  } finally {
    decrypting.value = false
  }
}

const handleSubmit = async () => {
  if (!masterPassword.value) {
    error.value = 'Master password required'
    return
  }

  loading.value = true
  error.value = ''
  
  const payload = {
    title: title.value,
    website: website.value || null,
    username: username.value || null,
    password: password.value,
    notes: notes.value || null,
    folder: folder.value || null,
    master_password: masterPassword.value,
  }

  try {
    if (isEdit.value) {
      const updated = await vaultApi.update(props.item.id, payload)
      updateItem(updated)
    } else {
      const created = await vaultApi.create(payload)
      addItem(created)
    }
    emit('close')
  } catch (err: any) {
    error.value = err.message || 'Operation failed'
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  if (props.item) {
    isEdit.value = true
    title.value = props.item.title
    website.value = props.item.website || ''
    username.value = props.item.username || ''
    folder.value = props.item.folder || ''
    // Password and notes are not provided in the list view, must be decrypted
  }
})
</script>

<template>
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-sm">
    <div class="bg-white rounded-2xl shadow-2xl w-full max-w-2xl max-h-[90vh] overflow-y-auto border border-slate-200">
      <div class="px-8 py-6 border-b border-slate-100 flex justify-between items-center sticky top-0 bg-white z-10">
        <h2 class="text-xl font-bold text-slate-900">{{ isEdit ? 'Edit Vault Item' : 'Add New Item' }}</h2>
        <button @click="$emit('close')" class="p-2 hover:bg-slate-100 rounded-full text-slate-400 transition">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <form @submit.prevent="handleSubmit" class="p-8 space-y-6">
        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div class="space-y-4">
            <div>
              <label class="block text-sm font-semibold text-slate-700 mb-1.5">Title</label>
              <input v-model="title" required type="text" class="input" placeholder="e.g. My GitHub Account" />
            </div>
            <div>
              <label class="block text-sm font-semibold text-slate-700 mb-1.5">Website URL</label>
              <input v-model="website" type="url" class="input" placeholder="https://github.com" />
            </div>
            <div>
              <label class="block text-sm font-semibold text-slate-700 mb-1.5">Username / Email</label>
              <input v-model="username" type="text" class="input" placeholder="johndoe" />
            </div>
            <div>
              <label class="block text-sm font-semibold text-slate-700 mb-1.5">Folder</label>
              <input v-model="folder" type="text" class="input" placeholder="Social, Work, etc." />
            </div>
          </div>

          <div class="space-y-4">
            <div>
              <label class="block text-sm font-semibold text-slate-700 mb-1.5">Password</label>
              <div class="relative">
                <input v-model="password" required type="text" class="input pr-10" placeholder="••••••••" />
                <div v-if="isEdit && !password" class="absolute inset-0 bg-white/90 backdrop-blur-sm flex items-center justify-center rounded-lg">
                  <button type="button" @click="handleDecrypt" :disabled="decrypting" class="text-xs font-bold text-indigo-600 hover:text-indigo-700">
                    {{ decrypting ? 'Decrypting...' : 'DECRYPT PASSWORD' }}
                  </button>
                </div>
              </div>
            </div>
            
            <PasswordGenerator @generated="(p) => password = p" />
            
            <div>
              <label class="block text-sm font-semibold text-slate-700 mb-1.5">Notes</label>
              <textarea v-model="notes" rows="3" class="input resize-none" placeholder="Additional details..."></textarea>
            </div>
          </div>
        </div>

        <div class="pt-6 border-t border-slate-100 flex flex-col gap-4">
          <div>
            <label class="block text-sm font-bold text-indigo-900 mb-2 uppercase tracking-wide">Master Password required</label>
            <input 
              v-model="masterPassword" 
              type="password" 
              required 
              class="w-full px-4 py-3 rounded-xl border-2 border-indigo-100 focus:border-indigo-500 focus:ring-4 focus:ring-indigo-500/10 outline-none transition" 
              placeholder="Confirm your master password" 
            />
          </div>

          <div v-if="error" class="p-3 bg-red-50 text-red-600 rounded-lg text-sm border border-red-100">
            {{ error }}
          </div>

          <div class="flex justify-end gap-3 mt-4">
            <button 
              type="button" 
              @click="$emit('close')"
              class="px-6 py-2.5 rounded-xl text-sm font-bold text-slate-600 hover:bg-slate-100 transition"
            >
              Cancel
            </button>
            <button 
              type="submit" 
              :disabled="loading"
              class="px-8 py-2.5 rounded-xl text-sm font-bold text-white bg-indigo-600 hover:bg-indigo-700 transition shadow-lg shadow-indigo-200 disabled:opacity-70"
            >
              {{ loading ? 'Saving...' : (isEdit ? 'Save Changes' : 'Create Item') }}
            </button>
          </div>
        </div>
      </form>
    </div>
  </div>
</template>

<style scoped>
@reference "tailwindcss";

.input {
  @apply w-full px-4 py-2.5 rounded-xl border border-slate-200 bg-slate-50 focus:bg-white focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 outline-none transition text-sm;
}
</style>
