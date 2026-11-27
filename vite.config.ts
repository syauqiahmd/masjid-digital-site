/// <reference types="vitest/config" />

import { fileURLToPath } from 'node:url'
import tailwindcss from '@tailwindcss/vite'
import { defineConfig, loadEnv } from 'vite'
import { seo } from './seo.ts'

const local = (path: string) => fileURLToPath(new URL(path, import.meta.url))

// A static multi-page site: each page is plain HTML (readable by search engines without
// JavaScript) with a small shared script.
export default defineConfig(({ mode }) => ({
  plugins: [tailwindcss(), seo(loadEnv(mode, process.cwd(), ''))],
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
    // the form and the server bersama numbers use /api on the same origin; in dev that is server/
    proxy: {
      '/api': {
        target: loadEnv(mode, process.cwd(), '').VITE_API_PROXY_TARGET || 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
  test: {
    environment: 'jsdom',
  },
}))
