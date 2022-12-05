# Flags

This folder contains the flags used (on a per-environment basis)
when running the backend server. To see the full set of optional flags,
or to add a new flag, check out the FlagSet declaration in `main.go`

Note for any secrets (API keys, Passwords, or other sensitive 
configuration) associated with your project, you should use SOPS,
a mechanism for passing encrypted secrets to your running server. It's
highly discouraged for security reasons to pass secrets to your backend
as flags.
