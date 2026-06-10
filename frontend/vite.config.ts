import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// In Docker the backend is a separate service reachable as http://backend:8080,
// so the proxy target is configurable. Locally it defaults to localhost:8080.
const apiTarget = process.env.API_PROXY_TARGET ?? 'http://localhost:8080'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    host: true,
    proxy: {
      '/api': apiTarget,
    },
  },
})
