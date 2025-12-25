import { onUnmounted, type Ref } from 'vue'

export function useAudioVisualization(stream: Ref<MediaStream | null>, canvasRef: Ref<HTMLCanvasElement | null>) {
  let animationId: number
  let audioCtx: AudioContext | null = null

  const startVisualizer = () => {
    if (!stream.value || !canvasRef.value) return

    const AudioContext = window.AudioContext || (window as any).webkitAudioContext
    audioCtx = new AudioContext()
    const analyser = audioCtx.createAnalyser()
    const source = audioCtx.createMediaStreamSource(stream.value)
    source.connect(analyser)

    analyser.fftSize = 2048
    const bufferLength = analyser.frequencyBinCount
    const dataArray = new Uint8Array(bufferLength)
    const canvas = canvasRef.value
    if (!canvas) return
    const canvasCtx = canvas.getContext('2d')

    if (!canvasCtx) return

      const draw = () => {
      animationId = requestAnimationFrame(draw)

      analyser.getByteTimeDomainData(dataArray)

      // Create gradient background
      const gradient = canvasCtx.createLinearGradient(0, 0, 0, canvas.height)
      gradient.addColorStop(0, '#f8fafc')
      gradient.addColorStop(1, '#e2e8f0')
      canvasCtx.fillStyle = gradient
      canvasCtx.fillRect(0, 0, canvas.width, canvas.height)

      // Create gradient for waveform
      const waveGradient = canvasCtx.createLinearGradient(0, 0, canvas.width, 0)
      waveGradient.addColorStop(0, '#0891b2')
      waveGradient.addColorStop(0.5, '#0d9488')
      waveGradient.addColorStop(1, '#0891b2')

      canvasCtx.lineWidth = 3
      canvasCtx.strokeStyle = waveGradient
      canvasCtx.lineCap = 'round'
      canvasCtx.lineJoin = 'round'

      canvasCtx.beginPath()

      const sliceWidth = canvas.width * 1.0 / bufferLength
      let x = 0

      for (let i = 0; i < bufferLength; i++) {
        const val = dataArray[i] ?? 128
        const v = val / 128.0
        const y = v * canvas.height / 2

        if (i === 0) {
          canvasCtx.moveTo(x, y)
        } else {
          canvasCtx.lineTo(x, y)
        }

        x += sliceWidth
      }

      canvasCtx.lineTo(canvas.width, canvas.height / 2)
      canvasCtx.stroke()
    }

    draw()
  }

  const stopVisualizer = () => {
    if (animationId) cancelAnimationFrame(animationId)
    if (audioCtx) {
      audioCtx.close()
      audioCtx = null
    }
  }

  onUnmounted(() => {
    stopVisualizer()
  })

  return {
    startVisualizer,
    stopVisualizer
  }
}

