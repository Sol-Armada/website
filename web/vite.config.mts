import { execSync } from 'node:child_process'
import { fileURLToPath, URL } from 'node:url'
import Vue from '@vitejs/plugin-vue'
import UnoCSS from 'unocss/vite'
import Fonts from 'unplugin-fonts/vite'
import { defineConfig } from 'vite'
import VueRouter from 'vue-router/vite'

const appVersion = (() => {
  if (process.env.VITE_APP_VERSION) {
    return process.env.VITE_APP_VERSION
  }

  try {
    return execSync('git describe --tags --abbrev=0', { encoding: 'utf8' }).trim()
  } catch {
    return 'dev'
  }
})()

export default defineConfig({
  envDir: '..',
  define: {
    __APP_VERSION__: JSON.stringify(appVersion),
  },
  plugins: [VueRouter({ dts: 'src/typed-router.d.ts' }), Vue(), Fonts({
    fontsource: {
      families: [
        {
          name: 'Roboto',
          weights: [100, 300, 400, 500, 700, 900],
          styles: ['normal', 'italic'],
        },
      ],
    },
  }), UnoCSS()],
  build: {
    outDir: '../api/dist',
    emptyOutDir: true,
  },
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('src', import.meta.url)),
    },
    extensions: ['.js', '.json', '.jsx', '.mjs', '.ts', '.tsx', '.vue'],
  },
  server: {
    port: 3000,
    proxy: {
      '/auth': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        ws: true,
      },
    },
  },
})
