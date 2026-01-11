<script setup lang="ts">
import { useAuthStore } from './store/authStore'

const { isAuthenticated, login, logout } = useAuthStore()
</script>

<template>
  <div class="card">
    <h1>Tresor Password Manager</h1>
    <p v-if="!isAuthenticated">Please sign in to access your vault.</p>
    <p v-else>Welcome back! Your vault is secure.</p>
    
    <div style="margin-top: 2rem;">
      <button v-if="!isAuthenticated" @click="login({ name: 'User' })">Login (Demo)</button>
      <button v-else @click="logout">Logout</button>
    </div>
  </div>
</template>

<style scoped>
h1 {
  background: linear-gradient(to right, var(--primary), var(--secondary));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  font-size: 2.5rem;
  margin-bottom: 1rem;
}
</style>
