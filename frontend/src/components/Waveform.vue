<script setup lang="ts">
import { ref } from 'vue'

const props = defineProps<{
  stream: MediaStream | null
}>()

const canvas = ref<HTMLCanvasElement | null>(null)

import { useAudioVisualization } from '../composables/useAudioVisualization'
import { toRef } from 'vue'

useAudioVisualization(toRef(props, 'stream'), canvas)
</script>

<template>
  <div class="my-4">
    <div class="flex items-center gap-2 mb-2">
      <div class="w-8 h-8 rounded-full bg-cyan-100 flex items-center justify-center">
        <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5 text-cyan-600" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M2 12h20M2 12c0 5.5 4.5 10 10 10s10-4.5 10-10S17.5 2 12 2 2 6.5 2 12z"></path>
          <path d="M2 12h20"></path>
        </svg>
      </div>
      <h4 class="font-semibold text-gray-700">Live Audio Waveform</h4>
    </div>
    <canvas 
      ref="canvas" 
      width="800" 
      height="120"
      class="w-full h-[120px] rounded-xl shadow-inner bg-slate-50 border border-slate-200"
    ></canvas>
  </div>
</template>

