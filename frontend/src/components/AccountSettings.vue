<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { twofaApi } from '../api/twofa'
import { importApi } from '../api/import'
import { useAuthStore } from '../store/authStore'

const { user } = useAuthStore()
const twofaStatus = ref(user.value?.two_factor_enabled || false)
const show2FASetup = ref(false)
const qrCode = ref('')
const backupCodes = ref<string[]>([])
const verificationToken = ref('')
const masterPassword = ref('')
const loading = ref(false)
const error = ref('')

const importFile = ref<File | null>(null)
const importSource = ref('chrome')
const importFormats = ref<any[]>([])
const importSession = ref<any>(null)
const importMasterPassword = ref('')

const fetchFormats = async () => {
  try {
    const res = await importApi.getSupportedFormats() as any
    importFormats.value = res.formats
  } catch (err) {}
}

const handleEnable2FA = async () => {
  if (!masterPassword.value) return
  loading.value = true
  error.value = ''
  try {
    const res = await twofaApi.enable(masterPassword.value) as any
    qrCode.value = res.qr_code
    backupCodes.value = res.backup_codes
    show2FASetup.value = true
  } catch (err: any) {
    error.value = err.message || 'Failed to initiate 2FA'
  } finally {
    loading.value = false
  }
}

const handleVerify2FA = async () => {
  loading.value = true
  error.value = ''
  try {
    await twofaApi.verifyAndEnable(verificationToken.value)
    twofaStatus.value = true
    show2FASetup.value = false
    alert('2FA enabled successfully!')
  } catch (err: any) {
    error.value = err.message || 'Verification failed'
  } finally {
    loading.value = false
  }
}

const handleFileChange = (e: any) => {
  importFile.value = e.target.files[0]
}

const handleUpload = async () => {
  if (!importFile.value) return
  loading.value = true
  error.value = ''
  try {
    const reader = new FileReader()
    reader.onload = async (e) => {
      const content = e.target?.result as string
      const res = await importApi.upload(content, importFile.value!.name, importSource.value) as any
      importSession.value = res
    }
    reader.readAsText(importFile.value)
  } catch (err: any) {
    error.value = err.message || 'Upload failed'
  } finally {
    loading.value = false
  }
}

const handleConfirmImport = async () => {
  if (!importMasterPassword.value) return
  loading.value = true
  error.value = ''
  try {
    await importApi.confirm(importSession.value.session_id, importMasterPassword.value)
    alert('Import completed successfully!')
    importSession.value = null
  } catch (err: any) {
    error.value = err.message || 'Import confirmation failed'
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchFormats()
})
</script>

