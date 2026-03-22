import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/ws/chat': {
        target: 'ws://localhost:8100',
        ws: true,
        changeOrigin: true,
      },
      '/ws/tasks': {
        target: 'ws://localhost:8100',
        ws: true,
        changeOrigin: true,
      },
      '/api': {
        target: 'http://localhost:8100',
        changeOrigin: true,
      },
    },
  },
})
