<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { Mic, Globe, StopCircle, Settings2, Loader2, Languages, CheckSquare } from 'lucide-vue-next'
import LoginForm from './components/LoginForm.vue'
import Navbar from './components/Navbar.vue'
import Footer from './components/Footer.vue'
import Waveform from './components/Waveform.vue'
import FileTable from './components/FileTable.vue'
import { useAuth } from './composables/useAuth'
import { useWebRTC } from './composables/useWebRTC'
import { useFileManager } from './composables/useFileManager'

// Composables
const { authState, login, logout } = useAuth()
const { files, deleteResult, transcribeFiles, addResult } = useFileManager()
const { rtcState, start, stop, getStream } = useWebRTC({
  onResult: (result) => {
    addResult(result)
  }
})

// State
const devices = ref<MediaDeviceInfo[]>([])
const selectedDeviceId = ref('')
const selectedLanguage = ref('auto')
const enableRecord = ref(true)
const enableTranscribe = ref(true)
const selectedFiles = ref<string[]>([])
const transcribingFiles = ref(false)

// Options
const languages = [
  { code: 'auto', name: 'Auto Detect', flag: '🌐' },
  { code: 'en', name: 'English', flag: '🇺🇸' },
  { code: 'zh', name: 'Chinese (中文)', flag: '🇨🇳' },
  { code: 'ja', name: 'Japanese (日本語)', flag: '🇯🇵' },
  { code: 'ko', name: 'Korean (한국어)', flag: '🇰🇷' },
  { code: 'es', name: 'Spanish (Español)', flag: '🇪🇸' },
  { code: 'fr', name: 'French (Français)', flag: '🇫🇷' },
  { code: 'de', name: 'German (Deutsch)', flag: '🇩🇪' },
  // Add more as needed
]

// Fetch devices
onMounted(async () => {
  try {
    const devs = await navigator.mediaDevices.enumerateDevices()
    devices.value = devs.filter(d => d.kind === 'audioinput')
  } catch (e) {
    console.error('Error fetching devices', e)
  }
})

// Actions
const handleLogin = async (u: string, p: string) => {
  await login(u, p)
}

const toggleAction = () => {
  if (rtcState.value.active) {
    stop()
  } else {
    if (!enableRecord.value && !enableTranscribe.value) {
      alert('Please select at least one option: Record or Transcribe')
      return
    }
    start(selectedDeviceId.value, selectedLanguage.value, enableTranscribe.value)
  }
}

const handleSelectFile = (fileName: string) => {
  if (selectedFiles.value.includes(fileName)) {
    selectedFiles.value = selectedFiles.value.filter(f => f !== fileName)
  } else {
    selectedFiles.value.push(fileName)
  }
}

const handleTranscribeSelected = async () => {
  if (selectedFiles.value.length === 0) return
  
  transcribingFiles.value = true
  try {
    await transcribeFiles(selectedFiles.value, selectedLanguage.value)
    selectedFiles.value = []
  } finally {
    transcribingFiles.value = false
  }
}

// Watchers for "Transcribe Only" mode auto-select
watch([enableRecord, enableTranscribe], ([rec, trans]) => {
  if (!rec && trans) {
    // Auto select first untranscribed
    const first = files.value.find(f => f.audio_file && !f.text_file)
    if (first && first.audio_file) {
      const name = first.audio_file.split('/').pop()
      if (name) selectedFiles.value = [name]
    }
  } else {
    selectedFiles.value = []
  }
})

const formatDuration = (sec: number) => {
  const m = Math.floor(sec / 60).toString().padStart(2, '0')
  const s = (sec % 60).toString().padStart(2, '0')
  return `${m}:${s}`
}
</script>

