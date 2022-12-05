import { defineNuxtPlugin } from '#app'
import axios from 'axios'
import { headersToForward } from '@/lib/headers'

export default defineNuxtPlugin(() => {
  const { app: { baseServerURL, clientsUseFullURL } } = useRuntimeConfig()
  let baseURL = baseServerURL.path
  if (process.server || clientsUseFullURL) {
    // To use a full server name.
    baseURL = baseServerURL.endpoint + baseServerURL.path
  }
  const client = axios.create({ baseURL })
  client.interceptors.request.use((config) => {
    if (!config.url) {
      return config
    }
    if (!config.headers) {
      config.headers = {}
    }

    const headers = useRequestHeaders(headersToForward)
    if (process.server && config.url.startsWith(baseURL) && headers) {
      for (const hdr of headersToForward) {
        const val = headers[hdr]
        if (val) {
          config.headers[hdr] = val
        }
      }
    }
    return config
  })

  return {
    provide: {
      axios: client,
    },
  }
})