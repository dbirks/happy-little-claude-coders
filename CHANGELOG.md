# Changelog

## [0.12.5](https://github.com/dbirks/happy-little-claude-coders/compare/v0.12.4...v0.12.5) (2026-01-16)


### Bug Fixes

* handle workspace PVC with non-git content in init container ([dd6b0c2](https://github.com/dbirks/happy-little-claude-coders/commit/dd6b0c29c91b80ed8229226ccf0e47e363fcbf65))

## [0.12.4](https://github.com/dbirks/happy-little-claude-coders/compare/v0.12.3...v0.12.4) (2026-01-16)


### Bug Fixes

* handle workspace PVC with non-git content in clone-repos ([09065fd](https://github.com/dbirks/happy-little-claude-coders/commit/09065fdd5f73959e8c29f4c1a569999cb397d52b))

## [0.12.2](https://github.com/dbirks/happy-little-claude-coders/compare/v0.12.1...v0.12.2) (2026-01-16)


### Bug Fixes

* skip Claude Code theme selection and onboarding prompts ([74da410](https://github.com/dbirks/happy-little-claude-coders/commit/74da41053d0f41136cd109f3147f13f23686e7d8))

## [0.12.1](https://github.com/dbirks/happy-little-claude-coders/compare/v0.12.0...v0.12.1) (2026-01-16)


### Bug Fixes

* create Happy config PVCs independent of workspace.persistent ([6ce34d3](https://github.com/dbirks/happy-little-claude-coders/commit/6ce34d39a5e24ea64434fe1a7b8957eb65778faf))

## [0.12.0](https://github.com/dbirks/happy-little-claude-coders/compare/v0.11.3...v0.12.0) (2026-01-15)


### Features

* add per-workspace Happy config PVC support for RWO storage ([1525240](https://github.com/dbirks/happy-little-claude-coders/commit/15252400bf41f84b94b4ba27b72bbb788840c2cf))

## [0.11.3](https://github.com/dbirks/happy-little-claude-coders/compare/v0.11.2...v0.11.3) (2026-01-15)


### Bug Fixes

* write Claude OAuth token to config file for Happy daemon ([ad40799](https://github.com/dbirks/happy-little-claude-coders/commit/ad4079982efedd3ab4a436c9cea0b8afd72f43e2))

## [0.11.2](https://github.com/dbirks/happy-little-claude-coders/compare/v0.11.1...v0.11.2) (2026-01-15)


### Bug Fixes

* remove redundant claude startup, happy daemon manages it ([e166f16](https://github.com/dbirks/happy-little-claude-coders/commit/e166f16687db9e08658056dbbc9b4629e009e17c))

## [0.11.1](https://github.com/dbirks/happy-little-claude-coders/compare/v0.11.0...v0.11.1) (2026-01-15)


### Bug Fixes

* switch image build from pnpm to npm and bump chart to v0.11.1 ([74f5c7d](https://github.com/dbirks/happy-little-claude-coders/commit/74f5c7dfb8c2cc31327208dcd6f3747246adcbd3))

## [0.11.0](https://github.com/dbirks/happy-little-claude-coders/compare/v0.10.0...v0.11.0) (2026-01-15)


### Features

* add local Kind cluster testing setup ([b1ae23c](https://github.com/dbirks/happy-little-claude-coders/commit/b1ae23c2221400f6074f0a63f91bfa60b17c7045))
* install Claude Code via npm and auto-start with Happy daemon ([df2eb90](https://github.com/dbirks/happy-little-claude-coders/commit/df2eb904f5ae6c47774c56fe5d3cde3b27d08604))

## [0.10.0](https://github.com/dbirks/happy-little-claude-coders/compare/v0.9.1...v0.10.0) (2026-01-12)


### Features

* implement retry loop for happy CLI authentication ([d70fa0e](https://github.com/dbirks/happy-little-claude-coders/commit/d70fa0ec00e26e51496e553a2d8fe7537e6d3e3c))


### Bug Fixes

* use happy daemon start-sync for proper background operation ([ed4f312](https://github.com/dbirks/happy-little-claude-coders/commit/ed4f312be40ab7c7c437dd92a3980b78b1a2654e))

## [0.9.1](https://github.com/dbirks/happy-little-claude-coders/compare/v0.9.0...v0.9.1) (2026-01-12)


### Bug Fixes

* expect script now stays alive indefinitely ([13d6a69](https://github.com/dbirks/happy-little-claude-coders/commit/13d6a69f6a467ce3480210c41b0049c0b1f85990))
* simplify happy launch with unbuffer ([cb55356](https://github.com/dbirks/happy-little-claude-coders/commit/cb55356ed19fec57be07e30e1d576dc3cbc85673))

## [0.9.0](https://github.com/dbirks/happy-little-claude-coders/compare/v0.8.0...v0.9.0) (2026-01-12)


### Features

* use expect script to auto-select mobile app auth ([970be5f](https://github.com/dbirks/happy-little-claude-coders/commit/970be5f605f9862725b1a49f7fcca5efc06f3995))

## [0.8.0](https://github.com/dbirks/happy-little-claude-coders/compare/v0.7.0...v0.8.0) (2026-01-12)


### Features

* auto-select mobile app auth in happy CLI ([cbbef0a](https://github.com/dbirks/happy-little-claude-coders/commit/cbbef0a1cbc7b2a7654e69de9367186e56e08b1b))

## [0.7.0](https://github.com/dbirks/happy-little-claude-coders/compare/v0.6.0...v0.7.0) (2026-01-12)


### Features

* run happy CLI with pseudo-TTY via script command ([09dd813](https://github.com/dbirks/happy-little-claude-coders/commit/09dd813a6f5fc90164c339dbd5a308516c0b3400))

## [0.6.0](https://github.com/dbirks/happy-little-claude-coders/compare/v0.5.0...v0.6.0) (2026-01-10)


### Features

* use --no-qr flag for text-based pairing code ([21d17b6](https://github.com/dbirks/happy-little-claude-coders/commit/21d17b6891587621d22f15669a448e50dbf32ced))

## [0.5.0](https://github.com/dbirks/happy-little-claude-coders/compare/v0.4.1...v0.5.0) (2026-01-10)


### Features

* **chart:** add shared PVC for happy CLI config ([790a788](https://github.com/dbirks/happy-little-claude-coders/commit/790a788b6434a75b1bf2475dd3bb1d9e830853fb))

## [0.4.1](https://github.com/dbirks/happy-little-claude-coders/compare/v0.4.0...v0.4.1) (2026-01-10)


### Performance Improvements

* use pre-installed happy-coder instead of dlx ([d3c8a48](https://github.com/dbirks/happy-little-claude-coders/commit/d3c8a48cd27977b651958ef91cbcca01abdc89bf))

## [0.4.0](https://github.com/dbirks/happy-little-claude-coders/compare/v0.3.1...v0.4.0) (2026-01-09)


### Features

* auto-start happy CLI on container startup ([92e4b74](https://github.com/dbirks/happy-little-claude-coders/commit/92e4b744bff951f75d202558f4e67d14fc7db550))

## [0.3.1](https://github.com/dbirks/happy-little-claude-coders/compare/v0.3.0...v0.3.1) (2026-01-09)


### Bug Fixes

* keep workspace container alive with sleep infinity ([666e69f](https://github.com/dbirks/happy-little-claude-coders/commit/666e69fad7f798d16b8e94aa96b43c625a2def06))

## [0.3.0](https://github.com/dbirks/happy-little-claude-coders/compare/v0.2.1...v0.3.0) (2026-01-09)


### Features

* add Claude Code skill for GitHub App setup ([077f250](https://github.com/dbirks/happy-little-claude-coders/commit/077f25077e47d68c0758df8e117d0a983b948475))
* add Playwright MCP config for browser-assisted setup ([afbf721](https://github.com/dbirks/happy-little-claude-coders/commit/afbf72190633b7c37ed2948771c81a0a5765c64e))
* add workflow to build sidecar image on path changes ([bc49a8b](https://github.com/dbirks/happy-little-claude-coders/commit/bc49a8babd334e37f4d3d3bc2b44e81e8ac7d477))
* enhance GitHub App setup skill with browser automation options ([0972460](https://github.com/dbirks/happy-little-claude-coders/commit/0972460ff849f900bdefbc7cd2a11758d9d083a0))


### Bug Fixes

* **chart:** set HOME env var in clone-repos init container ([aa7870b](https://github.com/dbirks/happy-little-claude-coders/commit/aa7870b2b9ae7ae5b49960e3ba941d4157732235))
* copy all Go source files in sidecar Dockerfile ([54dea13](https://github.com/dbirks/happy-little-claude-coders/commit/54dea13c60b84e1d3f839261da77d5400086a17d))
* update ghinstallation to v2.17.0 and go-github to v75 ([9452ccd](https://github.com/dbirks/happy-little-claude-coders/commit/9452ccd9b6acdf87878daea45995bc0aefaf6825))
* use alpine base for shell support in init containers ([8e44d84](https://github.com/dbirks/happy-little-claude-coders/commit/8e44d847a7f8e3bab9f2dd755722a2a179198524))
* use AppsTransport for CreateInstallationToken endpoint ([3819e0c](https://github.com/dbirks/happy-little-claude-coders/commit/3819e0c4bd4f5dc1723e61ff2c90f963f1ce1263))

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
