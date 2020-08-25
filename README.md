# tfmigrate
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![GitHub release](https://img.shields.io/github/release/minamijoyo/tfmigrate.svg)](https://github.com/minamijoyo/tfmigrate/releases/latest)
[![GoDoc](https://godoc.org/github.com/minamijoyo/tfmigrate/tfmigrate?status.svg)](https://godoc.org/github.com/minamijoyo/tfmigrate)

A tfstate migration tool for GitOps.

## Features

- GitOps-friendly: Write terraform state mv/rm/import commands in HCL, plan and apply it.
- Refactoring across tfstates: Move resources to other tfstates to split and merge tfstates easily.
- Dry-run: Simulate state operations with a temporary local tfstate and check to see if terraform plan has no changes after the migration without updating remote tfstate.

## Why?

If you have been using Terraform in production for a long time, tfstate manipulations are unavoidable for various reasons. As you know, the terraform state command is your friend, but it's error-prone and not suitable for a GitOps workflow.

In team development, Terraform configurations are generally managed by git and states are shared via remote state storage which is outside of version control. It's a best practice for Terraform.
However, most Terraform refactorings require not only configuration changes, but also state operations such as state mv/rm/import. It's not desirable to change the remote state before merging configuration changes. Your colleagues may be working on something else and your CI/CD pipeline continuously plan and apply their changes automatically. At the same time, you probably want to check to see if terraform plan has no changes after the migration before merging configuration changes.

To fit into the GitOps workflow, the answer is obvious. We should commit all terraform state operations to git.
This brings us to a new paradigm, Terraform state operation as Code!

## Install

### Homebrew

If you are macOS user:

```
$ brew install minamijoyo/tfmigrate/tfmigrate
```

### Download

Download the latest compiled binaries and put it anywhere in your executable path.

https://github.com/minamijoyo/tfmigrate/releases

### Source

If you have Go 1.15+ development environment:

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
    plan     Compute a new state
```

```
$ tfmigrate plan --help
Usage: tfmigrate plan <PATH>

Plan computes a new state by applying state migration operations to a temporary state.
It will fail if terraform plan detects any diffs with the new state.

Arguments
  PATH               A path of migration file
```

```
$ tfmigrate apply --help
Usage: tfmigrate apply <PATH>

Apply computes a new state and pushes it to remote state.
It will fail if terraform plan detects any diffs with the new state.

Arguments
  PATH               A path of migration file
```

## Settings

You can customize the behavior by setting environment variables.

- `TFMIGRATE_LOG`: A log level. Valid values are `TRACE`, `DEBUG`, `INFO`, `WARN`, `ERROR`. Default to `INFO`.
- `TFMIGRATE_EXEC_PATH`: A string how terraform command is executed. Default to `terraform`. It's intended to inject a wrapper command such as direnv. e.g.) `direnv exec . terraform`.

## Migration file examples

Examples of migration file are follows. The syntax is described below.

### state mv

```hcl
migration "state" "test" {
  dir = "dir1"
  actions = [
    "mv aws_security_group.foo aws_security_group.foo2",
    "mv aws_security_group.bar aws_security_group.bar2",
  ]
}
```

### state rm

```hcl
migration "state" "test" {
  dir = "dir1"
  actions = [
    "rm aws_security_group.baz",
  ]
}
```

### state import

```hcl
migration "state" "test" {
  dir = "dir1"
  actions = [
    "import aws_security_group.qux qux",
  ]
}
```

### multi_state mv

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

## Migration file syntax

### migration file

- A migration file must be written in the HCL2.
- The extension of file must be `.hcl`(for HCL native syntax) or `.hcl.json`(for HCL JSON syntax).

There are no rules for naming files for now, but we will limit it when we add support for an history management feature.

### migration block

- The file must contain exactly one `migration` block.
- The first label is the migration type. There are two types of `migration` block, `state` and `multi_state`, and specify one of them.
- The second label is the migration name, which is an arbitrary string.

The file must contain only one block, and multiple blocks are not allowed, because it's hard to re-run the file if partially failed.

### migration block (state)

The `state` migration updates the state in a single directory. It has the following attributes.

- `dir` (optional): A working directory for executing terraform command. Default to `.` (current directory).
- `actions` (required): Actions is a list of state action. An action is a plain text for state operation. Valid formats are the following.
  - `"mv <source> <destination>"`
  - `"rm <addresses>...`
  - `"import <address> <id>"`

Note that `dir` is relative path to the current working directory where `tfmigrate` command is invoked.

We could define strict block schema for action, but intentionally use a schema-less string to allow us to easily copy terraform state command to action.

### migration block (multi_state)

The `multi_state` migration updates states in two different directories. It is intended for moving resources across states. It has the following attributes.

- `from_dir` (required): A working directory where states of resources move from.
- `to_dir` (required): A working directory where states of resources move to.
- `actions` (required): Actions is a list of multi state action. An action is a plain text for state operation. Valid formats are the following.
  - `"mv <source> <destination>"`

Note that `from_dir` and `to_dir` are relative path to the current working directory where `tfmigrate` command is invoked.

## License

MIT
