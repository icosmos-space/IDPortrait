import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// Dual access: Wails loads Vite directly; LAN browsers load via Echo HTTPS reverse-proxy.
// Never pin server.origin / hmr.host to 127.0.0.1 — remote clients would request their own localhost.
export default defineConfig({
  plugins: [vue()],
  server: {
    host: '127.0.0.1',
    port: 5173,
    strictPort: true,
    hmr: {
      // Client uses the current page host/protocol (ws on Wails, wss via Echo).
      path: '/vite-hmr',
    },
  },
})
