<script setup lang="ts">
const currentYear = new Date().getFullYear()
</script>

<template>
  <footer class="bg-gradient-to-r from-cyan-600 to-teal-600 text-white py-6 mt-auto">
    <div class="container mx-auto text-center">
      <p class="mb-1 opacity-90">© {{ currentYear }} Lazy Speech To Text Converter. All rights reserved.</p>
      <p class="text-sm opacity-75">
        Created by <a href="mailto:walter.fan@gmail.com" class="underline hover:text-white">walter.fan@gmail.com</a>
      </p>
    </div>
  </footer>
</template>

