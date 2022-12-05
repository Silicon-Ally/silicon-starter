<script setup lang="ts">
import { AuthProvider, AuthError, AuthErrorType } from '@/lib/auth'
import { computed } from 'vue'
enum AuthMode {
  SignIn = 'signin',
  SignUp = 'signup',
  ForgotPassword = 'forgotpw',
}
const router = useRouter()
const { fromQuery } = useURLParams()
// We use a query param instead of a hash because hashes don't get sent to the
// server, which means we can't use SSR if we use a hash.
const getAuthModeFromQuery = (): AuthMode | undefined => {
  const am = fromQuery('mode')
  if (!am) {
    return undefined
  }
  switch (am) {
  case AuthMode.SignIn:
    return AuthMode.SignIn
  case AuthMode.SignUp:
    return AuthMode.SignUp
  case AuthMode.ForgotPassword:
    return AuthMode.ForgotPassword
  default:
    return undefined
  }
}
const prefix = 'signIn'
const authMode = useState<AuthMode>(`${prefix}.authMode`, () => getAuthModeFromQuery() || AuthMode.SignIn)
const emailPassLoading = useState<boolean>(`${prefix}.emailPassLoading`, () => false)
const showPWResetDialog = useState<boolean>(`${prefix}.showPWResetDialog`, () => false)
const name = useState<string>(`${prefix}.name`, () => '')
const email = useState<string>(`${prefix}.email`, () => '')
const password = useState<string>(`${prefix}.password`, () => '')
const error = useState<string>(`${prefix}.error`, () => '')
const authLabel = computed(() => {
  switch (authMode.value) {
  case AuthMode.SignIn:
    return 'Sign In'
  case AuthMode.SignUp:
    return 'Sign Up'
  case AuthMode.ForgotPassword:
    return 'Can\'t sign in?'
  default:
    return ''
  }
})
const authButtonLabel = computed(() => {
  switch (authMode.value) {
  case AuthMode.SignIn:
    return 'Sign In'
  case AuthMode.SignUp:
    return 'Create Account'
  case AuthMode.ForgotPassword:
    return 'Reset Password'
  default:
    return ''
  }
})
const disableAuthButton = computed(() => {
  switch (authMode.value) {
  case AuthMode.SignIn:
    return !email.value || !password.value
  case AuthMode.SignUp:
    return !name.value || !email.value || !password.value || password.value.length < 6
  case AuthMode.ForgotPassword:
    return !email.value
  default:
    return false
  }
})
const isSignUp = computed(() => authMode.value === AuthMode.SignUp)
const isForgotPW = computed(() => authMode.value === AuthMode.ForgotPassword)
const defaultRedirectPath = '/'
const redirectPath = (): string => {
  const fq = fromQuery('redirect')
  return fq ? decodeURIComponent(fq) : defaultRedirectPath
}
const goToRedirectPath = (): Promise<void> => {
  const url = redirectPath()
  const parts = url.split('?')
  if (parts.length === 1) {
    return router.push(url).then(() => { /* cast to void */})
  }
  if (parts.length > 2) {
    throw new Error('Had more than two ? in a url: ' + url)
  }
  return router.push({
    path: parts[0],
    query: Object.fromEntries(new URLSearchParams(parts[1])),
  }).then(() => { /* cast to void */})
}
const {
  signIn,
  signInWithProvider,
  createAccount,
  signedIn,
  sendPasswordResetEmailFor,
} = useSession()
if (signedIn.value) {
  goToRedirectPath()
}
const toggleAuthMode = () => {
  switch (authMode.value) {
  case AuthMode.SignIn:
    authMode.value = AuthMode.SignUp
    break
  case AuthMode.SignUp:
    authMode.value = AuthMode.SignIn
    break
  default:
  }
  router.push({ query: { mode: authMode.value, redirect: encodeURIComponent(redirectPath()) } })
}
const handleAuthError = (errorTitle: string): ((e: Error | AuthError) => void) => {
  return (e: Error | AuthError) => {
    let errorDetails = 'Something went wrong'
    if (e instanceof AuthError) {
      switch (e.errorType()) {
      case AuthErrorType.UserNotFound:
        errorDetails = 'Your user account was not found.'
        break
      case AuthErrorType.IncorrectPassword:
        errorDetails = 'Your password is invalid.'
        break
      case AuthErrorType.InvalidPassword:
        errorDetails = 'Your password is invalid, passwords should be made of letters, numbers, and symbols.'
        break
      case AuthErrorType.WeakPassword:
        // See https://firebase.google.com/docs/auth/admin/errors,
        // specifically 'auth/weak-password'.
        errorDetails = 'Your password must be at least six characters.'
        break
      case AuthErrorType.MissingEmail:
        errorDetails = 'Please enter an email address.'
        break
      case AuthErrorType.AccountExistsWithDifferentCreds:
        errorDetails = 'Your account is registered with a different login method.'
        break
      }
    }
    console.log(e) 
    error.value = `${errorTitle}: ${errorDetails}`
  }
}
const forgotPassword = () => {
  authMode.value = AuthMode.ForgotPassword
}
const signInWithAuthProvider = (ap: AuthProvider) => {
  signInWithProvider(ap)
    .then(() => {
      return goToRedirectPath()
    })
    .catch(handleAuthError('Failed to sign in'))
}
const signInWithGoogle = () => signInWithAuthProvider(AuthProvider.Google)
const signInWithFacebook = () => signInWithAuthProvider(AuthProvider.Facebook)
const handleEmailPass = () => {
  let doAuth: Promise<void> | undefined
  let errorTitle = ''
  switch (authMode.value) {
  case AuthMode.SignIn:
    doAuth = signIn(email.value, password.value)
      .then(goToRedirectPath)
      .then(() => Promise.resolve())
    errorTitle = 'Failed to sign in'
    break
  case AuthMode.SignUp:
    doAuth = createAccount(name.value, email.value, password.value)
      .then(goToRedirectPath)
      .then(() => Promise.resolve())
    errorTitle = 'Failed to sign up'
    break
  case AuthMode.ForgotPassword:
    doAuth = sendPasswordResetEmailFor(email.value)
      .then(() => { showPWResetDialog.value = true })
    break
  }
  if (!doAuth) {
    return
  }
  emailPassLoading.value = true
  doAuth
    .catch(handleAuthError(errorTitle))
    .finally(() => {
      emailPassLoading.value = false
    })
}
</script>



