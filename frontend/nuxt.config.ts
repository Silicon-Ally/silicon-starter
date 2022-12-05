import { defineNuxtConfig } from 'nuxt/config'
import dotenv from 'dotenv'

dotenv.config({ path: process.env.DOTENV_PATH })

export default defineNuxtConfig({
  app: {
    head: {
      htmlAttrs: {
        lang: 'en',
      },
      title: 'Silicon Starter',
      meta: [
        { charset: 'utf-8' },
        { name: 'viewport', content: 'width=device-width,initial-scale=1' },
        { name: 'description', content: 'A fast and performant web appp stack.' },
        { name: 'theme-color', content: '#0F6CC2' },
      ],
      link: [
        { rel: 'manifest', href: 'manifest.json' },
      ],
    },
  },
  runtimeConfig: {
    public: {},
    app: {
      clientsUseFullURL: process.env.CLIENTS_USE_FULL_URL === 'TRUE' || process.env.LOCAL_PREVIEW_MODE === 'true',
      baseServerURL: {
        endpoint: process.env.BASE_SERVER_ENDPOINT ?? '',
        path: process.env.BASE_SERVER_PATH ?? '',
      },
      graphQLServerURL: {
        endpoint: process.env.GRAPHQL_SERVER_ENDPOINT ?? '',
        path: process.env.GRAPHQL_SERVER_PATH ?? '',
      },
      firebaseConfig: {
        apiKey: process.env.FIREBASE_API_KEY ?? '',
        appId: process.env.FIREBASE_APP_ID ?? '',
        authDomain: process.env.FIREBASE_AUTH_DOMAIN ?? '',
        projectId: process.env.FIREBASE_PROJECT_ID ?? '',
        storageBucket: process.env.FIREBASE_STORAGE_BUCKET ?? '',
      },
    },
  },
  typescript: {
    strict: true,
  },
  css: [
    '@/assets/scss/theme.scss',
  ],
  nitro: {
    devProxy: {
      '/api/': 'http://localhost:8080/api/',
    },
  },
})
