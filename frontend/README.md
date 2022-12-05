# Frontend

The frontend is built using
[NuxtJS](https://v3.nuxtjs.org), which is a framework built around
[Vue.js](https://v3.vuejs.org/) that provides server-side rendering,
middleware, and other useful features for production applications.

Specifically, we use V3 of Vue + Nuxt, which introduces the [Composition
API](https://v3.vuejs.org/api/composition-api.html), which is used throughout
the project 

## Getting Started

Follow our [Developer
Handbook](https://siliconally.getoutline.com/share/d984f195-3e5e-410f-bce8-63676496661f)
to set up your development environment.

Specifically, you'll need
[npm](https://siliconally.getoutline.com/share/d984f195-3e5e-410f-bce8-63676496661f#h-nvm-npm)
to run the frontend, and then run `npm install`.

## Development

To start the development server on http://localhost:3000, `cd` to the `frontend` folder, then
run

```bash
npm run frontend 
```

This will start up a server featuring hot reload.

For typechecking, run:

```bash
npm run typecheck
```

For linting and fixing any fixable lint issues, run

```bash
npm run lint:fix
```

For running any tests that you write with Jest, run 

```bash
npm run test 
```

## Compilation 

To build the application for a given environment, run the corresponding `npm run` command:

```bash
npm run build:local
npm run build:dev
```

Checkout the [deployment documentation](https://v3.nuxtjs.org/docs/deployment)
for more info.

## GraphQL

We use [GraphQL] to describe the API between this web frontend and our
[Adventure Scientists server](/cmd/server). To take advantage of strong schema
typing and TypeScript, we use a [GraphQL Code
Generator](https://github.com/dotansimha/graphql-code-generator), which takes
in our [API Schema](/cmd/server/graph/schema.graphqls) and [a set of desired
operations](/frontend/operations/queries) to produce a [generated TypeScript
file](/frontend/graphql/generated/index.ts) that contains pre-formed queries
for us to execute with
[graphql-request](https://github.com/prisma-labs/graphql-request), a popular
and simple GraphQL client.

To update the bindings after changing the GraphQL schema or operations, run:

```bash
npm run graphql-gen
```

## Deployment

Frontend deployment is handled via Firebase. Specifically, we serve the SSR
component using [Cloud Functions for
Firebase](https://firebase.google.com/docs/functions), and the frontend using
[Firebase Hosting](https://firebase.google.com/docs/hosting).

To manually run a deployment, first authenticate with Firebase locally using
one of the two options below:

```bash
# If you've never logged into Firebase before.
npx firebase login

# If you're already authenticated with a different account.
npx firebase login:add
npx firebase login:use <email>
```

Then, to build the server and deployment it, run:

```bash
npm run deploy:dev
```

Make sure you've updated the `scripts['deploy:dev']` in `package.json` with your dev project ID.
