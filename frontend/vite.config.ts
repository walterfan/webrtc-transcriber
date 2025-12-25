import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue()],
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:9070',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, '')
      },
      // Proxy these specific endpoints directly to the backend since they don't have /api prefix in Go code
      '/session': 'http://localhost:9070',
      '/files': 'http://localhost:9070',
      '/login': 'http://localhost:9070',
      '/logout': 'http://localhost:9070',
      '/auth': 'http://localhost:9070',
      '/delete': 'http://localhost:9070',
      '/transcribe': 'http://localhost:9070',
      '/recordings': 'http://localhost:9070',
    }
  },
  build: {
    outDir: 'dist',
    emptyOutDir: true
  }
})
