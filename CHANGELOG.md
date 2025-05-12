# Changelog

## [0.7.2](https://github.com/CleverCloud/terraform-provider-clevercloud/compare/v0.7.1...v0.7.2) (2025-05-12)


### Bug Fixes

* **apps:** cleverapps default domain ([cd3c305](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/cd3c30524cf6ea24af78eef0cdba2b4269a8fcf6))
* **git:** return erro if commit does not exists ([cd6a7f3](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/cd6a7f3d240c627b63d30d91d3b5a957e8b59aac))
* magic image ([d093643](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/d093643ade4e7789793fe93dc9b771c9bde00417))
* push ([9dea69c](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/9dea69ce2ff6a821d111a97667cabb5331e0fac5))

## [0.7.1](https://github.com/CleverCloud/terraform-provider-clevercloud/compare/v0.7.0...v0.7.1) (2025-04-01)


### Bug Fixes

* **nodejs:** apply build-flavor ([ef82251](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/ef82251922b8f0f03f1aff54bdd7abc955a46812)), closes [#128](https://github.com/CleverCloud/terraform-provider-clevercloud/issues/128)

## [0.7.0](https://github.com/CleverCloud/terraform-provider-clevercloud/compare/v0.6.0...v0.7.0) (2025-03-19)


### Features

* add docker update ([ffd278b](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/ffd278bb4facd753fe0d9616a0992212836ac6bd))
* add scala update ([788a1cd](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/788a1cd551dbb511e60ded77b9b2ee2f7acbea01))
* add static update ([69497d5](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/69497d5225c7b6075e2b769eb051fa3869872164))
* go ([c07c7d9](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/c07c7d99b24c606ce9a5346c0d5f97ea71db1999))
* java update ([d943091](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/d9430912bab1bd735a912cbe9bd7307eb5e0971a))
* python update ([b5cb9eb](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/b5cb9ebd9c857e4de5cd91b68b2126754ec25702))


### Bug Fixes

* **plan:** remove test plan ([16eaedb](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/16eaedb5ab55b13d6c2d4a7c2ea72750adae7b32))
* **tests:** materiaKV add status DELETED ([9f3f3e3](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/9f3f3e355be7a0b2f86a8a21008d77b846ee0488))

## [0.6.0](https://github.com/CleverCloud/terraform-provider-clevercloud/compare/v0.5.1...v0.6.0) (2025-02-19)


### Features

* commit docs update ([1b08092](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/1b08092032cb2bedef894aba0b290711645bdfa6))
* **PHP:** implem update ([b2b2fef](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/b2b2fef5f190cd246288649685084a373b835f5b))
* redis ([73eb8a0](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/73eb8a099e529ecf634a76548b270c3cf50f6e57))
* update ([9b682c3](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/9b682c3d1937281737ba563267be3b411f47505c))
* use stefanzweifel/git-auto-commit-action@v5 to commit ([8f36ccb](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/8f36ccbda19ee1468f99c986625c2b1ca9dfa06c))


### Bug Fixes

* **docker:** remove default values for specific docker attributes ([538e703](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/538e703881f1ea1643c145981146010a30b3b606))
* **docker:** support IPv6 CIDR ([3d94909](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/3d949097114d355fc72cdeaf36613871d454136c))
* **org:** make the organisation parameter optional ([05ace76](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/05ace76560a5568c0f07ad8ceaa893b8bb926c9c))

## [0.5.1](https://github.com/CleverCloud/terraform-provider-clevercloud/compare/v0.5.0...v0.5.1) (2024-12-27)


### 🐛 Bug Fixes

* **runtime:** deploy right git ref ([d5607a6](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/d5607a63ac030d97dd1e0c11f41d3457860bb33d)), closes [#96](https://github.com/CleverCloud/terraform-provider-clevercloud/issues/96)

## [0.5.0](https://github.com/CleverCloud/terraform-provider-clevercloud/compare/v0.4.1...v0.5.0) (2024-12-24)


### 🚀 Features

* keycloak ([#90](https://github.com/CleverCloud/terraform-provider-clevercloud/issues/90)) ([d9df803](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/d9df803de7ade0a60bd69e6febbfd8f5fc056c3f)), closes [#86](https://github.com/CleverCloud/terraform-provider-clevercloud/issues/86)
* metabase ([5d9dc9a](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/5d9dc9a1bd6e28b92171e0994b41f5988bd344ad))


### 🐛 Bug Fixes

* Read for tests ([5d9dc9a](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/5d9dc9a1bd6e28b92171e0994b41f5988bd344ad))
* suggested chances ([5d9dc9a](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/5d9dc9a1bd6e28b92171e0994b41f5988bd344ad))

## [0.4.1](https://github.com/CleverCloud/terraform-provider-clevercloud/compare/v0.4.0...v0.4.1) (2024-07-12)


### 🐛 Bug Fixes

* **mongodb:** plan is nil ([b4545cb](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/b4545cb86561e55c54baefc19574be7d874070b6)), closes [#82](https://github.com/CleverCloud/terraform-provider-clevercloud/issues/82)

## [0.4.0](https://github.com/CleverCloud/terraform-provider-clevercloud/compare/v0.3.0...v0.4.0) (2024-07-03)


### 🚀 Features

* docker ([#79](https://github.com/CleverCloud/terraform-provider-clevercloud/issues/79)) ([8c02ab3](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/8c02ab3b5d5823487abd7707cabb6ba8d7616bf2))

## [0.3.0](https://github.com/CleverCloud/terraform-provider-clevercloud/compare/v0.2.0...v0.3.0) (2024-06-10)


### 🚀 Features

* set par as default regions ([fcb7dab](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/fcb7dab06844dacf5ede92f73cfee920f4a98855))


### 🐛 Bug Fixes

* remove all hardcoded region parameters + default="par" ([80b0224](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/80b0224a531ac3d5f36d69040cc0361142446da8))

## [0.2.0](https://github.com/CleverCloud/terraform-provider-clevercloud/compare/v0.1.1...v0.2.0) (2024-05-03)


### 🚀 Features

* Mongo ([936f845](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/936f8451a7118b66141ca59301315ec384a58909))


### 🐛 Bug Fixes

* **test:** change ressource name (%s -&gt; rName) ([0c3a592](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/0c3a5922629f61dcf9a3b1a20b9b0bb2cb589857))

## [0.1.1](https://github.com/CleverCloud/terraform-provider-clevercloud/compare/v0.1.0...v0.1.1) (2024-04-15)


### 🐛 Bug Fixes

* **materia:** rename resource to match product ([#66](https://github.com/CleverCloud/terraform-provider-clevercloud/issues/66)) ([075a5a1](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/075a5a122567efa4a19da4f6aae57261fb7480c7))

## [0.1.0](https://github.com/CleverCloud/terraform-provider-clevercloud/compare/v0.0.16...v0.1.0) (2024-04-12)


### 🚀 Features

* MateriaKV ([#65](https://github.com/CleverCloud/terraform-provider-clevercloud/issues/65)) ([0938b93](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/0938b93639934f7b3001ddfa03423a2e321c47b1))
* **python:** new python runtime ([#58](https://github.com/CleverCloud/terraform-provider-clevercloud/issues/58)) ([2981e50](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/2981e5097520c62d6d4a15306752f5c9d404299c))
