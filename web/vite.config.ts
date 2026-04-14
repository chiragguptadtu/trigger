import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// Pass browser page-navigation requests through to index.html so React Router
// handles them. Only proxy requests that actually want JSON (API calls).
function bypass(req: { headers: { accept?: string } }) {
  if ((req.headers.accept ?? '').includes('text/html')) return '/index.html'
}

export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/auth':       { target: 'http://localhost:8080', bypass },
      '/commands':   { target: 'http://localhost:8080', bypass },
      '/executions': { target: 'http://localhost:8080', bypass },
      '/admin':      { target: 'http://localhost:8080', bypass },
    },
  },
})
