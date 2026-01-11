<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { healthApi } from '../api/healthAndSharing'

const report = ref<any>(null)
const loading = ref(false)
const error = ref('')

const fetchReport = async () => {
  loading.value = true
  error.value = ''
  try {
    report.value = await healthApi.getReport()
  } catch (err: any) {
    error.value = err.message || 'Failed to fetch health report'
  } finally {
    loading.value = false
  }
}

const getScoreColor = (score: number) => {
  if (score >= 80) return 'text-green-600 bg-green-50'
  if (score >= 60) return 'text-yellow-600 bg-yellow-50'
  return 'text-red-600 bg-red-50'
}

onMounted(() => {
  fetchReport()
})
</script>

<template>
  <div class="space-y-8">
    <div v-if="loading" class="flex justify-center py-12">
      <div class="animate-spin rounded-full h-10 w-10 border-b-2 border-indigo-600"></div>
    </div>

    <div v-else-if="error" class="bg-red-50 text-red-600 p-4 rounded-xl border border-red-100 text-center">
      {{ error }}
      <button @click="fetchReport" class="block mx-auto mt-2 font-bold underline">Try again</button>
    </div>

    <div v-else-if="report" class="space-y-8">
      <!-- Summary Cards -->
      <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
        <div class="bg-white p-6 rounded-2xl border border-slate-100 shadow-sm">
          <p class="text-xs font-bold text-slate-500 uppercase tracking-wider mb-1">Overall Health</p>
          <div class="flex items-end gap-2">
            <span class="text-3xl font-black text-slate-900">{{ report.overall_score }}</span>
            <span class="text-sm font-bold text-slate-400 mb-1">/100</span>
          </div>
          <div class="mt-3 h-2 bg-slate-100 rounded-full overflow-hidden">
            <div 
              class="h-full bg-indigo-600 transition-all duration-1000" 
              :style="{ width: `${report.overall_score}%` }"
            ></div>
          </div>
        </div>
        
        <div class="bg-white p-6 rounded-2xl border border-slate-100 shadow-sm">
          <p class="text-xs font-bold text-slate-500 uppercase tracking-wider mb-1">Breached</p>
          <p class="text-3xl font-black text-red-600">{{ report.statistics?.breached_passwords || 0 }}</p>
          <p class="text-xs text-slate-400 mt-1">Found in data leaks</p>
        </div>

        <div class="bg-white p-6 rounded-2xl border border-slate-100 shadow-sm">
          <p class="text-xs font-bold text-slate-500 uppercase tracking-wider mb-1">Weak</p>
          <p class="text-3xl font-black text-yellow-600">{{ report.statistics?.weak_passwords || 0 }}</p>
          <p class="text-xs text-slate-400 mt-1">Easy to crack</p>
        </div>

        <div class="bg-white p-6 rounded-2xl border border-slate-100 shadow-sm">
          <p class="text-xs font-bold text-slate-500 uppercase tracking-wider mb-1">Reused</p>
          <p class="text-3xl font-black text-orange-600">{{ report.statistics?.reused_passwords || 0 }}</p>
          <p class="text-xs text-slate-400 mt-1">Across multiple sites</p>
        </div>
      </div>

      <!-- Vulnerable Passwords -->
      <div class="bg-white rounded-2xl border border-slate-100 shadow-sm overflow-hidden">
        <div class="px-6 py-4 border-b border-slate-100 bg-slate-50/50">
          <h3 class="font-bold text-slate-900">Security Issues</h3>
        </div>
        <div class="divide-y divide-slate-100">
          <div v-if="!report.details || report.details.length === 0" class="p-12 text-center text-slate-500 italic">
            No major security issues found. Great job!
          </div>
          <div 
            v-for="vuln in report.details" 
            :key="vuln.vault_id"
            class="p-6 flex flex-col md:flex-row md:items-center justify-between gap-4 hover:bg-slate-50 transition"
          >
            <div class="flex items-center gap-4">
              <div class="w-10 h-10 rounded-xl flex items-center justify-center font-bold" :class="getScoreColor(vuln.score)">
                {{ vuln.score }}
              </div>
              <div>
                <h4 class="font-bold text-slate-900">{{ vuln.title }}</h4>
                <p class="text-xs text-slate-500">{{ vuln.website || 'No website' }}</p>
              </div>
            </div>
            
            <div class="flex flex-wrap gap-2">
              <span v-if="vuln.is_breached" class="px-3 py-1 bg-red-100 text-red-700 text-xs font-black rounded-full">BREACHED</span>
              <span v-if="vuln.score < 60" class="px-3 py-1 bg-yellow-100 text-yellow-700 text-xs font-black rounded-full">WEAK</span>
              <span v-if="vuln.is_reused" class="px-3 py-1 bg-orange-100 text-orange-700 text-xs font-black rounded-full">REUSED</span>
              <span v-if="vuln.age_days > 180" class="px-3 py-1 bg-blue-100 text-blue-700 text-xs font-black rounded-full">OLD</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
