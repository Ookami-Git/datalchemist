import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vitejs.dev/config/
export default defineConfig({
  
  optimizeDeps: {
    include: ["dagre", "mermaid"]
  },
  build: {
      rollupOptions: {
          output: {
              manualChunks: {
                  mermaid: ["mermaid"]
              }
          }
      }
  },
  css: {
    preprocessorOptions: {
       scss: {
         silenceDeprecations: [
           'import',
           'mixed-decls',
           'color-functions',
           'global-builtin',
         ],
       },
    },
  },
  plugins: [
    vue(),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  },
  base: './',
})