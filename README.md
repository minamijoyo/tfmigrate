# tfmigrate
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![GitHub release](https://img.shields.io/github/release/minamijoyo/tfmigrate.svg)](https://github.com/minamijoyo/tfmigrate/releases/latest)
[![GoDoc](https://godoc.org/github.com/minamijoyo/tfmigrate/tfmigrate?status.svg)](https://godoc.org/github.com/minamijoyo/tfmigrate)

A tfstate migration tool for GitOps.

## Features

- GitOps-friendly: Write terraform state mv/rm/import commands in HCL.
- Refactoring across tfstates: Move resources to other tfstates to merge and split tfstates easily.
- Plan (dry-run mode): Simulate state operations with a temporary local tfstate and check to see if terraform plan has no changes after the migration without updating remote tfstate.
- Apply: Compute a new state with plan and ensure no changes and push it to remote.

## Why?

If you have been operating Terraform for a long time, tfstate manipulation is unavoidable for various reasons. As you know, terraform state command is your friend, but it's error-prone and difficult to do safely with GitOps.

In team development, Terraform configurations are generally managed by git and states are shared via remote state storage outside of version control. It's a best practice for Terraform.
However, most Terraform refactorings require not only configuration changes, but also state operations such as state mv/rm/import.
It's not desirable to change the remote state before merging configuration changes. Your colleagues may be working on something else and your CI/CD pipeline continuously plan and apply their changes automatically.

To fit into a GitOps workflow, the answer is obvious. We should commit all terraform state operations to git.
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

## License

MIT
