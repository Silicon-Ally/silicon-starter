# Initial Setup

Even for local development, the project has some dependencies on cloud
infrastructure, mainly for secret management (KMS keys) and authentication
(Firebase Auth). We use [Terraform](https://www.terraform.io/), a tool for
automating infrastructure management, to handle setting up resources.

We break our Terraform configuration into two main pieces:

- **Per-env config** - This lives in `terraform/` and contains all the
resources needed for a given environment, like `local`, `dev`, `prod`,
etc. Each environment is representated as a
[Terraform Workspace](https://developer.hashicorp.com/terraform/language/state/workspaces).
- **Shared config** - This lives in `terraform/shared` and contains all the
resources that are shared between projects, like Docker images and developer
KMS keys. It also contains a `terraform@` service account, which is used to
create all the per-env projects.

Note: Both of these configs use the
[`local` backend](https://developer.hashicorp.com/terraform/language/settings/backends/local),
meaning that any resources Terraform creates are stored locally on your
machine, and not checked into source control. For more robust Terraform
management, use one of the other backends, like
[`gcs`](https://developer.hashicorp.com/terraform/language/settings/backends/gcs).

Throughout the repo, there are placeholders of the form `<some identifier>`
that refer to app-specific resources, like project IDs or domain names. These
should be replaced as you proceed through setup and come across them.

You can use the following hideous `grep` invocation to get a sense of where
they are throughout the repo:

```bash
grep \
  -l '<\([[:alpha:]]\|[[:blank:]]\)\+>' \
  -r . \
  --exclude-dir=bazel-* \
  --exclude-dir=node_modules \
  --exclude-dir=.git \
  --exclude-dir=.postgres-data \
  --exclude-dir=.output \
  --exclude-dir=.nuxt \
  --exclude='*.vue' \
  --exclude='*.ts' \
  --exclude='*.xml' \
  --exclude='*.svg'
```

## Install Terraform

Download and install Terraform for your platform using [the instructions on
Terraform's downloads section](https://www.terraform.io/downloads).

## Populate service variables

The first step is to populate the
[`terraform/terraform.tfvars`](terraform/terraform.tfvars)
file, which contains all the basic configuration for the service and it's
various GCP resources. Here's an example config:


```terraform
  app_name          = "my-app"
  project_id_prefix = "myorg-myapp"
  billing_account   = "DEADBE-EF1234-567890"
  region            = "us-central1"
  zone              = "us-central1-a"
  org_id            = "123456789012"
  folder_id         = "" # optional
```

- `app_name` - The `snake-case`d name of the application. Used in resource
names and descriptions.
- `project_id_prefix` - The `snake-case`d project prefix, which all projects
will use in their names. For a prefix like `myorg-myapp`, your shared project
would be `myorg-myapp-shared-<random suffix>` and your `dev` project would be
`myorg-myapp-dev-<random suffix>`

## Create the shared project + resources

The first step is to create the shared project. This is created using your
local GCP credentials, usually configured with `gcloud auth login`.

```bash
cd terraform/shared
terraform init
terraform plan --var-file=../terraform.tfvars
terraform apply --var-file=../terraform.tfvars
```

You can then get the newly created project ID with `terraform output`

## Create the local environment

Once the shared project is created, you can create the resources required to
run the project locally, which will live in your `-local` project.

```bash
cd terraform

# Download providers and create the local workspace
terraform init
terraform workspace new local

# Look at the resources to created, verify they match your expectations, and
# then create them.
terraform plan -var-file "tfvars/$(terraform workspace show).tfvars"
terraform apply -var-file "tfvars/$(terraform workspace show).tfvars"
```

With this done, you're all set to do local development. Run `terraform output`
to view the various Firebase credentials you'll need to configure auth in the
frontend.

If you don't want to deploy your app yet, you can stop here.

Note: The `local` env is special cased, it doesn't create a SQL database,
Cloud Run service, or any other serving infrastructure. Any other workspace/
environment name will create a full, functioning environment.

## Create the dev database

Creating the dev config needs to be done in multiple parts, because the Cloud
Run service can't be deployed successfully until the database exists. So first,
we create the dev database.

```bash
cd terraform
terraform workspace new dev
terraform plan -var-file "tfvars/$(terraform workspace show).tfvars" -target 'module.database[0]'
terraform apply -var-file "tfvars/$(terraform workspace show).tfvars" -target 'module.database[0]'
```

This will create the dev database, and assign a strong random password to the
root `postgres` user, which we'll use for applying migrations and creating
less-privileged service-specific users.

We manage the database users manually because using the
[`google_sql_user` resource](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/sql_user)
in Terraform creates users with
[near `SUPERUSER` permissions](https://cloud.google.com/sql/docs/postgres/users#other_postgresql_users),
and that doesn't align with the principal of least privilege.

You can get the password for the `postgres` user by running `terraform
output postgres_password`. To connect to the database, which has a private
IP, follow
[these instructions](https://cloud.google.com/sql/docs/postgres/connect-instance-private-ip).

Once connected, you can create a service-specific user account with the
following commands, run as the aforementioned `postgres` user:

```sql
\c <db name>
CREATE USER <username> WITH PASSWORD '<password>';

GRANT CONNECT ON DATABASE <db name> TO <username>;
GRANT USAGE ON SCHEMA public TO <username>;

GRANT SELECT, INSERT, UPDATE, DELETE
ON ALL TABLES IN SCHEMA public
TO <username>;

ALTER DEFAULT PRIVILEGES
IN SCHEMA public
GRANT SELECT, INSERT, UPDATE, DELETE
ON TABLES TO <username>;
```

This will create a new SQL user named `<username>` with password `<password>`,
and permission to read and write data to tables in your main service database,
`<db name>`.`<db name>` is `<app name>`, but with any hyphens replaced with
underscores

To apply initial migrations to the database, you can run:

`bazel run //cmd/ tools/migatesqldb -- apply --dsn=<dsn>`

where `<dsn>` is a tunnel to the dev database through a bastion machine
[as mentioned above](https://cloud.google.com/sql/docs/postgres/connect-instance-private-ip).

## Deploy the dev backend service

With the database configured, you can now populate your secrets files. First
update `.sops.yaml` with all of the relevant project IDs and app name, then
run:

```bash
cd cmd/server/configs/secrets
sops dev.enc.json
```

If everything is configured correctly and you have active GCP credentials with
KMS access to the `shared` project, `sops` should have opened an editor for
you. Populate this file with:

```json
{
	"postgres": {
		"host": "10.X.Y.Z",
		"port": 5432,
		"database": "<db name>",
		"user": "<username>",
		"password": "<password>"
	}
}
```

`host` is the private IP address of the Cloud SQL DB.

The last thing to do before you can push a functional image is replace the
`<shared project ID>` in `cmd/server/BUILD.bazel`, using output from `cd
terraform/shared && terraform output project_id`

`bazel run //cmd/server:server_push`

Note that your user account won't have owner permission on any env projects
by default, make sure to configure any needed permissions with the terraform
service account or in the UI.

You can now deploy the service with:

```bash
cd terraform
terraform plan -var-file "tfvars/$(terraform workspace show).tfvars"
terraform apply -var-file "tfvars/$(terraform workspace show).tfvars"
```

This will create your Cloud Run service, using the image you just pushed, which
contains the `sops` credentials.

# Deploy the dev frontend service

Now that the backend is deployed, you can update the
`frontend/envs/dev.env` config with the host Cloud Run is running on, e.g.
`https://<app name>-abcdefghij-uc.a.run.app`.

Then deploy the frontend with `npm run deploy:dev`
