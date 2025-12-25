<script setup lang="ts">
import { ref } from 'vue'
import { LogIn, User, Lock } from 'lucide-vue-next'

const props = defineProps<{
  loading?: boolean
  error?: string
}>()

const emit = defineEmits<{
  (e: 'login', u: string, p: string): void
}>()

const username = ref('')
const password = ref('')

const handleSubmit = () => {
  emit('login', username.value, password.value)
}
</script>

<template>
  <div class="max-w-md mx-auto mt-20 p-10 bg-white rounded-lg shadow-xl">
    <div class="text-center mb-8">
      <div class="inline-flex items-center justify-center w-20 h-20 rounded-full bg-gradient-to-br from-cyan-600 to-teal-600 mb-6 shadow-lg">
        <LogIn class="w-10 h-10 text-white" />
      </div>
      <h2 class="text-2xl font-bold bg-gradient-to-r from-cyan-600 to-teal-600 bg-clip-text text-transparent">
        Welcome Back
      </h2>
      <p class="text-gray-500 mt-2">Sign in to continue</p>
    </div>

    <form @submit.prevent="handleSubmit">
      <div class="mb-4">
        <label class="block text-gray-700 text-sm font-bold mb-2">Username</label>
        <div class="relative">
          <input 
            v-model="username"
            type="text"
            required
            placeholder="Enter username"
            class="w-full pl-10 pr-4 py-3 rounded-lg border border-gray-300 focus:outline-none focus:border-cyan-500 focus:ring-1 focus:ring-cyan-500"
          />
          <User class="absolute left-3 top-3.5 w-5 h-5 text-cyan-600" />
        </div>
      </div>

      <div class="mb-6">
        <label class="block text-gray-700 text-sm font-bold mb-2">Password</label>
        <div class="relative">
          <input 
            v-model="password"
            type="password"
            required
            placeholder="Enter password"
            class="w-full pl-10 pr-4 py-3 rounded-lg border border-gray-300 focus:outline-none focus:border-cyan-500 focus:ring-1 focus:ring-cyan-500"
          />
          <Lock class="absolute left-3 top-3.5 w-5 h-5 text-cyan-600" />
        </div>
      </div>

      <div v-if="error" class="mb-6 p-4 bg-red-50 text-red-600 rounded-lg text-sm">
        {{ error }}
      </div>

      <button 
        type="submit"
        :disabled="loading"
        class="w-full py-3 px-4 bg-gradient-to-r from-cyan-600 to-teal-600 text-white font-bold rounded-lg shadow-md hover:from-cyan-700 hover:to-teal-700 focus:outline-none disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2 transition-all"
      >
        <span v-if="loading" class="w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin"></span>
        <span v-else>Sign In</span>
      </button>
    </form>
  </div>
</template>

