## master (Unreleased)

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
