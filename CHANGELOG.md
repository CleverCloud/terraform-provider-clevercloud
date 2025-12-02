# Changelog

## [1.7.1](https://github.com/CleverCloud/terraform-provider-clevercloud/compare/v1.7.0...v1.7.1) (2025-12-02)


### Bug Fixes

* **apps:** fill networkgroups on state-upgrader ([5254aa1](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/5254aa1bd762ac05bcf69aa75e12b9de3ada9243))

## [1.7.0](https://github.com/CleverCloud/terraform-provider-clevercloud/compare/v1.6.0...v1.7.0) (2025-12-01)


### Features

* **postgres:** support encryption ([efb44b6](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/efb44b66a10d40a81dd93835212f17f240c9808e))
* reboot action ([a3a335e](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/a3a335e75fb3cd26368cb3ab8e1e75e6911f3351))


### Bug Fixes

* **apps:** flavor validation ([0a9c3c4](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/0a9c3c40233feb3fcbb597d0c440f37599ca5488))
* **apps:** missing parameters for github apps ([0bd8e24](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/0bd8e24086c0b52b62d31a83458b7e67bffecd03))
* better error messages ([7b5327b](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/7b5327bfc9f39efc748a2f194a4046db064cccfc)), closes [#296](https://github.com/CleverCloud/terraform-provider-clevercloud/issues/296)
* **cellar:** empty bucket before deletion ([0e54f46](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/0e54f46d13ac9e25b225eeec0fcdb938e4095e88))

## [1.6.0](https://github.com/CleverCloud/terraform-provider-clevercloud/compare/v1.5.0...v1.6.0) (2025-11-19)


### Features

* **provider:** allow to disable NG ([1275333](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/1275333033dbe550a931d29d279a2255af9635f7))


### Bug Fixes

* **vhosts:** handle empty vhosts config ([1d4b880](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/1d4b8807eb8372d0631dd3ed43b85cc42d274efc))

## [1.5.0](https://github.com/CleverCloud/terraform-provider-clevercloud/compare/v1.4.0...v1.5.0) (2025-11-17)


### Features

* **app/adons:** networkgroup support ([1052168](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/1052168fb44847b99d1e8b08d9cf40783e4c3785))
* **application:** allow authentication to clone private repositories ([56d182d](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/56d182d9a78f2dd5206ab5dee38d000b6d773c03))
* **applications:** support Github applications ([6b0f22e](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/6b0f22ec5f69d9b3ed149b0e6d54d922bdde2394))
* **postgres:** support migration ([aead77d](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/aead77d45cb62ed7798039ac904a79473ff702de))


### Bug Fixes

* **applications:** split app creation and deployment ([1b400e3](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/1b400e39526dd19ab4afee280af5050366545782))
* **cellar:** use resource instead of manually create dependencies ([f9ec529](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/f9ec529e497a726840a5a4512b8238f279e16ecf))
* use configured endpoint instead of hardcoded default ([4e5dd8f](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/4e5dd8faf426ddad65328c7e1c38b3b6ff9a35bf))

## [1.4.0](https://github.com/CleverCloud/terraform-provider-clevercloud/compare/v1.3.0...v1.4.0) (2025-11-05)


### Features

* elasticsearch ([de8521b](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/de8521be446644aa23cba0b27dd59d3f20f5b4da))
* **lb:** add loadbalancer datasource ([4f6c864](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/4f6c8644d4e55e926d838dd52fba93abd6b7f7fd))
* replace applications when region changes ([f51cc56](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/f51cc564d0ca42f6fda482f1f8486464bbed285b)), closes [#269](https://github.com/CleverCloud/terraform-provider-clevercloud/issues/269)


### Bug Fixes

* **software:** use dedicated API when available ([f8b57f2](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/f8b57f2cbf5ccb7a59803f10f82e4ffe4567ebb9))

## [1.3.0](https://github.com/CleverCloud/terraform-provider-clevercloud/compare/v1.2.0...v1.3.0) (2025-10-16)


### Features

* **addon:** add ConfigProvider addon ([#256](https://github.com/CleverCloud/terraform-provider-clevercloud/issues/256)) ([48e5e93](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/48e5e93016aee0fb78ad8b3748ece863b0dc5470))
* **addons:** add matomo ([#253](https://github.com/CleverCloud/terraform-provider-clevercloud/issues/253)) ([e8f2206](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/e8f22064b85d0d5489178c15ed80460e3e87f18e))
* **mongodb:** return full uri (as sensitive field) ([#233](https://github.com/CleverCloud/terraform-provider-clevercloud/issues/233)) ([25229d1](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/25229d1615e12ef90f33332b37551f8e37f184c9))
* **mysql:** read read-only users generated from the console ([598fdfd](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/598fdfd746c62ad89a75de68031d506899887592))
* **runtime:** add dotnet ([#245](https://github.com/CleverCloud/terraform-provider-clevercloud/issues/245)) ([2894249](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/2894249f8cb7a53072de7d05881b923f0c2cfa4f))
* **v:** add v runtime ([#248](https://github.com/CleverCloud/terraform-provider-clevercloud/issues/248)) ([40174ed](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/40174edc88465fa664286c9b4f2dbbcfba1c3e9b))


### Bug Fixes

* **addons:** do not use plan name is only one plan ([#271](https://github.com/CleverCloud/terraform-provider-clevercloud/issues/271)) ([f6df981](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/f6df9819fa6399538ddb312655fa3c5dc1c46386))
* **apps:** trigger deployment after all others steps ([21e9348](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/21e9348e0519675f6e59fe72c8b5b878ec5b9ba1))
* **diags:** diags are always references, so that errors arn't lost ([c25d727](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/c25d72791958314bd1422af69b4e9b9b3d399b5e))
* **makefile:** make release uses ldflags too ([#255](https://github.com/CleverCloud/terraform-provider-clevercloud/issues/255)) ([b916b72](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/b916b726b7abd84185f2f7e8ea6689a10c1cd326))
* **mysql:** return full uri (with login and password) as sensitive field ([#234](https://github.com/CleverCloud/terraform-provider-clevercloud/issues/234)) ([47b82b1](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/47b82b1477efba54daccb138ffd9e7cdf06dd366))
* **postgresql:** return uri (with login/passord) as sensitive field ([#235](https://github.com/CleverCloud/terraform-provider-clevercloud/issues/235)) ([6626039](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/662603971274b769e76bc2ee1247bdfbe15c2e33))
* **python:** missing upgrade state ([eff72a0](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/eff72a02b628cda44235183f914e8d562a87c55a))

## [1.2.0](https://github.com/CleverCloud/terraform-provider-clevercloud/compare/v1.1.0...v1.2.0) (2025-09-24)


### Features

* drains ([61e5bc7](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/61e5bc7e3ae504f7d0953673e01edbfc2f6b4037))
* ruby runtime ([807f55a](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/807f55a92b21ecab48532b9b5aa133d0feb20c62))
* **ruby:** cleanup ([6e7c32f](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/6e7c32f37b54e75e5949b138ff3b032581f26f7c))


### Bug Fixes

* **addons:** support ulid formats ([3a36da6](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/3a36da6b5115fef8dc3106761f248f409fefab6e))
* **applications:** always update separatebuild, update on app updates ([df8bcbd](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/df8bcbd4910ec86de7e69d05a31771e16feb50f9))
* **apps:** urlencode vhosts for deletion ([37be86b](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/37be86b14669fdeb7f4d54d96faaad7a01438713))
* **bin:** by default, provide a stripped binary ([f68470d](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/f68470d47d68d281ab524d864d5459e9d87a7abc))
* change to typed IfIsSet ([ce0edec](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/ce0edec6c4173c9bc440abd7aba2abf080482b99))
* **frankenphp:** diag pointer and vhosts ([bedc2f1](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/bedc2f1b1abdd006b90a5855594898c3be4b1dac))
* **frankenphp:** update env/vhosts ([c2143ad](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/c2143adefddceeb44c3dd7e69ce26b03764e093c))
* **keycloak:** Fix potential malformed url ([afd92af](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/afd92afd4c5f2978895f0f19a8d5790ea45fbcfe))
* **ng:** add support for matching ng (beta) ([5c5e62f](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/5c5e62fd8ae8213b1cca6f91d7b0a4e7624a45d2))
* **nodejs:** update test website ([bc369e6](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/bc369e6b4b63bbcc773445a58577b99be57df7a4))
* **python:** reload infos from api after an update ([77c61b4](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/77c61b4cfaa4789f3add418851cb4cd269b01d92))
* **python:** urls are returned with final / by api ([8cb1ee5](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/8cb1ee52e523d12d9b56453ae792aee3fc6f4af9))
* **runtime:** avoid sticky_sessions and redirect_https reevaluation ([1ac3bea](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/1ac3beac0e99ecc4c51fce17cf80eb9bcc2d050d))
* **secrets:** mark tokens, passwords and secrets as Sensitive ([ea1bd9d](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/ea1bd9dbdb2beed0a12376d901c2b6c1a012799e))
* **test:** cleanup unused imports ([6e7c32f](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/6e7c32f37b54e75e5949b138ff3b032581f26f7c))
* **test:** domains returned by the api now always ends with / ([cbd9458](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/cbd9458b71f0230d4e4c690ede13028aaf341ed4))
* **test:** remove unused function ([6e7c32f](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/6e7c32f37b54e75e5949b138ff3b032581f26f7c))

## [1.1.0](https://github.com/CleverCloud/terraform-provider-clevercloud/compare/v1.0.1...v1.1.0) (2025-09-12)


### Features

* addon update ([fc8d6a0](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/fc8d6a0466dbe046142cd19c8c5975b557ee3484))
* bucket cannot be update ([5168d9b](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/5168d9be1c5f19ed6e07e333dafbd2eef9003567))
* credentials configuration detection ([3da1f6b](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/3da1f6b68150063f9a22dc4e4a944dfcb22c5ba4))
* frankenphp ([0ee6fd4](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/0ee6fd45eb13a062e7ba92ee11fc5e4c647866e0)), closes [#127](https://github.com/CleverCloud/terraform-provider-clevercloud/issues/127)
* keycloak update ([2a95a83](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/2a95a8307b89c355c13734f11a38950655a325e5))
* materiakv update ([3f9e8ea](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/3f9e8ea0c0547c8e445cb9392b97d110caa01338))
* metabase update ([6be8d8c](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/6be8d8c59743f7053ce25823d8cf16ac913f9482))
* mongodb update ([919fee9](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/919fee94efa88a0bc00e6ad199ad8376788fa94a))
* **mysql:** add mysql provider ([66865f8](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/66865f8c96b8f745f57050e57653952db8271ce4))
* otoroshi runtime ([b491d97](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/b491d9758132252588b3b9e883ef125cf7bcdbf6))
* psql update ([652d225](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/652d22571ab8c02ff29eb3f1a4e58ea70df2a064))
* pulsar update ([3cd75a4](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/3cd75a40e41e0b1e0a279437499010b23b428e39))
* **pulsar:** supports retention ([663c2cf](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/663c2cf87cbfe2e8e939a83103eda1641cf4bc4e))
* redis update ([432ddd2](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/432ddd2d0a0e101a8a71fcd5709ba7bd53d0b35a))
* upgrade README ([37c6df0](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/37c6df06d93a8f636502d77d3ae2e135887f6f30))


### Bug Fixes

* **app:** don't drop cleverapps.io test domain ([5a3531e](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/5a3531ede4520ebb312047045979096a9b90d96b))
* **app:** don't drop cleverapps.io test domain ([#221](https://github.com/CleverCloud/terraform-provider-clevercloud/issues/221)) ([28af15b](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/28af15b1685ea07c4cdd8ca545d9bd0aa86e0c90))
* backup file to prevent panic ([2ea9c1a](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/2ea9c1a3649a4f5d66e15ed797d9c064e1319f5b))
* isolate test with no credentials ([5377312](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/5377312f38c52114489706331c820d6eff84c718))
* **mongodb:** return field database ([76ae363](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/76ae3634521a3bed8718d27620e3c2ed0e0589a9))
* **mongodb:** test for required fields ([315c83d](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/315c83d08551dcb2d8f615b949becba33329abee))
* **pulsar:** handle non-tls clusters ([5a0206e](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/5a0206ed19100229e5ba0b81446e02fb7a19d364))
* **python:** fix test for cleverapps domain, check the final / ([dbb4054](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/dbb40545758a43df8d82629c016dfcb29ade4003))
* remove credential config detection ([5ed7b02](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/5ed7b02c09a035d23cc4f0d17e5e225d3fae477b))
* **runtime:** add missing Dependencies field in all runtime Update methods ([9390643](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/93906436dc55d388fac1cc151076dcb58b1d46cf))
* save ID as soon as available ([34e7d9c](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/34e7d9ca32dc039257da0d88f74abe6e158b4458))

## [1.0.1](https://github.com/CleverCloud/terraform-provider-clevercloud/compare/v1.0.0...v1.0.1) (2025-08-04)


### Bug Fixes

* **mongodb:** allow mongo ID  as dependency ([714d9d1](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/714d9d180f5edaa76aef45e3964a7a57abdbf1d5))
* **postgres:** fill name and creation date during import ([7c03568](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/7c035682b0d98a4d98dd5eeb566de2474c67eaa1)), closes [#189](https://github.com/CleverCloud/terraform-provider-clevercloud/issues/189)

## [1.0.0](https://github.com/CleverCloud/terraform-provider-clevercloud/compare/v0.11.1...v1.0.0) (2025-07-29)


### ⚠ BREAKING CHANGES

* **addons:** use real ID everywhere

### Features

* rust ([b7b98b1](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/b7b98b1e10b5d3a2ead2acb3beb609edf46bc977)), closes [#185](https://github.com/CleverCloud/terraform-provider-clevercloud/issues/185)


### Bug Fixes

* **addons:** use real ID everywhere ([41bb166](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/41bb166cc4619996cd4044fced62ba92a0a255d5))
* **app:** fix condition to check if app needs to be restarted on update ([f3bea0a](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/f3bea0adc57c92e97b2bfa28c44767f97acade30))
* **go:** build instance ([58e72ab](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/58e72aba341bb38cf5df62f21dd1cf50a24a42f8))
* **postgresal:** check plan == nil ([4b1500a](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/4b1500a8de8d2a75b5c9fa3ef6e65d97f14abe54)), closes [#175](https://github.com/CleverCloud/terraform-provider-clevercloud/issues/175)

## [0.11.1](https://github.com/CleverCloud/terraform-provider-clevercloud/compare/v0.11.0...v0.11.1) (2025-07-11)


### Bug Fixes

* drop vhost ([9667b27](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/9667b27921b3204f1485b88f5fac0c1c8ceb0089))
* **git:** do not throw if already up to date ([1e64c69](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/1e64c69e404c45198533f6b438dc32a253c65bfb))
* **nodejs:** clean cleverapps vhosts ([2b26ec8](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/2b26ec8c73a36e78790aff9d0aa066eaaa0b14f3))

## [0.11.0](https://github.com/CleverCloud/terraform-provider-clevercloud/compare/v0.10.0...v0.11.0) (2025-07-08)


### Features

* **postgresql:** allow to disable backups ([ebf7d26](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/ebf7d26b527459f56df3490fe28a803c766e9ac1))

## [0.10.0](https://github.com/CleverCloud/terraform-provider-clevercloud/compare/v0.9.0...v0.10.0) (2025-06-23)


### Features

* networkgroup ([c37ce75](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/c37ce75cd3ae34f858a4f24ad8fd0be4ff7e19d0))
* ng ([c37ce75](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/c37ce75cd3ae34f858a4f24ad8fd0be4ff7e19d0))
* **postgresql:** support version ([8153c96](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/8153c9670c03ca00ec35430167f734216bfccb38)), closes [#162](https://github.com/CleverCloud/terraform-provider-clevercloud/issues/162)
* **pulsar:** init ([cd83d5c](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/cd83d5c09ef0cd38308b8f96f134329ef5f33d64))


### Bug Fixes

* **java:** correct env vars to state ([c6d0e59](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/c6d0e593db22327c6e9a2a0ac9cb78aab28b2eb6)), closes [#159](https://github.com/CleverCloud/terraform-provider-clevercloud/issues/159)

## [0.9.0](https://github.com/CleverCloud/terraform-provider-clevercloud/compare/v0.8.0...v0.9.0) (2025-06-03)


### Features

* **app:** support local Git repositories ([199939f](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/199939f334e4b3fb56d20348444bf4f0f3323e9f)), closes [#147](https://github.com/CleverCloud/terraform-provider-clevercloud/issues/147)
* **app:** trigger restart on env changes ([b6494b1](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/b6494b1f5460f2542a07e41c5c94829a7d450f8b)), closes [#146](https://github.com/CleverCloud/terraform-provider-clevercloud/issues/146)
* **play2:** init ([f600a03](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/f600a0346752c8fd4c55a5058b951a287ed5d0a1)), closes [#148](https://github.com/CleverCloud/terraform-provider-clevercloud/issues/148)

## [0.8.0](https://github.com/CleverCloud/terraform-provider-clevercloud/compare/v0.7.2...v0.8.0) (2025-05-16)


### Features

* **cellar:** support update ([9ad14d8](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/9ad14d8cf7430ccc209dcf4f5bed903a7de3c9e8)), closes [#141](https://github.com/CleverCloud/terraform-provider-clevercloud/issues/141)
* **fsbucket:** init ([e06943b](https://github.com/CleverCloud/terraform-provider-clevercloud/commit/e06943ba79061c2073e39bf0776b6f48d7ef879b)), closes [#144](https://github.com/CleverCloud/terraform-provider-clevercloud/issues/144)

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
