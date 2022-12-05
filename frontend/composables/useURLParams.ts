import type { RouteParams, LocationQuery } from 'vue-router'
import { useRoute } from 'vue-router'

export const useURLParams = () => {
  const route = useRoute()

  const getVal = (src: RouteParams | LocationQuery, key: string): string | undefined => {
    const val = src[key]
    if (!val) {
      return undefined
    }

    if (Array.isArray(val)) {
      if (val.length === 0) {
        return undefined
      }
      if (!val[0]) {
        return undefined
      }
      return val[0]
    }

    return val
  }

  return {
    fromQuery: (key: string): string | undefined => {
      return getVal(route.query, key)
    },
    fromParams: (key: string): string | undefined => {
      return getVal(route.params, key)
    },
  }
}