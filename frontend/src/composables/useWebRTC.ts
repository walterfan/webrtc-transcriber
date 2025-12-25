import { ref, onUnmounted } from 'vue'

export interface RtcState {
  active: boolean
  processing: boolean
  offer: string | null
  answer: string | null
  error: string | null
  recordingDuration: number
  stats: {
    codec: string
    transport: string
  }
}

export interface TranscriptionResult {
  audio_file: string
  text_file?: string
  text?: string
  confidence?: number
}

interface WebRTCOptions {
  onResult: (result: TranscriptionResult) => void
}

export function useWebRTC(options?: WebRTCOptions) {
  const rtcState = ref<RtcState>({
    active: false,
    processing: false,
    offer: null,
    answer: null,
    error: null,
    recordingDuration: 0,
    stats: { codec: '-', transport: '-' }
  })

  let pc: RTCPeerConnection | null = null
  let stream: MediaStream | null = null
  let statsInterval: number | null = null

  // Helper to start the session with the server
  const startSession = async (offer: string, language: string, enableTranscribe: boolean) => {
    const res = await fetch('/session', {
      method: 'POST',
      body: JSON.stringify({
        offer,
        language,
        transcribe: enableTranscribe
      }),
      headers: { 'Content-Type': 'application/json' }
    })
    
    if (res.status === 401) throw new Error('Unauthorized')
    
    const msg = await res.json()
    return msg.answer
  }

  const decodeDataChannelPayload = async (data: any): Promise<string> => {
    if (data instanceof ArrayBuffer) {
      return new TextDecoder('utf-8').decode(data)
    } else if (data instanceof Blob) {
      return await data.text()
    }
    return ''
  }

  const start = async (deviceId: string, language = 'auto', enableTranscribe = true) => {
    // Reset state
    rtcState.value = {
      ...rtcState.value,
      offer: null,
      answer: null,
      error: null,
      recordingDuration: 0,
      stats: { codec: '-', transport: '-' }
    }

    try {
      const constraints = deviceId ? { audio: { deviceId: { exact: deviceId } } } : { audio: true }
      stream = await navigator.mediaDevices.getUserMedia(constraints)
      
      pc = new RTCPeerConnection({
        iceServers: [{ urls: 'stun:stun.l.google.com:19302' }]
      })

      const resChan = pc.createDataChannel('results', { ordered: true, protocol: 'tcp' })
      
      resChan.onmessage = async (evt) => {
        const strData = await decodeDataChannelPayload(evt.data)
        const result = JSON.parse(strData)
        if (options?.onResult) {
          options.onResult(result)
        }
        // If we get a result, we assume processing for that chunk is done
        rtcState.value.processing = false
      }

      resChan.onclose = () => {
        stop()
      }

      pc.onicecandidate = async (evt) => {
        if (!evt.candidate && pc?.localDescription) {
          const { sdp: offer } = pc.localDescription
          rtcState.value.offer = offer
          
          try {
            const answer = await startSession(offer, language, enableTranscribe)
            rtcState.value.answer = answer
            
            const rd = new RTCSessionDescription({
              sdp: answer,
              type: 'answer'
            })
            await pc.setRemoteDescription(rd)
          } catch (err: any) {
            rtcState.value.error = err.message || 'Session start failed'
          }
        }
      }

      // Add tracks
      stream.getAudioTracks().forEach(track => {
        if (pc) pc.addTrack(track, stream!)
      })

      // Create offer
      const ld = await pc.createOffer({
        offerToReceiveAudio: false,
        offerToReceiveVideo: false
      })
      await pc.setLocalDescription(ld)

      rtcState.value.active = true
      rtcState.value.processing = false
      
      // Start stats interval
      const startTime = Date.now()
      statsInterval = window.setInterval(async () => {
        if (pc && rtcState.value.active) {
          const duration = Math.floor((Date.now() - startTime) / 1000)
          const statsReport = await pc.getStats()
          
          let codec = '-'
          let transport = '-'
          
          statsReport.forEach(report => {
            if (report.type === 'outbound-rtp' && report.mediaType === 'audio') {
              const codecReport = statsReport.get(report.codecId)
              if (codecReport && codecReport.mimeType) {
                codec = codecReport.mimeType.split('/')[1] || codecReport.mimeType
              }
            }
            if (report.type === 'candidate-pair' && report.state === 'succeeded') {
              const localCandidate = statsReport.get(report.localCandidateId)
              if (localCandidate && localCandidate.protocol) {
                transport = localCandidate.protocol.toUpperCase()
              }
            }
          })

          rtcState.value.recordingDuration = duration
          rtcState.value.stats = { codec, transport }
        }
      }, 1000)

    } catch (err: any) {
      console.error('Error starting WebRTC:', err)
      rtcState.value.error = err.message || 'Could not access microphone'
      rtcState.value.processing = false
    }
  }

  const stop = () => {
    if (statsInterval) {
      clearInterval(statsInterval)
      statsInterval = null
    }

    if (stream) {
      stream.getAudioTracks().forEach(tr => tr.stop())
      stream = null
    }

    if (pc) {
      pc.close()
      pc = null
    }

    if (rtcState.value.active) {
      rtcState.value.active = false
      rtcState.value.processing = true
      
      // Auto-timeout for processing state
      setTimeout(() => {
        if (rtcState.value.processing) {
          rtcState.value.processing = false
        }
      }, 30000)
    }
  }

  onUnmounted(() => {
    stop()
  })

  return {
    rtcState,
    start,
    stop,
    getStream: () => stream // expose stream for waveform
  }
}

