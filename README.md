# tfmigrate-storage

This repository maintains a history storage implementation for [tfmigrate](https://github.com/minamijoyo/tfmigrate).

It's natural for most of tfmigrate users to expect the history storage implementation to have the same authentication settings and priorities as the Terraform core backend. To do this, we need to reuse some code from the upstream. The problem is that the license of hashicorp/terraform is the MPL2.0, but the tfmigrate is distributed under the terms of MIT.

This repository contains a "Larger Work" (MPL2.0 Section 1.7) of [hashicorp/terraform](https://github.com/hashicorp/terraform).

This repository is kept separate from the tfmigrate only for license reason and doesn't accept any feature request or bug report. Please open it to the [tfmigrate](https://github.com/minamijoyo/tfmigrate) repository.

Since this repository is dedicated to the tfmigrate, the compatibility of the public interface as a Go library is not guaranteed and there is no versioning policy. If you want to test storage implementation changes with the tfmigrate before merging, use the [replace](https://go.dev/ref/mod#go-mod-file-replace) directive.

## License

Mozilla Public License v2.0
