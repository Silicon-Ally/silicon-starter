export interface UserInfo {
	email: string;
}

export enum AuthProvider {
  Google,
  EmailAndPass,
  Facebook,
}

export enum AuthErrorType {
  // Any error not described by another AuthErrorType value.
  Generic,
  // No user with the given credentials was found.
  UserNotFound,
  // The password given was incorrect.
  IncorrectPassword,
  // The email is already attached to an account.
  EmailAlreadyExists,
  // The password uses some invalid characters.
  InvalidPassword,
  // The password doesn't meet the minimum password requirements.
  WeakPassword,
  // No email was provided.
  MissingEmail,
  // This account already exists, but is tied to a different auth mechanism.
  AccountExistsWithDifferentCreds,
}

export interface AuthErrorOpts {
  message?: string;
  type?: AuthErrorType;
  cause?: Error;
}

// See https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Error#custom_error_types
export class AuthError extends Error {
  private errType!: AuthErrorType
  cause?: Error

  constructor(opts: AuthErrorOpts) {
    super()

    if (Error.captureStackTrace) {
      Error.captureStackTrace(this, AuthError)
    }

    this.name = 'AuthError'
    this.message = opts.message || ''
    this.errType = opts.type || AuthErrorType.Generic
    this.cause = opts.cause
  }

  public errorType(): AuthErrorType {
    return this.errType
  }
}
