import { useNuxtApp } from '#app'

interface HasPath {
  path: string
}

export default defineNuxtRouteMiddleware(async (to: HasPath) => {
  const { $sauth } = useNuxtApp()
  const { signedIn, userInfo, sessionCookie } = useSession()

  const maybeRedirectToSignIn = () => {
    if (to.path === '/sign-in') {
      // They're already on the sign in page, just return.
      return
    }

    return navigateTo({
      path: '/sign-in',
      query: {
        redirect: encodeURIComponent(to.path),
      },
    }, { redirectCode: 302 })
  }

  // On the server, we verify their JWT to determine access.
  if (process.server) {
    if (!sessionCookie.value) {
      return maybeRedirectToSignIn()
    }

    const decodedJWT = await $sauth.verifySessionCookie(sessionCookie.value, false /* checkRevoked */)
    if (!decodedJWT) {
      return maybeRedirectToSignIn()
    }
    
    if (decodedJWT.email) {
      userInfo.value = { email: decodedJWT.email }
    }
    return
  }

  // If we aren't on the server, we just trust what the server has previously set.
  if (!signedIn.value) {
    return maybeRedirectToSignIn()
  }
  return
})
