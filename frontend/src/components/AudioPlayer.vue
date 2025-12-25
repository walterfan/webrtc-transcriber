<script setup lang="ts">
import { ref, onUnmounted } from 'vue'
import { Play, Pause, FileAudio } from 'lucide-vue-next'

const props = defineProps<{
  src: string
  fileName?: string
}>()

const audio = ref<HTMLAudioElement | null>(null)
const isPlaying = ref(false)
const currentTime = ref(0)
const duration = ref(0)
const isDragging = ref(false)

const formatTime = (time: number) => {
  if (isNaN(time) || time === 0) return '0:00'
  const mins = Math.floor(time / 60)
  const secs = Math.floor(time % 60)
  return `${mins}:${secs.toString().padStart(2, '0')}`
}

const togglePlay = () => {
  if (!audio.value) return
  if (isPlaying.value) {
    audio.value.pause()
  } else {
    audio.value.play()
  }
  isPlaying.value = !isPlaying.value
}

const onTimeUpdate = () => {
  if (audio.value && !isDragging.value) {
    currentTime.value = audio.value.currentTime
  }
}

const onLoadedMetadata = () => {
  if (audio.value) {
    duration.value = audio.value.duration
  }
}

const onEnded = () => {
  isPlaying.value = false
  currentTime.value = 0
}

const handleProgressClick = (e: MouseEvent) => {
  const bar = e.currentTarget as HTMLElement
  const rect = bar.getBoundingClientRect()
  const pos = (e.clientX - rect.left) / rect.width
  if (audio.value) {
    const newTime = pos * duration.value
    audio.value.currentTime = newTime
    currentTime.value = newTime
  }
}

onUnmounted(() => {
  if (audio.value) {
    audio.value.pause()
    audio.value.src = ''
  }
})
</script>

<template>
  <div class="w-full max-w-xs">
    <audio 
      ref="audio" 
      :src="src" 
      preload="metadata"
      @timeupdate="onTimeUpdate"
      @loadedmetadata="onLoadedMetadata"
      @ended="onEnded"
    ></audio>

    <div v-if="fileName" class="flex items-center gap-2 mb-2 text-sm text-gray-600">
      <FileAudio class="w-4 h-4" />
      <span class="truncate">{{ fileName }}</span>
    </div>

    <div class="bg-gradient-to-br from-cyan-50 to-teal-50 rounded-xl p-3 border border-teal-100">
      <!-- Progress Bar -->
      <div 
        class="relative w-full h-2 bg-gray-200 rounded-full cursor-pointer mb-2"
        @click="handleProgressClick"
      >
        <div 
          class="absolute top-0 left-0 h-full bg-gradient-to-r from-cyan-600 to-teal-600 rounded-full pointer-events-none"
          :style="{ width: `${(currentTime / duration) * 100}%` }"
        ></div>
        <div 
          class="absolute top-1/2 -mt-2 w-4 h-4 bg-cyan-600 rounded-full shadow-md cursor-grab hover:scale-110 transition-transform"
          :style="{ left: `calc(${(currentTime / duration) * 100}% - 8px)` }"
        ></div>
      </div>

      <!-- Controls -->
      <div class="flex items-center justify-between">
        <button 
          @click="togglePlay"
          class="w-8 h-8 flex items-center justify-center rounded-full text-white shadow-sm transition-colors"
          :class="isPlaying ? 'bg-amber-500 hover:bg-amber-600' : 'bg-cyan-600 hover:bg-cyan-700'"
        >
          <Pause v-if="isPlaying" class="w-4 h-4" />
          <Play v-else class="w-4 h-4 ml-0.5" />
        </button>
        
        <span class="text-xs font-mono text-teal-700">
          {{ formatTime(currentTime) }} / {{ formatTime(duration) }}
        </span>
      </div>
    </div>
  </div>
</template>

