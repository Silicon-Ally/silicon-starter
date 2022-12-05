import { defineNuxtPlugin } from '#app'
import { GraphQLClient } from 'graphql-request'
import { getSdk, SdkFunctionWrapper } from '@/graphql/generated'
import { headersToForward } from '@/lib/headers'

// Copied from graphql-request/src/types.dom
interface Headers {
  append(name: string, value: string): void
  delete(name: string): void
  get(name: string): string | null
  has(name: string): boolean
  set(name: string, value: string): void
  // We've copied this from a separate module, disable the no-explicit-any check.
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  forEach(callbackfn: (value: string, key: string, parent: Headers) => void, thisArg?: any): void
}

type HeadersInit = Headers | string[][] | Record<string, string>

type RequestCredentials = 'omit' | 'same-origin' | 'include'

interface RequestInit {
  credentials?: RequestCredentials
  headers?: HeadersInit
}

export default defineNuxtPlugin(() => {
  const { app: { graphQLServerURL, clientsUseFullURL } } = useRuntimeConfig()
  const opts: RequestInit = {
    // Needed to forward session cookie from client to server.
    credentials: 'include',
  }

  let baseURL = graphQLServerURL.path
  if (process.server || clientsUseFullURL) {
    // To use a full server name.
    baseURL = graphQLServerURL.endpoint + graphQLServerURL.path
  }


  if (process.server) {
    // To forward the client's cookies + any recaptcha header onward to our API.
    const reqHeaders = useRequestHeaders(headersToForward)
    const headers: Record<string, string> = {}
    for (const key of Object.keys(reqHeaders)) {
      const val = reqHeaders[key as Lowercase<string>]
      if (typeof(val) === 'string') {
        headers[key] = val
      }
    }
    opts.headers = headers
  }

  return {
    provide: {
      rawGraphQL: getSdk(new GraphQLClient(baseURL, opts)),
      graphQLWithWrapper: (wrapper: SdkFunctionWrapper): ReturnType<typeof getSdk> => {
        return getSdk(new GraphQLClient(baseURL, opts), wrapper)
      },
    },
  }
})
