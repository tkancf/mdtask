import { defineConfig } from 'vite'
import { fileURLToPath } from 'url'
import path from 'path'

const __dirname = path.dirname(fileURLToPath(import.meta.url))

export default defineConfig({
  root: './internal/web/static',
  build: {
    outDir: './js',
    rollupOptions: {
      input: {
        app: path.resolve(__dirname, 'internal/web/static/ts/app.ts')
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