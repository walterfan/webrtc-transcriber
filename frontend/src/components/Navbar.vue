<script setup lang="ts">
import { Mic, User, LogOut } from 'lucide-vue-next'

defineProps<{
  username: string
}>()

const emit = defineEmits<{
  (e: 'logout'): void
}>()
</script>

<template>
  <nav class="fixed top-0 w-full bg-white border-b border-gray-100 z-50 px-4 py-2 shadow-sm">
    <div class="container mx-auto flex items-center justify-between">
      <div class="flex items-center gap-3">
        <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-cyan-600 to-teal-600 flex items-center justify-center shadow-md">
          <Mic class="w-6 h-6 text-white" />
        </div>
        <span class="text-xl font-bold bg-gradient-to-r from-cyan-600 to-teal-600 bg-clip-text text-transparent">
          Lazy Speech To Text
        </span>
      </div>

      <div class="flex items-center gap-4">
        <div class="flex items-center gap-2 px-3 py-1.5 bg-gray-50 rounded-full text-sm text-gray-600 border border-gray-200">
          <User class="w-4 h-4" />
          <span>Welcome, {{ username }}</span>
        </div>
        <button 
          @click="emit('logout')"
          class="flex items-center gap-2 px-3 py-1.5 text-sm text-gray-600 hover:bg-gray-100 rounded-full transition-colors"
        >
          <LogOut class="w-4 h-4" />
          Logout
        </button>
      </div>
    </div>
  </nav>
</template>

