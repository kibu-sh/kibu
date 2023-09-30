import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  appType: 'spa',
  base: '/__admin/ui/',
  build: {
    minify: true,
  },
  server: {
    proxy: {
      '/__admin/api/snapshot/stream': {
        target: 'wss://127.0.0.1:9091',
        ws: true,
        secure: false,
        ssl: {},
      },
    },
  },
})
