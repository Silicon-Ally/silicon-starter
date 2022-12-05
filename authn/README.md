# Authentication

This starter app has Authentication baked into it, so that you can trivially support user accounts
and profiles. These logins are stateful and reversible, and this initial demo demonstrates how to 
safely plumb authentication data throughout an App. 

Note: This app is built to require authentication on every page except for the Log-In page (which
logged out users will have to be able to see). You can modify this by changing the `WithAuthorization`
method.

Authentication is split into session management (setting and retrieving cookies, user creation, etc),
and Firebase specific logic implementations of generic authentication functions.

The three primary methods in this package are all in `session/session.go`:

- `LoginHandler` is an HTTP handler for handling a login attempt. It handles cases of existing
and new users, and finds or creates information about the user in the database.
- `LogoutHandler` is the analogous HTTP handler for handling a logout attempt - it clears the
user's cookies, and makes it so that the user cannot refresh their session going forward.
- `WithAuthorization` is an HTTP handler which redirects the user to login if they aren't yet
authorized, and populates authorization data (most critically, the UserID of the authorized
user) on the context for all business logic (and in our case, graphql resolvers) to use.