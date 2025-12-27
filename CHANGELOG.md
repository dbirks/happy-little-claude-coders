# Changelog

## [0.2.1](https://github.com/dbirks/happy-little-claude-coders/compare/v0.2.0...v0.2.1) (2025-12-27)


### Bug Fixes

* make config PVCs conditional on workspace persistence ([a632a20](https://github.com/dbirks/happy-little-claude-coders/commit/a632a2060bf134cc433b542b758291510e8f532d))

## [0.2.0](https://github.com/dbirks/happy-little-claude-coders/compare/v0.1.0...v0.2.0) (2025-12-18)


### ⚠ BREAKING CHANGES

* Redesigned Helm chart to support multiple workspaces in a single HelmRelease.

### Features

* support multiple workspaces in single HelmRelease ([f1c3af4](https://github.com/dbirks/happy-little-claude-coders/commit/f1c3af4f0b1913a95d6424a97cf37edac0eea67e))

## [0.1.0](https://github.com/dbirks/happy-little-claude-coders/compare/happy-little-claude-coders-v0.0.1...happy-little-claude-coders-v0.1.0) (2025-12-18)


### Features

* add Claude config persistence with PVC ([4dd1984](https://github.com/dbirks/happy-little-claude-coders/commit/4dd1984e7443423b644db0f7d6db4272197373a1))
* add entrypoint and clone-repos scripts ([2906415](https://github.com/dbirks/happy-little-claude-coders/commit/29064155e44911fe569790775843dc613d185bf1))
* add GitHub App authentication with token refresh sidecar ([f1ecb9d](https://github.com/dbirks/happy-little-claude-coders/commit/f1ecb9dfeb56376b7a66ab8a684673d35eeb32b8))
* add Helm chart for happy-claude-coders ([64eeb53](https://github.com/dbirks/happy-little-claude-coders/commit/64eeb534e95364d33cd27ecdec37ba0e9a600d52))
* add init container for repo cloning ([f31ddc5](https://github.com/dbirks/happy-little-claude-coders/commit/f31ddc5a1a1c19849f124f60f1f550cd7c7e4598))
* add package.json and Dockerfile ([c71f34f](https://github.com/dbirks/happy-little-claude-coders/commit/c71f34fea6538fb3092b78a9d47a5657622154e8))
* add Release Please and CI/CD workflows ([a9b238f](https://github.com/dbirks/happy-little-claude-coders/commit/a9b238fa0c6977812cd04b7f46e6559338be8661))
* auto-continue repo cloning after GitHub authentication ([978963f](https://github.com/dbirks/happy-little-claude-coders/commit/978963f8d1f6e540b2a128b2dc49985d6907c7d1))
* migrate to Node 24 and pnpm ([f286d66](https://github.com/dbirks/happy-little-claude-coders/commit/f286d664c914890b06de0b0f2ec28a4c4a80f3d2))
* run containers as non-root user for security ([1da4932](https://github.com/dbirks/happy-little-claude-coders/commit/1da49328ae042baf8c3aa7b91c0ca1424aca75f0))


### Bug Fixes

* add pnpm setup before global package install ([326b3d2](https://github.com/dbirks/happy-little-claude-coders/commit/326b3d2d0191ddf4d3eb6c2573a1a5a3cb5e5136))
* add proper spacing before YAML comments in Chart.yaml ([445d312](https://github.com/dbirks/happy-little-claude-coders/commit/445d312694d2aaa682b2fd951e6104a2c2003b3a))
* **ci:** override entrypoint in Docker test to prevent hang ([35712fa](https://github.com/dbirks/happy-little-claude-coders/commit/35712faf1f9db32ba43e77583775a5f61480c617))
* print entrypoint banners to stderr to avoid polluting stdout ([992c1a7](https://github.com/dbirks/happy-little-claude-coders/commit/992c1a7f10b5c632a693345a6ab048c34f5f6a32))
* set pnpm ENV variables instead of using pnpm setup ([67bcb61](https://github.com/dbirks/happy-little-claude-coders/commit/67bcb61df6a79e148d3fbf0ccf8e533a8ebac10f))
* use RELEASE_PLEASE_TOKEN for PR creation ([8024303](https://github.com/dbirks/happy-little-claude-coders/commit/802430300a7605eeb0f283a9013a71588709709d))
