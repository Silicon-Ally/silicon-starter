import { initializeApp, getApp } from 'firebase-admin/app'
import { getAuth } from 'firebase-admin/auth'

export default defineNuxtPlugin(() => {
  const { app: { firebaseConfig } } = useRuntimeConfig()
  // Plugins run on every request, but Firebase errors out if it's initialized
  // more than once, so we only initialize it once per server start. We could
  // likely also use Lifecycle Hooks [1] for this, but this works fine.
  // [1] https://v3.nuxtjs.org/api/advanced/hooks/
  try {
    getApp()
  } catch {
    initializeApp(firebaseConfig)
  }

  return {
    provide: {
      sauth: getAuth(),
    },
  }
})
