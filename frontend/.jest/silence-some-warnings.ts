import {beforeAll, afterAll} from '@jest/globals'

const originalWarn = console.warn.bind(console.warn)

// If adding a new warning, please include ample detail about why it's being silenced.
const silencedWarnings = [
    // See comment in nuxt-shims.ts 
    "Failed to resolve component: NuxtLink",
]

beforeAll(() => {
  console.warn = (msg: string) => {
      for (const silencedWarning of silencedWarnings) {
          if (msg.indexOf(silencedWarning) > 0) {
              return
          }
      }
      originalWarn("working" + msg)
  }
})

afterAll(() => {
    console.warn = originalWarn
})