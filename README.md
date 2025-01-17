# tfmigrate
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![GitHub release](https://img.shields.io/github/release/minamijoyo/tfmigrate.svg)](https://github.com/minamijoyo/tfmigrate/releases/latest)
[![GoDoc](https://godoc.org/github.com/minamijoyo/tfmigrate/tfmigrate?status.svg)](https://godoc.org/github.com/minamijoyo/tfmigrate)

A Terraform / OpenTofu state migration tool for GitOps.

## Table of content
<!--ts-->
   * [Features](#features)
   * [Why?](#why)
   * [Requirements](#requirements)
   * [Getting Started](#getting-started)
   * [Install](#install)
      * [Homebrew](#homebrew)
      * [Download](#download)
      * [Source](#source)
   * [Usage](#usage)
   * [Configurations](#configurations)
      * [Environment variables](#environment-variables)
      * [Configuration file](#configuration-file)
         * [tfmigrate block](#tfmigrate-block)
         * [history block](#history-block)
         * [storage block](#storage-block)
         * [storage block (local)](#storage-block-local)
         * [storage block (s3)](#storage-block-s3)
         * [storage block (gcs)](#storage-block-gcs)
   * [Migration file](#migration-file)
      * [Environment Variables](#environment-variables-1)
      * [migration block](#migration-block)
      * [migration block (state)](#migration-block-state)
         * [state mv](#state-mv)
         * [state xmv](#state-xmv)
         * [state rm](#state-rm)
         * [state import](#state-import)
         * [state replace-provider](#state-replace-provider)
      * [migration block (multi_state)](#migration-block-multi_state)
         * [multi_state mv](#multi_state-mv)
         * [multi_state xmv](#multi_state-xmv)
   * [Integrations](#integrations)
   * [License](#license)
<!--te-->

## Features

- GitOps friendly: Write terraform state mv/rm/import commands in HCL, plan and apply it.
- Monorepo style support: Move resources to other tfstates to split and merge easily for refactoring.
- Dry run migration: Simulate state operations with a temporary local tfstate and check to see if terraform plan has no changes after the migration without updating remote tfstate.
- Migration history: Keep track of which migrations have been applied and apply all unapplied migrations in sequence.

You can apply terraform state operations in a declarative way.

In short, write the following migration file and save it as `state_mv.hcl`:

```hcl
migration "state" "test" {
  dir = "dir1"
  actions = [
    "mv aws_security_group.foo aws_security_group.foo2",
    "mv aws_security_group.bar aws_security_group.bar2",
  ]
}
```

Then, apply it:

```
$ tfmigrate apply state_mv.hcl
```

It works as you expect, but it's just a text file, so you can commit it to git.

## Why?

If you have been using Terraform in production for a long time, tfstate manipulations are unavoidable for various reasons. As you know, the terraform state command is your friend, but it's error-prone and not suitable for a GitOps workflow.

In team development, Terraform configurations are generally managed by git and states are shared via remote state storage which is outside of version control. It's a best practice for Terraform.
However, most Terraform refactorings require not only configuration changes, but also state operations such as state mv/rm/import. It's not desirable to change the remote state before merging configuration changes. Your colleagues may be working on something else and your CI/CD pipeline continuously plan and apply their changes automatically. At the same time, you probably want to check to see if terraform plan has no changes after the migration before merging configuration changes.

To fit into the GitOps workflow, the answer is obvious. We should commit all terraform state operations to git.
This brings us to a new paradigm, that is to say, Terraform state operation as Code!

## Requirements

The tfmigrate invokes `terraform` or `tofu` command under the hood. This is because we want to support multiple Terraform / OpenTofu versions in a stable way.

### Terraform

The minimum required version is Terraform v0.12 or higher, but we recommend the Terraform v1.x.

### OpenTofu

If you want to use OpenTofu, a community fork of Terraform, you need to set the environment variable `TFMIGRATE_EXEC_PATH` to `tofu`.

The minimum required version is OpenTofu v1.6 or higher.

### Terragrunt

#### Without dynamic state

If you are not leveraging `terragrunt`s [dynamic state generation](https://terragrunt.gruntwork.io/docs/reference/config-blocks-and-attributes/#remote_state) the environment variable `TF_MIGRATE_EXEC_PATH` must be set to `terragrunt`.

```shell
# As part of the command or via exporting the variable to your shell. 
TFMIGRATE_EXEC_PATH=terragrunt tfmigrate $OTHEROPTIONS
```

#### With dynamic state

If you are leveraging `terragrunt`s [dynamic state generation](https://terragrunt.gruntwork.io/docs/reference/config-blocks-and-attributes/#remote_state), the `remote_state` block must include a `generate` block.

This ensures that that `terragrunt` doesn't utilize command line flags for remote state configuration that are incompatible with the local backend, which is utilized by `tfmigrate` for planning.

```
remote_state {
  backend = "s3"

  config = {
    bucket         = "highway-terraform-state"
    # Other config here
  }
  # This ensures that a file instead of command line flags are used.
  # allowing tfmigrate to work as expected.  
  generate = {
    path      = "backend.tf"
    if_exists = "overwrite_terragrunt"
  }
}
```

## Getting Started

As you know, terraform state operations are dangerous if you don't understand what you are actually doing. If I were you, I wouldn't use a new tool in production from the start. So, we recommend you to play an example sandbox environment first, which is safe to run terraform state command without any credentials. The sandbox environment mocks the AWS API with `localstack` and doesn't actually create any resources. So you can safely run the `tfmigrate` and `terraform` commands, and easily understand how the tfmigrate works.

Build a sandbox environment with docker compose and run bash:

```
$ git clone https://github.com/minamijoyo/tfmigrate
$ cd tfmigrate/
$ docker compose build
$ docker compose run --rm tfmigrate /bin/bash
```

In the sandbox environment, create and initialize a working directory from test fixtures:

```
# mkdir -p tmp && cp -pr test-fixtures/backend_s3 tmp/dir1 && cd tmp/dir1
# terraform init
# cat main.tf
```

This example contains two `aws_security_group` resources:

```hcl
resource "aws_security_group" "foo" {
  name = "foo"
}

resource "aws_security_group" "bar" {
  name = "bar"
}
```

Apply it and confirm that the state of resources are stored in the tfstate:

```
# terraform apply -auto-approve
# terraform state list
aws_security_group.bar
aws_security_group.foo
```

Now, let's rename `aws_security_group.foo` to `aws_security_group.baz`:

```
# cat << EOF > main.tf
resource "aws_security_group" "baz" {
  name = "foo"
}

resource "aws_security_group" "bar" {
  name = "bar"
}
EOF
```

At this point, of course, there are differences in the plan:

```
# terraform plan
(snip.)
Plan: 1 to add, 0 to change, 1 to destroy.
```

Now it's time for tfmigrate. Create a migration file:

```
# cat << EOF > tfmigrate_test.hcl
migration "state" "test" {
  actions = [
    "mv aws_security_group.foo aws_security_group.baz",
  ]
}
EOF
```

Run `tfmigrate plan` to check to see if `terraform plan` has no changes after the migration without updating remote tfstate:

```
# tfmigrate plan tfmigrate_test.hcl
(snip.)
YYYY/MM/DD hh:mm:ss [INFO] [migrator] state migrator plan success!
# echo $?
0
```

The plan command computes a new state by applying state migration operations to a temporary state. It will fail if terraform plan detects any diffs with the new state. If you are wondering how the `tfmigrate` command actually works, you can see all `terraform` commands executed by the tfmigrate with log level `DEBUG`:

```
# TFMIGRATE_LOG=DEBUG tfmigrate plan tfmigrate_test.hcl
```

If looks good, apply it:

```
# tfmigrate apply tfmigrate_test.hcl
(snip.)
YYYY/MM/DD hh:mm:ss [INFO] [migrator] state migrator apply success!
# echo $?
0
```

The apply command computes a new state and pushes it to remote state.
It will fail if terraform plan detects any diffs with the new state.

You can confirm the latest remote state has no changes with terraform plan:

```
# terraform plan
(snip.)
No changes. Infrastructure is up-to-date.

# terraform state list
aws_security_group.bar
aws_security_group.baz
```

There is no magic. The tfmigrate just did the boring work for you.

Furthermore, you can also move resources to another directory. Let's split the tfstate in two.
Create a new empty directory with a different remote state path:

```
# mkdir dir2
# cat config.tf | sed 's/test\/terraform.tfstate/dir2\/terraform.tfstate/' > dir2/config.tf
```

Move the resource definition of `aws_security_group.baz` in `main.tf` to `dir2/main.tf` and rename it to `aws_security_group.baz2`:

```
# cat << EOF > main.tf
resource "aws_security_group" "bar" {
  name = "bar"
}
EOF

# cat << EOF > dir2/main.tf
resource "aws_security_group" "baz2" {
  name = "foo"
}
EOF
```

Create a `multi_state` migration file:

```
# cat << EOF > tfmigrate_multi_state_test.hcl
migration "multi_state" "test" {
  from_dir = "."
  to_dir   = "dir2"

  actions = [
    "mv aws_security_group.baz aws_security_group.baz2",
  ]
}
EOF
```

Run tfmigrate plan & apply:

```
# tfmigrate plan tfmigrate_multi_state_test.hcl
# tfmigrate apply tfmigrate_multi_state_test.hcl
```

You can see the tfstate was split in two:

```
# terraform state list
aws_security_group.bar
# cd dir2 && terraform state list
aws_security_group.baz2
```

## Install

### Homebrew

If you are macOS user:

```
$ brew install tfmigrate
```

### Download

Download the latest compiled binaries and put it anywhere in your executable path.

https://github.com/minamijoyo/tfmigrate/releases

### Source

If you have Go 1.22+ development environment:

```
$ git clone https://github.com/minamijoyo/tfmigrate
$ cd tfmigrate/
$ make install
$ tfmigrate --version
```

## Usage

```
$ tfmigrate --help
Usage: tfmigrate [--version] [--help] <command> [<args>]

Available commands are:
    apply    Compute a new state and push it to remote state
    list     List migrations
    plan     Compute a new state
```

```
$ tfmigrate plan --help
Usage: tfmigrate plan [PATH]

Plan computes a new state by applying state migration operations to a temporary state.
It will fail if terraform plan detects any diffs with the new state.

Arguments:
  PATH                     A path of migration file
                           Required in non-history mode. Optional in history-mode.

Options:
  --config                 A path to tfmigrate config file
  --backend-config=path    A backend configuration, a path to backend configuration file or
                           key=value format backend configuraion.
                           This option is passed to terraform init when switching backend to remote.

  --out=path               Save a plan file after dry-run migration to the given path.
                           Note that the saved plan file is not applicable in Terraform 1.1+.
                           It's intended to use only for static analysis.
```

```
$ tfmigrate apply --help
Usage: tfmigrate apply [PATH]

Apply computes a new state and pushes it to remote state.
It will fail if terraform plan detects any diffs with the new state.

Arguments
  PATH                     A path of migration file
                           Required in non-history mode. Optional in history-mode.

Options:
  --config                 A path to tfmigrate config file
  --backend-config=path    A backend configuration, a path to backend configuration file or
                           key=value format backend configuraion.
                           This option is passed to terraform init when switching backend to remote.
```

```
$ tfmigrate list --help
Usage: tfmigrate list

List migrations.

Options:
  --config           A path to tfmigrate config file
  --status           A filter for migration status
                     Valid values are as follows:
                       - all (default)
                       - unapplied
```

## Configurations
### Environment variables

You can customize the behavior by setting environment variables.

- `TFMIGRATE_LOG`: A log level. Valid values are `TRACE`, `DEBUG`, `INFO`, `WARN`, `ERROR`. Default to `INFO`.
- `TFMIGRATE_EXEC_PATH`: A string how terraform command is executed. Default to `terraform`. It's intended to inject a wrapper command such as direnv. e.g.) `direnv exec . terraform`. To use OpenTofu, set this to `tofu`.

Some history storage implementations may read additional cloud provider-specific environment variables. For details, refer to a configuration file section for storage block described below.

### Configuration file

You can customize the behavior by setting a configuration file.
The path of configuration file defaults to `.tfmigrate.hcl`. You can change it with command line flag `--config`.

The syntax of configuration file is as follows:

- A configuration file must be written in the HCL2.
- The extension of file must be `.hcl`(for HCL native syntax) or `.json`(for HCL JSON syntax).
- The file must contain exactly one `tfmigrate` block.

An example of configuration file is as follows.

```hcl
tfmigrate {
  migration_dir = "./tfmigrate"
  is_backend_terraform_cloud = true
  history {
    storage "s3" {
      bucket = "tfmigrate-test"
      key    = "tfmigrate/history.json"
    }
  }
}
```

#### is_backend_terraform_cloud
Whether the remote backend specified in Terraform files references a
[terraform cloud remote backend](https://www.terraform.io/language/settings/terraform-cloud),
in particular specified as a `cloud` block within the `terraform` config block. This backend
type was introduced in Terraform 1.1.+ and is the recommended way to specify a Terraform backend.
Attribute defaults to `false`.

Note that when using tfmigrate with Terraform Cloud, you also need to set a workspace name in a migration file.

#### tfmigrate block

The `tfmigrate` block has the following attributes:

- `migration_dir` (optional): A path to directory where migration files are stored. Default to `.` (current directory).

The `tfmigrate` block has the following blocks:

- `history` (optional): Keep track of which migrations have been applied.

#### history block

The `history` block has the following blocks:

- `storage` (required): A migration history data store

#### storage block

The storage block has one label, which is a type of storage. Valid types are as follows:

- `local`: Save a history file to local filesystem.
- `s3`: Save a history file to AWS S3.
- `gcs`: Save a history file to GCS (Google Cloud Storage).

If your cloud provider has not been supported yet, as a workaround, you can use `local` storage and synchronize a history file to your cloud storage with a wrapper script.

#### storage block (local)

The `local` storage has the following attributes:

- `path` (required): A path to a migration history file.

An example of configuration file is as follows.

```hcl
tfmigrate {
  migration_dir = "./tfmigrate"
  history {
    storage "local" {
      path = "tmp/history.json"
    }
  }
}
```

#### storage block (s3)

The `s3` storage has the following attributes:

- `bucket` (required): Name of the bucket.
- `key` (required): Path to the migration history file.
- `region` (optional): AWS region. This can also be sourced from the `AWS_DEFAULT_REGION` and `AWS_REGION` environment variables.
- `access_key` (optional): AWS access key. This can also be sourced from the `AWS_ACCESS_KEY_ID` environment variable, AWS shared credentials file, or AWS shared configuration file.
- `secret_key` (optional): AWS secret key. This can also be sourced from the `AWS_SECRET_ACCESS_KEY` environment variable, AWS shared credentials file, or AWS shared configuration file.
- `profile` (optional): Name of AWS profile in AWS shared credentials file or AWS shared configuration file to use for credentials and/or configuration. This can also be sourced from the `AWS_PROFILE` environment variable.
- `role_arn` (optional): Amazon Resource Name (ARN) of the IAM Role to assume.
- `kms_key_id` (optional): Amazon Server-Side Encryption (SSE) KMS Key Id. When specified, this encryption key will be used and server-side encryption will be enabled. See the [terraform s3 backend](https://www.terraform.io/language/settings/backends/s3#kms_key_id).

The following attributes are also available, but they are intended to use with `localstack` for testing.

- `endpoint` (optional): Custom endpoint for the AWS S3 API.
- `skip_credentials_validation` (optional): Skip credentials validation via the STS API.
- `skip_metadata_api_check` (optional): Skip usage of EC2 Metadata API.
- `force_path_style` (optional): Enable path-style S3 URLs (`https://<HOST>/<BUCKET>` instead of `https://<BUCKET>.<HOST>`).

An example of configuration file is as follows.

```hcl
tfmigrate {
  migration_dir = "./tfmigrate"
  history {
    storage "s3" {
      bucket  = "tfmigrate-test"
      key     = "tfmigrate/history.json"
      region  = "ap-northeast-1"
      profile = "dev"
    }
  }
}
```

#### storage block (gcs)

The `gcs` storage has the following attributes:

- `bucket` (required): Name of the bucket.
- `name` (required): Path to the migration history file.

Note that this storage implementation refers the Application Default Credentials (ADC) for authentication.

An example of configuration file is as follows.

```hcl
tfmigrate {
  migration_dir = "./tfmigrate"
  history {
    storage "gcs" {
      bucket = "tfstate-test"
      name   = "tfmigrate/history.json"
    }
  }
}
```

If you want to connect to an emulator instead of GCS, set the `STORAGE_EMULATOR_HOST` environment variable as required by the [Go library for GCS](https://pkg.go.dev/cloud.google.com/go/storage).

## Migration file

You can write terraform state operations in HCL. The syntax of migration file is as follows:

- A migration file must be written in the HCL2.
- The extension of file must be `.hcl`(for HCL native syntax) or `.json`(for HCL JSON syntax).

Although the filename can be arbitrary string, note that in history mode unapplied migrations will be applied in alphabetical order by filename. It's possible to use a serial number for a filename (e.g. `123.hcl`), but we recommend you to use a timestamp as a prefix to avoid git conflicts (e.g. `20201114000000_dir1.hcl`)

An example of migration file is as follows.

```hcl
migration "state" "test" {
  dir = "dir1"
  actions = [
    "mv aws_security_group.foo aws_security_group.foo2",
    "mv aws_security_group.bar aws_security_group.bar2",
  ]
}
```

The above example is written in HCL native syntax, but you can also write them in HCL JSON syntax.
This is useful when generating a migration file from other tools.

```json
{
  "migration": {
    "state": {
      "test": {
        "dir": "dir1",
        "actions": [
          "mv aws_security_group.foo aws_security_group.foo2",
          "mv aws_security_group.bar aws_security_group.bar2"
        ]
      }
    }
  }
}
```

If you want to move a resource using `for_each`, you need to escape as follows:

```hcl
migration "state" "test" {
  dir = "dir1"
  actions = [
    "mv aws_security_group.foo[0] 'aws_security_group.foo[\"baz\"]'",
  ]
}
```

### Environment Variables

Environment variables can be accessed in migration files via the `env` variable:

```hcl
migration "state" "test" {
  dir = "dir1"
  workspace = env.TFMIGRATE_WORKSPACE
  actions = [
    "mv aws_security_group.foo aws_security_group.foo2"
  ]
}
```

### migration block

- The file must contain exactly one `migration` block.
- The first label is the migration type. There are two types of `migration` block, `state` and `multi_state`, and specify one of them.
- The second label is the migration name, which is an arbitrary string.

The file must contain only one block, and multiple blocks are not allowed, because it's hard to re-run the file if partially failed.

### migration block (state)

The `state` migration updates the state in a single directory. It has the following attributes.

- `dir` (optional): A working directory for executing terraform command. Default to `.` (current directory).
- `workspace` (optional): A terraform workspace. Defaults to "default".
- `actions` (required): Actions is a list of state action. An action is a plain text for state operation. Valid formats are the following.
  - `"mv <source> <destination>"`
  - `"xmv <source> <destination>"`
  - `"rm <addresses>...`
  - `"import <address> <id>"`
  - `"replace-provider <address> <address>"`
- `force` (optional): Apply migrations even if plan show changes
- `skip_plan` (optional): If true, `tfmigrate` will not perform and analyze a `terraform plan`.

Note that `dir` is relative path to the current working directory where `tfmigrate` command is invoked.

We could define strict block schema for action, but intentionally use a schema-less string to allow us to easily copy terraform state command to action.

Examples of migration block (state) are as follows.

#### state mv

```hcl
migration "state" "test" {
  dir = "dir1"
  actions = [
    "mv aws_security_group.foo aws_security_group.foo2",
    "mv aws_security_group.bar aws_security_group.bar2",
  ]
}
```

#### state xmv

The `xmv` command works like the `mv` command but allows usage of wildcards `*` in the source definition.
The source expressions will be matched against resources defined in the terraform state.
The matched value can be used in the destination definition via a dollar sign and their ordinal number (e.g. `$1`, `$2`, ...).
When there is ambiguity, you need to put the ordinal number in curly braces, in this case, the dollar sign need to be escaped and therefore are placed twice (e.g. `$${1}`).

For example if `foo` and `bar` in the `mv` command example above are the only 2 security group resources
defined at the top level then you can rename them using:

```hcl
migration "state" "test" {
  dir = "dir1"
  actions = [
    "xmv aws_security_group.* aws_security_group.$${1}2",
  ]
}
```

#### state rm

```hcl
migration "state" "test" {
  dir = "dir1"
  actions = [
    "rm aws_security_group.baz",
  ]
}
```

#### state import

```hcl
migration "state" "test" {
  dir = "dir1"
  actions = [
    "import aws_security_group.qux qux",
  ]
}
```

#### state replace-provider

```hcl
migration "state" "test" {
  dir = "dir1"
  actions = [
    "replace-provider registry.terraform.io/-/null registry.terraform.io/hashicorp/null",
  ]
}
```

### migration block (multi_state)

The `multi_state` migration updates states in two different directories. It is intended for moving resources across states. It has the following attributes.

- `from_dir` (required): A working directory where states of resources move from.
- `from_skip_plan` (optional): If true, `tfmigrate` will not perform and analyze a `terraform plan` in the `from_dir`.
- `from_workspace` (optional): A terraform workspace in the FROM directory. Defaults to "default".
- `to_dir` (required): A working directory where states of resources move to.
- `to_skip_plan` (optional): If true, `tfmigrate` will not perform and analyze a `terraform plan` in the `to_dir`.
- `to_workspace` (optional): A terraform workspace in the TO directory. Defaults to "default".
- `actions` (required): Actions is a list of multi state action. An action is a plain text for state operation. Valid formats are the following.
  - `"mv <source> <destination>"`
  - `"xmv <source> <destination>"`
- `force` (optional): Apply migrations even if plan show changes

Note that `from_dir` and `to_dir` are relative path to the current working directory where `tfmigrate` command is invoked.

Example of migration block (multi_state) are as follows.

#### multi_state mv

```hcl
migration "multi_state" "mv_dir1_dir2" {
  from_dir = "dir1"
  to_dir   = "dir2"
  actions = [
    "mv aws_security_group.foo aws_security_group.foo2",
    "mv aws_security_group.bar aws_security_group.bar2",
  ]
}
```

#### multi_state xmv

The `xmv` command works like the `mv` command but allows usage of
wildcards `*` in the source definition.
The wildcard expansion rules are the same as for the single state xmv.

```hcl
migration "multi_state" "mv_dir1_dir2" {
  from_dir = "dir1"
  to_dir   = "dir2"
  actions = [
    "xmv aws_security_group.* aws_security_group.$${1}2",
  ]
}
```

If you want to move all resources to another dir for merging two tfstates, you can write something like this:

```hcl
migration "multi_state" "merge_dir1_to_dir2" {
  from_dir = "dir1"
  to_dir   = "dir2"
  actions = [
    "xmv * $1",
  ]
}
```

## Integrations

You can integrate tfmigrate with your favorite CI/CD services. Examples are as follows:

- [Atlantis](https://github.com/runatlantis/atlantis): [minamijoyo/tfmigrate-atlantis-example](https://github.com/minamijoyo/tfmigrate-atlantis-example)
- [Digger](https://github.com/diggerhq/digger): [minamijoyo/tfmigrate-digger-example](https://github.com/minamijoyo/tfmigrate-digger-example)

## License

MIT
