import { fileURLToPath, URL } from 'node:url'
import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  const sageLocalDev = env.SAGE_LOCAL_DEV === 'true'

  return {
    plugins: [vue()],
    envPrefix: ['VITE_', 'VUE_APP_'],
    resolve: {
      extensions: ['.mjs', '.js', '.mts', '.ts', '.jsx', '.tsx', '.json', '.vue'],
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url)),
        '~@': fileURLToPath(new URL('./src', import.meta.url)),
        '@/app/utility/debounce': fileURLToPath(new URL('./src/app/utility/debounce.ts', import.meta.url)),
        path: 'path-browserify',
        querystring: 'querystring-es3',
        'highlight.js/lib/highlight': 'highlight.js/lib/core',
        'v-runtime-template': fileURLToPath(new URL('./src/components/runtime/TrustedHtml.vue', import.meta.url)),
      },
    },
    define: {
      __VUE_OPTIONS_API__: true,
      __VUE_PROD_DEVTOOLS__: false,
    },
    server: {
      host: '0.0.0.0',
      watch: {
        ignored: ['**/node_modules/**', '**/public/**'],
      },
      proxy: sageLocalDev
        ? {
            '/eqsage': {
              changeOrigin: true,
              target: 'http://127.0.0.1:4100',
              rewrite: (path) => path.replace(/^\/eqsage/, ''),
            },
            '/static': {
              changeOrigin: true,
              target: 'http://127.0.0.1:4100',
            },
          }
        : undefined,
    },
    build: {
      outDir: 'dist',
      emptyOutDir: true,
      sourcemap: false,
      chunkSizeWarningLimit: 40000,
      rollupOptions: {
        output: {
          entryFileNames: '[name].[hash].js',
          chunkFileNames: '[name].[hash].js',
          assetFileNames: '[name].[hash][extname]',
          manualChunks: {
            vendors: ['vue', 'vue-router', 'pinia', 'axios'],
          },
        },
      },
    },
  }
})
