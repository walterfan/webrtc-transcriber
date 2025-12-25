<script setup lang="ts">
import { computed } from 'vue'
import { FileText, Download, Trash2, CheckCircle, Eye, Loader2 } from 'lucide-vue-next'
import AudioPlayer from './AudioPlayer.vue'
import type { FileGroup } from '../composables/useFileManager'

const props = defineProps<{
  files: FileGroup[]
  processing?: boolean
  selectable?: boolean
  selectedFiles: string[]
}>()

const emit = defineEmits<{
  (e: 'delete', file: FileGroup): void
  (e: 'select', fileName: string): void
}>()

const untranscribedCount = computed(() => 
  props.files.filter(f => f.audio_file && !f.text_file).length
)
</script>

<template>
  <div class="mt-8">
    <div class="flex items-center justify-between mb-4">
      <div class="flex items-center gap-2">
        <FileText class="w-5 h-5 text-cyan-600" />
        <h3 class="text-lg font-bold text-gray-700">Transcription Results</h3>
        <span v-if="processing" class="flex items-center gap-2 text-sm text-cyan-600 ml-2">
          <Loader2 class="w-4 h-4 animate-spin" />
          Processing...
        </span>
      </div>
    </div>

    <div v-if="selectable && untranscribedCount > 0" class="mb-4 p-3 bg-emerald-50 border border-emerald-200 rounded-lg flex items-center gap-2 text-emerald-700 text-sm">
      <CheckCircle class="w-5 h-5" />
      <span>{{ untranscribedCount }} file(s) available for transcription. Select files to transcribe.</span>
    </div>

    <div v-if="files.length === 0 && !processing" class="text-center py-12 bg-gray-50 rounded-xl border border-dashed border-gray-300">
      <div class="inline-flex items-center justify-center w-16 h-16 rounded-full bg-gray-100 mb-4">
        <FileText class="w-8 h-8 text-gray-400" />
      </div>
      <p class="text-gray-500">No recordings yet. Click Start to begin.</p>
    </div>

    <div v-else class="overflow-x-auto rounded-xl border border-gray-200 shadow-sm">
      <table class="w-full text-left text-sm">
        <thead class="bg-gradient-to-r from-cyan-600 to-teal-600 text-white">
          <tr>
            <th v-if="selectable" class="p-3 w-10 text-center">
              <input type="checkbox" disabled class="rounded border-gray-300 opacity-50 cursor-not-allowed" />
            </th>
            <th class="p-3 font-semibold w-12">#</th>
            <th class="p-3 font-semibold">Transcription</th>
            <th class="p-3 font-semibold w-64">Audio File</th>
            <th class="p-3 font-semibold w-48">Text File</th>
            <th class="p-3 font-semibold w-24">Actions</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-100 bg-white">
          <tr v-if="processing" class="bg-cyan-50 animate-pulse">
            <td v-if="selectable" class="p-3">-</td>
            <td class="p-3 font-mono text-cyan-700">{{ files.length + 1 }}</td>
            <td class="p-3">
              <div class="flex items-center gap-3">
                <Loader2 class="w-5 h-5 text-cyan-600 animate-spin" />
                <div>
                  <div class="font-medium text-cyan-700">Transcribing audio...</div>
                  <div class="text-xs text-gray-500">Please wait...</div>
                </div>
              </div>
            </td>
            <td class="p-3 text-gray-400">Processing...</td>
            <td class="p-3 text-gray-400">Processing...</td>
            <td class="p-3 text-gray-400">-</td>
          </tr>

          <tr v-for="(file, idx) in files" :key="file.baseName" class="hover:bg-gray-50 transition-colors" :class="{'bg-emerald-50': selectedFiles.includes(file.audio_file?.split('/').pop() || '')}">
            <td v-if="selectable" class="p-3 text-center align-middle">
              <div v-if="file.audio_file && !file.text_file">
                <input 
                  type="checkbox" 
                  :checked="selectedFiles.includes(file.audio_file.split('/').pop() || '')"
                  @change="emit('select', file.audio_file?.split('/').pop() || '')"
                  class="w-4 h-4 rounded border-gray-300 text-cyan-600 focus:ring-cyan-500 cursor-pointer"
                />
              </div>
              <div v-else class="flex justify-center">
                <div class="w-5 h-5 rounded bg-emerald-100 flex items-center justify-center" title="Already transcribed">
                  <CheckCircle class="w-3 h-3 text-emerald-600" />
                </div>
              </div>
            </td>
            <td class="p-3 font-mono text-gray-500 align-middle">{{ idx + 1 }}</td>
            <td class="p-3 align-middle max-w-md">
              <div v-if="file.text_file">
                <div class="font-medium text-gray-800 line-clamp-2">{{ file.text }}</div>
                <div class="text-xs text-gray-400 mt-1">Confidence: {{ ((file.confidence || 0) * 100).toFixed(1) }}%</div>
              </div>
              <div v-else class="text-gray-400 italic">Not transcribed yet</div>
            </td>
            <td class="p-3 align-middle">
              <div v-if="file.audio_file">
                <AudioPlayer :src="'/' + file.audio_file" :file-name="file.audio_file.split('/').pop()" />
              </div>
              <span v-else class="text-gray-300">-</span>
            </td>
            <td class="p-3 align-middle">
              <div v-if="file.text_file" class="space-y-2">
                <div class="flex items-center gap-1 text-xs text-gray-500 mb-1">
                  <FileText class="w-3 h-3" />
                  <span class="truncate max-w-[150px]">{{ file.text_file.split('/').pop() }}</span>
                </div>
                <div class="flex gap-2">
                  <a :href="'/' + file.text_file" target="_blank" class="px-2 py-1 bg-blue-50 text-blue-600 rounded text-xs hover:bg-blue-100 flex items-center gap-1">
                    <Eye class="w-3 h-3" /> View
                  </a>
                  <a :href="'/' + file.text_file" download class="px-2 py-1 bg-emerald-50 text-emerald-600 rounded text-xs hover:bg-emerald-100 flex items-center gap-1">
                    <Download class="w-3 h-3" /> DL
                  </a>
                </div>
              </div>
              <span v-else class="text-gray-300">-</span>
            </td>
            <td class="p-3 align-middle">
              <button 
                @click="emit('delete', file)"
                class="p-2 text-red-500 hover:bg-red-50 rounded-full transition-colors"
                title="Delete files"
              >
                <Trash2 class="w-4 h-4" />
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