<template>
  <div>
    <div>
      <h1 class="text-xl font-medium mt-4 mb-2 text-gray-600">
        {{ authLabel }}
      </h1>
      <div style="display: flex; flex-direction: column; align-items: flex-start; width: 100%; max-width: 300px; gap: 5px;">
        <input
          v-if="isSignUp"
          v-model="name"
          type="text"
          placeholder="Name"
          aria-label="name"
        >
        <div
          v-if="isForgotPW"
          class="text-gray-600 text-sm"
        >
          We'll send a recovery link to
        </div>
        <input
          v-model="email"
          type="text"
          placeholder="Email"
          aria-label="email"
        >
        <input
          v-if="!isForgotPW"
          v-model="password"
          type="password"
          placeholder="Password"
          aria-label="password"
        >
        <button
          v-if="!isSignUp && !isForgotPW"
          @click="forgotPassword"
        >
          Can't sign in?
        </button>
        <button
          id="auth-button"
          :disabled="disableAuthButton"
          @click="handleEmailPass"
        >
          {{ authButtonLabel }}
        </button>
        <div style="padding: 10px 0;">
          or
        </div>
        <button
          @click="signInWithGoogle"
        >
          Sign In with Google
        </button>
        <button
          @click="signInWithFacebook"
        >
          Sign In with Facebook 
        </button>
        <button
          id="create-account"
          @click="toggleAuthMode"
        >
          <span
            v-if="isSignUp"
          >
            Alerady have an account? <strong>Sign In</strong>
          </span>
          <span
            v-else
          >
            Don't have an account? <strong>Create Account</strong>
          </span>
        </button>
        <i style="font-size: 12px;">As you might be able to tell, we are non-proscriptive on what set of UI components/framework you should use. We have used PrimeVue quite happily in the past, and endorse it, but it isn't baked into this starter repor to allow for alternatives. 
        </i>
      </div>
    </div>
    <div
      v-if="error"
    >
      {{ error }} 
    </div>
    <div
      v-if="showPWResetDialog"
    >
      <h3 class="text-center mt-0 mb-3 font-medium">
        Password reset link sent to <em>{{ email }}</em>
      </h3>

      <div class="text-center text-sm text-gray-600 font-normal">
        If that email is associated with an existing account, a password
        reset link has been sent to it. Use that link to reset your password
        and sign in.
      </div>
      <button @click="showPWResetDialog = false">
        Okay
      </button>
    </div>
  </div>
</template>