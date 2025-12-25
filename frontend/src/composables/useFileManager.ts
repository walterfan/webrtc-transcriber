import { ref, onMounted } from 'vue'
import type { TranscriptionResult } from './useWebRTC'

export interface FileGroup {
  baseName: string
  modTime: number
  audio_file?: string
  text_file?: string
  text?: string
  confidence?: number
}

export function useFileManager() {
  const files = ref<FileGroup[]>([])
  const loading = ref(false)

  const fetchFiles = async () => {
    loading.value = true
    try {
      const res = await fetch('/files')
      const rawFiles = await res.json()
      
      const groups: Record<string, FileGroup> = {}
      
      rawFiles.forEach((fileInfo: any) => {
        const f = fileInfo.name
        const modTime = fileInfo.modTime
        const baseName = f.substring(0, f.lastIndexOf('.'))
        const ext = f.substring(f.lastIndexOf('.') + 1)
        
        if (!groups[baseName]) {
          groups[baseName] = { baseName, modTime, text: baseName }
        }
        
        if (modTime > groups[baseName].modTime) {
          groups[baseName].modTime = modTime
        }
        
        if (ext === 'wav') {
          groups[baseName].audio_file = `recordings/${f}`
        } else if (ext === 'txt') {
          groups[baseName].text_file = `recordings/${f}`
          groups[baseName].text = "Transcription available"
        }
      })
      
      files.value = Object.values(groups)
        .filter(g => g.audio_file || g.text_file)
        .sort((a, b) => b.modTime - a.modTime)
        
    } catch (error) {
      console.error('Error fetching files:', error)
    } finally {
      loading.value = false
    }
  }

  const deleteFile = async (filename: string) => {
    try {
      await fetch(`/delete/${filename}`, { method: 'DELETE' })
      return true
    } catch (error) {
      console.error('Delete failed:', error)
      return false
    }
  }

  const deleteResult = async (group: FileGroup) => {
    const promises = []
    if (group.audio_file) {
      const audioName = group.audio_file.split(/[/\\]/).pop()
      if (audioName) promises.push(deleteFile(audioName))
    }
    if (group.text_file) {
      const textName = group.text_file.split(/[/\\]/).pop()
      if (textName) promises.push(deleteFile(textName))
    }
    
    await Promise.all(promises)
    files.value = files.value.filter(f => f.baseName !== group.baseName)
  }

  const transcribeFiles = async (filenames: string[], language = 'auto') => {
    try {
      const res = await fetch('/transcribe', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ files: filenames, language })
      })
      const results: TranscriptionResult[] = await res.json()
      
      // Update local state
      files.value = files.value.map(f => {
        const audioName = f.audio_file?.split(/[/\\]/).pop()
        const match = results.find(r => r.audio_file && r.audio_file.includes(audioName || ''))
        if (match) {
          return { ...f, ...match, text: match.text || "Transcription available" }
        }
        return f
      })
      
      return results
    } catch (error) {
      console.error('Transcription failed:', error)
      throw error
    }
  }

  const addResult = (result: TranscriptionResult) => {
    const fileName = result.audio_file ? result.audio_file.split(/[/\\]/).pop() : null
    if (!fileName) return

    const baseName = fileName.substring(0, fileName.lastIndexOf('.'))
    
    // Check if exists
    const existingIdx = files.value.findIndex(f => f.baseName === baseName)
    
    const newEntry: FileGroup = {
      baseName,
      modTime: Date.now(),
      audio_file: result.audio_file,
      text_file: result.text_file,
      text: result.text || baseName,
      confidence: result.confidence
    }

    if (existingIdx >= 0) {
      files.value[existingIdx] = { ...files.value[existingIdx], ...newEntry }
    } else {
      files.value = [newEntry, ...files.value]
    }
  }

  onMounted(() => {
    fetchFiles()
  })

  return {
    files,
    loading,
    fetchFiles,
    deleteResult,
    transcribeFiles,
    addResult
  }
}

