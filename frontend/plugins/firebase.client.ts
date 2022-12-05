import { initializeApp } from 'firebase/app'
import { getAuth, setPersistence, inMemoryPersistence } from 'firebase/auth'

export default defineNuxtPlugin(() => {
  const { app: { firebaseConfig } } = useRuntimeConfig()
  const app = initializeApp(firebaseConfig)

  const auth = getAuth(app)
  setPersistence(auth, inMemoryPersistence)

  return {
    provide: {
      firebase: app,
      auth,
    },
  }
})
