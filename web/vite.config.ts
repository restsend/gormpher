import * as path from 'path'
import { defineConfig } from 'vite'
import Vue from '@vitejs/plugin-vue'
import Unocss from 'unocss/vite'
import { visualizer } from 'rollup-plugin-visualizer'
import Inspector from 'unplugin-vue-inspector/vite'

// https://vitejs.dev/config/
export default defineConfig({
  resolve: {
    alias: {
      '@': path.resolve(__dirname, 'src'),
    },
  },
  plugins: [
    Vue(),
    // https://github.com/unocss/unocss
    Unocss(), // unocss.config.ts
    visualizer({
      open: true,
      gzipSize: true,
      brotliSize: true,
    }),
    Inspector({
      toggleButtonVisibility: 'always',
    }),
  ],
  server: {
    host: '0.0.0.0',
    proxy: {
      '/api': {
        target: 'http://localhost:5000',
        changeOrigin: true,
      },
    },
  },
  build: {
    rollupOptions: {
      output: {
        entryFileNames: 'assets/[name].js',
        chunkFileNames: 'assets/[name].js',
        assetFileNames: 'assets/[name].[ext]',
      },
    },
  },
})
