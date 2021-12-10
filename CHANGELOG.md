## master (Unreleased)

BREAKING CHANGES:

* Deprecate the tfmigrate plan --out=tfplan option ([#63](https://github.com/minamijoyo/tfmigrate/pull/63))

Deprecate the tfmigrate plan --out=tfplan option without replacement and it will be removed in a future release.
It was based on a bug prior to Terraform 1.1 and doesn't work with Terraform 1.1 or later.
Fortunately, Terraform 1.1 added [a new moved block feature](https://www.hashicorp.com/blog/terraform-1-1-improves-refactoring-and-the-cloud-cli-experience), so some use-cases could be covered by the moved block.

NEW FEATURES:

* Support workspace for state migrator ([#61](https://github.com/minamijoyo/tfmigrate/pull/61))

ENHANCEMENTS:

* Add support for Terraform v1.1 ([#50](https://github.com/minamijoyo/tfmigrate/pull/50))
* /usr/local/bin directory is not guaranteed to exist ([#60](https://github.com/minamijoyo/tfmigrate/pull/60))

## 0.2.13 (2021/11/30)

NEW FEATURES:

* Add list command for listing unapplied migrations ([#56](https://github.com/minamijoyo/tfmigrate/pull/56))

ENHANCEMENTS:

* Update Go to v1.17.3 and Alpine to 3.14 ([#59](https://github.com/minamijoyo/tfmigrate/pull/59))
* Add Apple Silicon (ARM 64) build ([#57](https://github.com/minamijoyo/tfmigrate/pull/57))

## 0.2.12 (2021/11/25)

ENHANCEMENTS:

* Log terraform command before run ([#55](https://github.com/minamijoyo/tfmigrate/pull/55))
* Add an example for integrating tfmigrate with atlantis ([#54](https://github.com/minamijoyo/tfmigrate/pull/54))
* Add a tip to README for using for_each ([#52](https://github.com/minamijoyo/tfmigrate/pull/52))

## 0.2.11 (2021/10/28)

ENHANCEMENTS:

* Remove a positional dir parameter from TerraformCLI ([#49](https://github.com/minamijoyo/tfmigrate/pull/49))

Note: This changes contains a breaking change for tfexec package, but it doesn't affect tfmigrate CLI users.

## 0.2.10 (2021/10/15)

ENHANCEMENTS:

* Skip workspace select if already selected ([#47](https://github.com/minamijoyo/tfmigrate/pull/47))
* Cancel test if stale ([#48](https://github.com/minamijoyo/tfmigrate/pull/48))

## 0.2.9 (2021/09/04)

BUG FIXES:

* Fix a bug of multi_state doesn't show diffs in to_dir if force=true ([#40](https://github.com/minamijoyo/tfmigrate/pull/40))

ENHANCEMENTS:

* Restrict permissions for GitHub Actions ([#41](https://github.com/minamijoyo/tfmigrate/pull/41))
* Set timeout for GitHub Actions ([#42](https://github.com/minamijoyo/tfmigrate/pull/42))

## 0.2.8 (2021/09/03)

NEW FEATURES:

* Add role_arn to S3StorageConfig ([#33](https://github.com/minamijoyo/tfmigrate/pull/33))

## 0.2.7 (2021/08/12)

NEW FEATURES:

* Add a new flag --out for saving a plan file after dry-run migrations ([#37](https://github.com/minamijoyo/tfmigrate/pull/37))

## 0.2.6 (2021/08/03)

NEW FEATURES:

* Support workspaces for multi-state migrations ([#31](https://github.com/minamijoyo/tfmigrate/pull/31))

## 0.2.5 (2021/06/10)

ENHANCEMENTS:

* Add support for Terraform v1.0 ([#28](https://github.com/minamijoyo/tfmigrate/pull/28))

All we need was adding Terraform v1.0.0 to a test matrix. This means it works with tfmigrate v0.2.4 as it is.

## 0.2.4 (2021/05/08)

ENHANCEMENTS:

* Update aws-sdk-go to v1.37.0 to support AWS SSO ([#26](https://github.com/minamijoyo/tfmigrate/pull/26))

## 0.2.3 (2021/04/19)

BUG FIXES:

* Create a plugin cache directory in advance ([#12](https://github.com/minamijoyo/tfmigrate/pull/12))
* Fix CI fail for TestExecutorDir in ubuntu-20.04 ([#18](https://github.com/minamijoyo/tfmigrate/pull/18))

ENHANCEMENTS:

* Support Terraform v0.15 ([#17](https://github.com/minamijoyo/tfmigrate/pull/17))

All we need was adding Terraform v0.15.0 to a test matrix. This means it works with tfmigrate v0.2.2 as it is.

## 0.2.2 (2020/12/28)

ENHANCEMENTS:

* Show diffs in log if force is set to true ([#11](https://github.com/minamijoyo/tfmigrate/pull/11))

## 0.2.1 (2020/12/17)

NEW FEATURES:

* Added force option to state and multistate migrations ([#10](https://github.com/minamijoyo/tfmigrate/pull/10))

ENHANCEMENTS:

* Support Terraform v0.14 ([#7](https://github.com/minamijoyo/tfmigrate/pull/7))

All we need was adding Terraform v0.14.0 to a test matrix. This means it works with tfmigrate v0.2.0 as it is.

## 0.2.0 (2020/11/18)

NEW FEATURES:

* Add support for migration history management ([#2](https://github.com/minamijoyo/tfmigrate/pull/2))

You can now keep track of which migrations have been applied and apply all unapplied migrations in sequence. The migration history can be saved to `local` or `s3` storage. See the `Configurations` section in the README for how to configure it.
If your cloud provider has not been supported yet, feel free to open an issue or submit a pull request. As a workaround, you can use `local` storage and synchronize a history file to your cloud storage with a wrapper script.

ENHANCEMENTS:

* Fix unstable tests ([#8](https://github.com/minamijoyo/tfmigrate/pull/8))
* Use hashicorp/aws-sdk-go-base to authenticate s3 storage ([#9](https://github.com/minamijoyo/tfmigrate/pull/9))

## 0.1.1 (2020/11/05)

BUG FIXES:

* Parse a state action string like a shell ([#6](https://github.com/minamijoyo/tfmigrate/pull/6))

## 0.1.0 (2020/09/17)

Initial release
