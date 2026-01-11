<script setup lang="ts">
import { ref } from 'vue'
import { vaultApi } from '../api/vault'

const props = defineProps<{
  initialLength?: number
}>()

const emit = defineEmits(['generated'])

const length = ref(props.initialLength || 20)
const useSpecial = ref(true)
const generating = ref(false)

const generate = async () => {
  generating.value = true
  try {
    const response = await vaultApi.generatePassword(length.value, useSpecial.value)
    emit('generated', response.password)
  } catch (err) {
    alert('Failed to generate password')
  } finally {
    generating.value = false
  }
}
</script>

<template>
  <div class="bg-indigo-50 p-4 rounded-xl border border-indigo-100">
    <div class="flex items-center justify-between mb-4">
      <h4 class="text-sm font-bold text-indigo-900 uppercase tracking-wider">Generator</h4>
      <div class="flex items-center gap-2">
        <input 
          v-model="length" 
          type="number" 
          min="8" 
          max="64" 
          class="w-16 px-2 py-1 bg-white border border-indigo-200 rounded text-xs outline-none"
        />
        <label class="flex items-center gap-1 text-xs text-indigo-700 font-medium">
          <input v-model="useSpecial" type="checkbox" class="rounded border-indigo-300 text-indigo-600 focus:ring-indigo-500" />
          Special
        </label>
      </div>
    </div>
    
    <button 
      type="button" 
      @click="generate"
      :disabled="generating"
      class="w-full bg-indigo-600 hover:bg-indigo-700 text-white text-xs font-bold py-2 rounded-lg transition transform active:scale-[0.98] disabled:opacity-70"
    >
      {{ generating ? 'Generating...' : 'Generate Secure Password' }}
    </button>
  </div>
</template>