<template>
  <div class="max-w-4xl mx-auto space-y-12">
    <!-- 2FA Section -->
    <section class="bg-white rounded-3xl border border-slate-100 shadow-sm overflow-hidden">
      <div class="px-8 py-6 border-b border-slate-100 flex justify-between items-center bg-slate-50/50">
        <div>
          <h3 class="text-xl font-bold text-slate-900">Two-Factor Authentication</h3>
          <p class="text-sm text-slate-500 font-medium">Add an extra layer of security to your account.</p>
        </div>
        <div 
          class="px-3 py-1 rounded-full text-[10px] font-black uppercase tracking-wider"
          :class="twofaStatus ? 'bg-green-100 text-green-700' : 'bg-slate-100 text-slate-500'"
        >
          {{ twofaStatus ? 'Enabled' : 'Disabled' }}
        </div>
      </div>

      <div class="p-8">
        <div v-if="!show2FASetup && !twofaStatus" class="space-y-6">
          <div class="flex items-start gap-4 p-4 bg-indigo-50 rounded-2xl border border-indigo-100">
            <div class="p-2 bg-indigo-600 text-white rounded-xl">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
              </svg>
            </div>
            <div>
              <h4 class="font-bold text-indigo-900">Recommended for all users</h4>
              <p class="text-sm text-indigo-700 font-medium mt-0.5">Protect your vault even if your master password is compromised.</p>
            </div>
          </div>

          <div class="max-w-sm">
            <label class="block text-sm font-bold text-slate-700 mb-2">Confirm Master Password</label>
            <div class="flex gap-2">
              <input 
                v-model="masterPassword" 
                type="password" 
                class="flex-1 px-4 py-2 bg-slate-50 border border-slate-200 rounded-xl focus:ring-2 focus:ring-indigo-500 outline-none transition" 
                placeholder="Enter master password"
              />
              <button 
                @click="handleEnable2FA" 
                :disabled="loading"
                class="bg-indigo-600 hover:bg-indigo-700 text-white px-6 py-2 rounded-xl text-sm font-bold transition shadow-lg shadow-indigo-100 disabled:opacity-50"
              >
                Enable
              </button>
            </div>
          </div>
        </div>

        <div v-if="show2FASetup" class="grid grid-cols-1 md:grid-cols-2 gap-12">
          <div class="space-y-6">
            <h4 class="font-bold text-slate-900">1. Scan QR Code</h4>
            <p class="text-sm text-slate-500 leading-relaxed">Scan this code with your authenticator app (Google Authenticator, Authy, etc.)</p>
            <div class="p-4 bg-white border-2 border-slate-100 rounded-3xl inline-block shadow-inner">
              <img :src="qrCode" alt="2FA QR Code" class="w-48 h-48" />
            </div>
          </div>

          <div class="space-y-6">
            <h4 class="font-bold text-slate-900">2. Verify Token</h4>
            <p class="text-sm text-slate-500 leading-relaxed">Enter the 6-digit code from your app to confirm setup.</p>
            <div class="space-y-4">
              <input 
                v-model="verificationToken" 
                type="text" 
                maxlength="6"
                class="w-full text-center text-4xl font-black tracking-[1em] px-4 py-4 bg-slate-50 border-2 border-slate-200 rounded-2xl focus:border-indigo-500 focus:bg-white outline-none transition"
                placeholder="000000"
              />
              <button 
                @click="handleVerify2FA" 
                :disabled="loading || verificationToken.length < 6"
                class="w-full bg-indigo-600 hover:bg-indigo-700 text-white py-4 rounded-2xl font-black text-lg transition shadow-xl shadow-indigo-200 disabled:opacity-50"
              >
                Verify & Enable
              </button>
            </div>

            <div class="p-6 bg-yellow-50 rounded-2xl border border-yellow-100">
              <h5 class="text-xs font-black text-yellow-800 uppercase tracking-widest mb-3">Backup Codes</h5>
              <div class="grid grid-cols-2 gap-2">
                <code v-for="code in backupCodes" :key="code" class="text-xs font-bold text-yellow-900 bg-white/50 px-2 py-1 rounded border border-yellow-200">{{ code }}</code>
              </div>
              <p class="text-[10px] text-yellow-700 font-bold mt-4">Save these codes in a safe place. You can use them to log in if you lose your phone.</p>
            </div>
          </div>
        </div>

        <div v-if="twofaStatus" class="p-8 text-center">
            <div class="w-16 h-16 bg-green-50 text-green-600 rounded-full flex items-center justify-center mx-auto mb-4">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
                </svg>
            </div>
            <h4 class="text-xl font-bold text-slate-900">2FA is currently active</h4>
            <p class="text-sm text-slate-500 mt-2">Your account is secured with two-factor authentication.</p>
            <button class="mt-6 text-red-600 font-bold text-sm hover:underline">Disable 2FA</button>
        </div>
      </div>
    </section>

    <!-- Import Section -->
    <section class="bg-white rounded-3xl border border-slate-100 shadow-sm overflow-hidden">
      <div class="px-8 py-6 border-b border-slate-100 bg-slate-50/50">
        <h3 class="text-xl font-bold text-slate-900">Import Passwords</h3>
        <p class="text-sm text-slate-500 font-medium">Easily move your data from other password managers.</p>
      </div>

      <div class="p-8">
        <div v-if="!importSession" class="space-y-8">
          <div class="grid grid-cols-1 md:grid-cols-2 gap-8">
            <div>
              <label class="block text-sm font-bold text-slate-700 mb-2">Select Source</label>
              <select v-model="importSource" class="w-full px-4 py-3 bg-slate-50 border border-slate-200 rounded-xl focus:ring-2 focus:ring-indigo-500 outline-none transition font-medium">
                <option v-for="format in importFormats" :key="format.id" :value="format.id">{{ format.name }}</option>
              </select>
            </div>
            <div>
              <label class="block text-sm font-bold text-slate-700 mb-2">Select File</label>
              <input type="file" @change="handleFileChange" class="w-full text-sm text-slate-500 file:mr-4 file:py-2.5 file:px-4 file:rounded-xl file:border-0 file:text-xs file:font-black file:bg-indigo-50 file:text-indigo-700 hover:file:bg-indigo-100 transition" />
            </div>
          </div>
          
          <button 
            @click="handleUpload" 
            :disabled="loading || !importFile"
            class="bg-slate-900 hover:bg-slate-800 text-white px-10 py-3 rounded-xl text-sm font-black transition shadow-xl disabled:opacity-50"
          >
            Upload & Preview
          </button>
        </div>

        <div v-else class="space-y-8 animate-in fade-in slide-in-from-bottom-4">
          <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
            <div class="p-4 bg-slate-50 rounded-2xl border border-slate-100">
              <p class="text-[10px] font-black text-slate-400 uppercase tracking-widest">Total</p>
              <p class="text-2xl font-black text-slate-900">{{ importSession.total_entries }}</p>
            </div>
            <div class="p-4 bg-green-50 rounded-2xl border border-green-100">
              <p class="text-[10px] font-black text-green-400 uppercase tracking-widest">Valid</p>
              <p class="text-2xl font-black text-green-700">{{ importSession.valid_entries }}</p>
            </div>
            <div class="p-4 bg-red-50 rounded-2xl border border-red-100">
              <p class="text-[10px] font-black text-red-400 uppercase tracking-widest">Invalid</p>
              <p class="text-2xl font-black text-red-700">{{ importSession.invalid_entries }}</p>
            </div>
            <div class="p-4 bg-yellow-50 rounded-2xl border border-yellow-100">
              <p class="text-[10px] font-black text-yellow-400 uppercase tracking-widest">Warnings</p>
              <p class="text-2xl font-black text-yellow-700">{{ importSession.warnings }}</p>
            </div>
          </div>

          <div class="space-y-4">
              <h4 class="font-black text-slate-900 text-sm">Preview (First 10 entries)</h4>
              <div class="border border-slate-100 rounded-2xl overflow-hidden divide-y divide-slate-100">
                  <div v-for="item in importSession.preview" :key="item.title" class="px-4 py-3 flex justify-between text-xs font-medium">
                      <span class="text-slate-900">{{ item.title }}</span>
                      <span class="text-slate-500">{{ item.username || '-' }}</span>
                  </div>
              </div>
          </div>

          <div class="p-8 bg-indigo-600 rounded-3xl text-white space-y-6">
              <div>
                <h4 class="text-xl font-black">Confirm Import</h4>
                <p class="text-indigo-100 text-sm">We need your master password to encrypt the newly imported data.</p>
              </div>
              <div class="flex flex-col md:flex-row gap-4">
                  <input 
                    v-model="importMasterPassword" 
                    type="password" 
                    class="flex-1 px-6 py-3 bg-white/10 border border-white/20 rounded-2xl focus:bg-white/20 outline-none transition placeholder:text-white/40" 
                    placeholder="Enter master password"
                  />
                  <button 
                    @click="handleConfirmImport"
                    class="bg-white text-indigo-600 px-10 py-3 rounded-2xl font-black hover:bg-slate-50 transition shadow-xl"
                  >
                    Start Import
                  </button>
              </div>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>
