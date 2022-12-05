import { useNuxtApp } from '#app'
import { computed, Ref } from 'vue'
import { User } from '@/graphql/generated'
import {
  signInWithPopup,
  signInWithEmailAndPassword,
  createUserWithEmailAndPassword,
  sendPasswordResetEmail,
  signOut,
  GoogleAuthProvider,
  FacebookAuthProvider,
  UserCredential,
  AuthProvider as FirebaseAuthProvider,
  AuthError as FirebaseAuthError,
} from 'firebase/auth'
import { AuthProvider, UserInfo, AuthError, AuthErrorType } from '@/lib/auth'
import { AxiosError } from 'axios'

type SessionError = AuthError | FirebaseAuthError | AxiosError

export const useSession = () => {
  const { $auth, $axios, $rawGraphQL } = useNuxtApp()
  const router = useRouter()

  const sessionCookieRaw = useCookie('__session')
  const csrfTokenCookie = useCookie('csrf-token')

  const prefix = 'useSession'
  // We use useState + computed instead of useCookie directly because we want
  // to share the reactive state between pages + components, like transitioning
  // from /sign- in to some redirect and having the nav bar show the correct
  // status, but updating a useCookie value **changes** the value of the
  // cookie.
  const csrfToken = useState(`${prefix}.csrfToken`, () => csrfTokenCookie.value)
  const sessionCookie = computed(() => sessionCookieRaw.value)

  const userInfo = useState<UserInfo | undefined>(`${prefix}.userInfo`, () => undefined)
  const signedIn = computed(() => !!userInfo.value)

  const exchangeUserCredsForAppLogin = (): ((userCreds: UserCredential) => Promise<void>) => {
    return (userCreds: UserCredential): Promise<void> => {
      const { user } = userCreds
      if (!user) {
        Promise.reject(new Error('no user in Firebase response'))
      }
      return user.getIdToken()
        .then((idToken) => {
          const req = { idToken, csrfToken: csrfToken.value }
          const config = {
            withCredentials: true,
            headers: { /* Add any additional headers here, e.g. ReCAPTCHA integration */ },
          }
          return $axios.post<UserInfo>('/sessionLogin', req, config)
        })
        .then((resp) => {
          if (!resp || !resp.data) {
            throw new Error('no user info in login response')
          }
          userInfo.value = resp.data
          return signOut($auth)
        })
    }
  }

  // See https://firebase.google.com/docs/auth/admin/errors
  const codeToAuthErrorType = (code: string): AuthErrorType => {
    switch (code) {
    case 'auth/user-not-found':
      return AuthErrorType.UserNotFound
    case 'auth/wrong-password':
      return AuthErrorType.IncorrectPassword
    case 'auth/email-already-exists':
      return AuthErrorType.EmailAlreadyExists
    case 'auth/invalid-password':
      return AuthErrorType.InvalidPassword
    case 'auth/weak-password':
      return AuthErrorType.WeakPassword
    case 'auth/missing-email':
      return AuthErrorType.MissingEmail
    case 'auth/account-exists-with-different-credential':
      return AuthErrorType.AccountExistsWithDifferentCreds
    default:
      return AuthErrorType.Generic
    }
  }

  const handleAuthError = (e: SessionError) => {
    let errCode = ''
    if ('code' in e) {
      errCode = e.code || ''
    }
    console.log(e)
    throw new AuthError({
      type: codeToAuthErrorType(errCode),
      message: e.message || '',
      cause: e,
    })
  }


  const currentUser = useState<User | undefined>(`${prefix}.currentUser`, () => undefined)
  const loadCurrentUser = async () => {
    if (currentUser.value === undefined) {
      await $rawGraphQL.me({})
        .then((resp) => {
          currentUser.value = resp.me
        })
    }
    return currentUser
  }
  const getMe = async () => {
    const lcu = await loadCurrentUser()
    // LoadCurrentUser's return is only undefined as a technicality to support
    // the single-lookup behavior above. This cast is safe.
    return lcu as Ref<User>
  }
  const getMaybeMe = async () => { 
    if (!signedIn.value) {
      // Will be a Ref with a value of undefined.
      return currentUser
    }
    const lcu = await loadCurrentUser()
    return lcu
  }
  const refreshMe = (): Promise<void> => {
    // Ideally, we'd use `!signedIn.value` here instead, but the computed
    // value inexplicably isn't updated here yet when we call refreshMe from
    // logOut. So we go directly to the source.
    if (!userInfo.value) {
      currentUser.value = undefined
      return Promise.resolve() 
    }
    return $rawGraphQL.me({})
      .then((resp) => currentUser.value = resp.me)
      .then(() => { /* noop */ })
      .catch((_err) => { currentUser.value = undefined })
  }
  const logOut = () => {
    return $axios.post('/sessionLogout', {}, { withCredentials: true })
      .then(() => router.push('/sign-in'))
      .then(() => {
        // Clear the state, the cookie itself is cleared by the server.
        userInfo.value = undefined
        return refreshMe()
      })
  }
  const sendPasswordResetEmailFor = (email: string): Promise<unknown> => sendPasswordResetEmail($auth, email)
  
  return {
    signIn: (email: string, password: string) => {
      return signInWithEmailAndPassword($auth, email, password)
        .then(exchangeUserCredsForAppLogin())
        .then(refreshMe)
        .catch(handleAuthError)
    },
    createAccount: (name: string, email: string, password: string) => {
      return createUserWithEmailAndPassword($auth, email, password)
        .then(exchangeUserCredsForAppLogin())
        .then(refreshMe)
        .catch(handleAuthError)
    },
    signInWithProvider: (provider: AuthProvider) => {
      let fbProvider: FirebaseAuthProvider | undefined
      switch (provider) {
      case AuthProvider.Google:
        fbProvider = new GoogleAuthProvider()
        break
      case AuthProvider.Facebook:
        fbProvider = new FacebookAuthProvider()
        break
      default:
        throw `unknown provider ${provider})`
      }

      return signInWithPopup($auth, fbProvider)
        .then(exchangeUserCredsForAppLogin())
        .then(refreshMe)
        .catch(handleAuthError)
    },
    signedIn,
    userInfo,
    sessionCookie,
    getMe,
    getMaybeMe,
    refreshMe,
    sendPasswordResetEmailFor,
    logOut,
  }
}