<template>
  <div class="min-h-screen flex flex-col bg-gray-50 font-sans text-gray-800">
    <!-- Auth Loading -->
    <div v-if="authState.checking" class="flex-1 flex items-center justify-center">
      <div class="animate-spin w-12 h-12 border-4 border-cyan-200 border-t-cyan-600 rounded-full"></div>
    </div>

    <!-- Login -->
    <div v-else-if="!authState.authenticated" class="flex-1 flex flex-col">
      <div class="container mx-auto px-4 py-8">
        <header class="text-center mb-8">
          <h1 class="text-4xl font-extrabold bg-gradient-to-r from-cyan-600 to-teal-600 bg-clip-text text-transparent mb-2">
            Lazy Speech To Text
          </h1>
          <p class="text-gray-500 text-lg">Convert your voice to text effortlessly</p>
        </header>
        <LoginForm @login="handleLogin" />
      </div>
    </div>

    <!-- Dashboard -->
    <div v-else class="flex-1 flex flex-col">
      <Navbar :username="authState.username" @logout="logout" />

      <main class="flex-1 container mx-auto px-4 pt-24 pb-8">
        <div class="bg-white rounded-2xl shadow-xl p-6 md:p-8">
          
          <!-- Controls Area -->
          <div class="bg-gradient-to-br from-cyan-50 to-teal-50 rounded-xl p-6 mb-6 border border-teal-100">
            <!-- Checkboxes -->
            <div class="flex flex-wrap gap-6 mb-6">
              <label class="flex items-center gap-3 cursor-pointer group">
                <div class="relative flex items-center">
                  <input type="checkbox" v-model="enableRecord" :disabled="rtcState.active" class="peer h-5 w-5 cursor-pointer appearance-none rounded border border-gray-300 transition-all checked:border-cyan-600 checked:bg-cyan-600 disabled:opacity-50" />
                  <div class="pointer-events-none absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 text-white opacity-0 peer-checked:opacity-100">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" /></svg>
                  </div>
                </div>
                <div class="flex items-center gap-2 text-teal-800 font-medium group-hover:text-cyan-700 transition-colors">
                  <Mic class="w-5 h-5" /> Record Audio
                </div>
              </label>

              <label class="flex items-center gap-3 cursor-pointer group">
                <div class="relative flex items-center">
                  <input type="checkbox" v-model="enableTranscribe" :disabled="rtcState.active" class="peer h-5 w-5 cursor-pointer appearance-none rounded border border-gray-300 transition-all checked:border-cyan-600 checked:bg-cyan-600 disabled:opacity-50" />
                  <div class="pointer-events-none absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 text-white opacity-0 peer-checked:opacity-100">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" /></svg>
                  </div>
                </div>
                <div class="flex items-center gap-2 text-teal-800 font-medium group-hover:text-cyan-700 transition-colors">
                  <Languages class="w-5 h-5" /> Transcribe Audio
                </div>
              </label>
            </div>

            <!-- Main Action Bar -->
            <div class="flex flex-col md:flex-row gap-4 items-end md:items-center">
              <!-- Start/Stop Button -->
              <button 
                @click="toggleAction"
                :disabled="rtcState.processing"
                class="w-full md:w-auto h-12 px-6 rounded-lg font-bold shadow-md transition-all flex items-center justify-center gap-2 min-w-[140px]"
                :class="[
                  rtcState.active 
                    ? 'bg-gradient-to-r from-rose-500 to-red-600 text-white hover:from-rose-600 hover:to-red-700 animate-pulse' 
                    : rtcState.processing
                      ? 'bg-amber-500 text-white cursor-not-allowed'
                      : 'bg-gradient-to-r from-teal-500 to-cyan-600 text-white hover:from-teal-600 hover:to-cyan-700'
                ]"
              >
                <template v-if="rtcState.processing">
                  <Loader2 class="w-5 h-5 animate-spin" />
                  Processing...
                </template>
                <template v-else-if="rtcState.active">
                  <StopCircle class="w-5 h-5" />
                  Stop
                </template>
                <template v-else>
                  <Mic class="w-5 h-5" />
                  Start
                </template>
              </button>

              <!-- Device Selector -->
              <div class="w-full md:flex-1">
                <div class="relative">
                  <select 
                    v-model="selectedDeviceId" 
                    :disabled="rtcState.active"
                    class="w-full h-12 pl-4 pr-10 rounded-lg border border-gray-200 bg-white focus:ring-2 focus:ring-cyan-500 focus:border-cyan-500 appearance-none disabled:bg-gray-100"
                  >
                    <option value="">Select Microphone...</option>
                    <option v-for="d in devices" :key="d.deviceId" :value="d.deviceId">{{ d.label || `Microphone ${d.deviceId.slice(0,5)}` }}</option>
                  </select>
                  <Settings2 class="absolute right-3 top-3.5 w-5 h-5 text-gray-400 pointer-events-none" />
                </div>
              </div>

              <!-- Language Selector -->
              <div class="w-full md:w-56">
                <div class="relative">
                  <select 
                    v-model="selectedLanguage" 
                    :disabled="rtcState.active"
                    class="w-full h-12 pl-4 pr-10 rounded-lg border border-gray-200 bg-white focus:ring-2 focus:ring-cyan-500 focus:border-cyan-500 appearance-none disabled:bg-gray-100"
                  >
                    <option v-for="l in languages" :key="l.code" :value="l.code">{{ l.flag }} {{ l.name }}</option>
                  </select>
                  <Globe class="absolute right-3 top-3.5 w-5 h-5 text-gray-400 pointer-events-none" />
                </div>
              </div>
            </div>
          </div>

          <!-- Stats & Waveform -->
          <div v-if="rtcState.active || rtcState.recordingDuration > 0" class="mb-8">
            <div class="grid grid-cols-3 gap-4 mb-4">
              <div class="bg-gray-50 rounded-lg p-3 text-center border border-gray-100">
                <div class="text-xs text-gray-500 uppercase font-semibold">Time</div>
                <div class="text-xl font-mono text-cyan-700">{{ formatDuration(rtcState.recordingDuration) }}</div>
              </div>
              <div class="bg-gray-50 rounded-lg p-3 text-center border border-gray-100">
                <div class="text-xs text-gray-500 uppercase font-semibold">Codec</div>
                <div class="text-xl font-mono text-cyan-700">{{ rtcState.stats.codec }}</div>
              </div>
              <div class="bg-gray-50 rounded-lg p-3 text-center border border-gray-100">
                <div class="text-xs text-gray-500 uppercase font-semibold">Transport</div>
                <div class="text-xl font-mono text-cyan-700">{{ rtcState.stats.transport }}</div>
              </div>
            </div>
            
            <Waveform v-if="rtcState.active" :stream="getStream()" />
          </div>

          <!-- Transcribe Selected Action -->
          <div v-if="!enableRecord && enableTranscribe && selectedFiles.length > 0" class="flex justify-end mb-4">
             <button 
              @click="handleTranscribeSelected"
              :disabled="transcribingFiles"
              class="px-4 py-2 bg-gradient-to-r from-cyan-600 to-teal-600 text-white rounded-lg shadow-sm hover:from-cyan-700 hover:to-teal-700 flex items-center gap-2 disabled:opacity-50"
            >
              <Loader2 v-if="transcribingFiles" class="w-4 h-4 animate-spin" />
              <CheckSquare v-else class="w-4 h-4" />
              Transcribe Selected ({{ selectedFiles.length }})
            </button>
          </div>

          <!-- Results Table -->
          <FileTable 
            :files="files" 
            :processing="rtcState.processing || transcribingFiles" 
            :selectable="!enableRecord && enableTranscribe"
            :selected-files="selectedFiles"
            @select="handleSelectFile"
            @delete="deleteResult"
          />

          <!-- Debug Info (Optional/Collapsible) -->
          <div class="mt-8 border-t pt-4">
             <details class="text-xs text-gray-400 cursor-pointer">
               <summary>Debug Info (SDP)</summary>
               <pre class="mt-2 p-2 bg-gray-100 rounded overflow-auto max-h-40">{{ rtcState.offer }}</pre>
               <pre class="mt-2 p-2 bg-gray-100 rounded overflow-auto max-h-40">{{ rtcState.answer }}</pre>
             </details>
          </div>

        </div>
      </main>

      <Footer />
    </div>
  </div>
</template>
