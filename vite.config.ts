import { defineConfig } from 'vite'
import { resolve } from 'path'

export default defineConfig({
  root: './internal/web/static',
  build: {
    outDir: './js',
    rollupOptions: {
      input: {
        app: resolve('./internal/web/static/ts/app.ts')
      },
      output: {
        entryFileNames: '[name].js',
        chunkFileNames: '[name].js',
        assetFileNames: '[name].[ext]'
      }
    },
    emptyOutDir: true,
    minify: true
  },
  server: {
    port: 3000
  }
})