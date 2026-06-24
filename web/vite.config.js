import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vitejs.dev/config/
export default defineConfig({
  // Bind mounts (Docker Desktop, remote workspaces and AI edits) do not always
  // propagate filesystem events to the container. Polling keeps HMR reliable
  // regardless of where the file change originated.
  server: {
    watch: {
      usePolling: true,
      interval: 250,
    },
  },
  
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
