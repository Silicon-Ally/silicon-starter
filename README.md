# Silicon Starter

Starter repo for a modern, performant, cheap, and developer-friendly stack.

This repo implements a to-do list app as a demo use case.

## Stack

- [Bazel](https://bazel.build/) (Build System)
- [Gazelle](https://github.com/bazelbuild/bazel-gazelle) (For managing Go dependencies)
- [Postgres](https://www.postgresql.org/) (DB)
- [Golang](https://go.dev/) (Backend Language)
- [GCP Cloud Run](https://cloud.google.com/run/) (Backend Serving Environment)
- [Firebase Authentication](https://firebase.google.com/docs/auth)
- [GraphQL](https://graphql.org/) (Transit / API Layer)
- [Nuxt 3](https://nuxt.com/)/[TypeScript](https://www.typescriptlang.org/) (Frontend Framework)
- [Firebase Cloud Functions](https://firebase.google.com/docs/functions/) (SSR Environemnt)
- [Firebase Hosting](https://firebase.google.com/docs/hosting) (Static Asset Serving)
- [Mozilla sops](https://github.com/mozilla/sops) (Secret Management)

If you plan to use this template, we HIGHLY recommend you first read:

- Our [SA Standard Tech Stack](https://siliconally.getoutline.com/s/7bb98b43-fa2c-43f5-ae3b-db4926a5036a), 
an overview of the function and motivation for each system in this list.
- The [SA Developer Handbook](https://siliconally.getoutline.com/s/d984f195-3e5e-410f-bce8-63676496661f), 
a set of instructions on how to set up the development environment for linux.

## Initial Setup

In order to get the app up and running, there's some initial setup you'll need
to do. Follow [the Initial Setup guide](INITIAL_SETUP.md) guide to get started.

## Run the App

Run each of these commands in separate shells (perhaps using a tool like [tmux](https://github.com/tmux/tmux)):

```
bazel run //scripts:run_db
bazel run //scripts:run_backend
cd frontend; npm run frontend
```

then, visit `localhost:3000` to see the results.

## Why make this?

Building a web-app from scratch is intimidating. There are a plethora of
technologies to choose from for every layer of a web-stack, and each comes with
its fair share of sharp edge cases. While it might be easy to get any given
technolgy spun up according to its starter instructions, figuring out how to
get technolgies working together, and deciding what set of technolgies to get
set up prior to development can be daunting.

That's why we made this repo. It has everything you need to make a web app that is:

- Cheap to run (minimizes serving overhead costs)
- Scales to arbitrary capacity (can serve any number of users, allowing the cloud platform to add more serving capacity when there are more users)
- Relational (uses a relational database for efficent lookups + joins)
- Fast (uses Golang, a performant language for traditional server workloads)
- Type-checked (uses languages and tools that run type validation to quickly catch a large percentage of bugs)
- Server-side Rendered (allows for fast initial page loads, and is great for network + compute constrained devices)

This isn't the right stack for every web app, but it's an excellent place for any web-app to get started,
and this configuration has worked exceptionally well for our projects.

## Repo Structure

The main code locations:

- `/db`: All database configuration and logic. See [`db/sqldb/README`](./db/sqldb/README.md) for details.
- `/cmd/server`: The code and configuration for the backend server. See [`cmd/server/README`](./cmd/server/README.md) for details.
- `/frontend`: The code and configuration for the frontend. See that [`frontend/README`](./frontend/README.md) for details.

Other important code locations:

- `/todo`: The application specific data structures used across packages, you'd want to replace this with a domain specific name.
- `/cmd/tools`: A place to put Go binaries used for tooling, testing, and other tasks.
- `/terraform`: The Terraform configuration for your service, which can be used to deploy it to one or more environments.
- `/authn`: Code for handling authentication using Firebase.

## Deployment

Follow [the initial setup guide](/INITIAL_SETUP.md) for instructions on how
to configure Terraform, and deploy the frontend and backend components of the
application.

## Status

This project is a work in progress. We're using it and updating it as we do,
but there are certain to be some rough edges for now. If you find anything that
is broken, unexplained, or unclear, please file a bug in this repo, and we'll
try to take a look shortly.

Please report security issues to security@siliconally.org, or by using one of
the contact methods available on our
[Contact Us page](https://siliconally.org/contact/).

## Contributing

Contribution guidelines can be found [on our website](https://siliconally.org/oss/contributor-guidelines).

