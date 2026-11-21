/// <reference types="vitest/config" />

import { fileURLToPath } from 'node:url'
import tailwindcss from '@tailwindcss/vite'
import { defineConfig, loadEnv } from 'vite'

const local = (path: string) => fileURLToPath(new URL(path, import.meta.url))

// A static multi-page site: each page is plain HTML (readable by search engines without
// JavaScript) with a small shared script.
export default defineConfig(({ mode }) => ({
  plugins: [tailwindcss()],
  build: {
    rollupOptions: {
      input: {
        home: local('./index.html'),
        docs: local('./docs/index.html'),
      },
    },
  },
  // fixed, next to the admin app's 5173 that the local demo links point at
  server: {
    port: 5174,
    strictPort: true,
    // the inquiry form posts to /api on the same origin, so in dev it reaches the Go backend
    proxy: {
      '/api': {
        target: loadEnv(mode, process.cwd(), '').VITE_API_PROXY_TARGET || 'http://localhost:3000',
        changeOrigin: true,
      },
    },
  },
  test: {
    environment: 'jsdom',
  },
}))
