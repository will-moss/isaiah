# [1.7.0](https://github.com/will-moss/isaiah/compare/1.6.1...1.7.0) (2024-02-18)


### Features

* **client:** added name of the resource in the menu's header to prevent mistakes ([cb97600](https://github.com/will-moss/isaiah/commit/cb97600cb3823100fbdc458d53954008680e831a))
* **client:** added prompt before container restart ([b8e6424](https://github.com/will-moss/isaiah/commit/b8e64249778ff15dc3efab458919a4a99c4377db))
* **client:** added randomly generated version number to custom.css file to prevent browser caching ([62e5b42](https://github.com/will-moss/isaiah/commit/62e5b429a46e784ea39563bf59d0884891a72d29))
* **client:** increased logs' rows height, and added alternate background color to ease reading ([65eef1d](https://github.com/will-moss/isaiah/commit/65eef1dffb0f4506618418b4b5795393631a3d5d))
* **client:** made logs' rows' background color a variable, and adjusted themes for better aesthetic ([e0acb53](https://github.com/will-moss/isaiah/commit/e0acb53c5db2c06daef6aab3d43bfb4c344435ee))
* **security:** added support for providing a sha256 hash of the authentication secret ([df13683](https://github.com/will-moss/isaiah/commit/df136836ccc6a61949eabea8532e1681fa73ceb9))

## [1.6.1](https://github.com/will-moss/isaiah/compare/1.6.0...1.6.1) (2024-02-07)


### Bug Fixes

* **multi-host:** added missing mobile control for switching host on mobile devices ([d5f14ab](https://github.com/will-moss/isaiah/commit/d5f14ab718cbcf6ab4a99c4426088ee47b8b563b))

# [1.6.0](https://github.com/will-moss/isaiah/compare/1.5.0...1.6.0) (2024-02-03)


### Features

* **project:** added support for multi-host deployment ([828116f](https://github.com/will-moss/isaiah/commit/828116f291a5783d0fc3fe892d12bc74ce0e6091))

# [1.5.0](https://github.com/will-moss/isaiah/compare/1.4.0...1.5.0) (2024-01-20)


### Bug Fixes

* **authentication:** fixed authentication denial when authentication is disabled as per settings ([72b3124](https://github.com/will-moss/isaiah/commit/72b3124d379b5d5c3ce8fb4de184edaec609cb8a))
* **client:** fixed a bug where menu navigation would break because of improper reset ([52a2e3b](https://github.com/will-moss/isaiah/commit/52a2e3b643907e899365af421b9abd2d08da990c))
* **settings:** updated default settings to increase SERVER_MAX_READ_SIZE to cover most needs ([9855fde](https://github.com/will-moss/isaiah/commit/9855fde031978568694557c853357f7635ef9f0c))


### Features

* **project:** added full support for multi-node deployment ([167668d](https://github.com/will-moss/isaiah/commit/167668df6661ebbc25717829bdeaa673d1f7cfc8))

# [1.4.0](https://github.com/will-moss/isaiah/compare/1.3.0...1.4.0) (2024-01-09)


### Bug Fixes

* **menu:** refactored the menu display logic, and now the menu's title is correct when using bulk ([18e65f6](https://github.com/will-moss/isaiah/commit/18e65f665d7bdd7ebbfd51b808286e8267ad5480))
* **mobile:** fixed CSS to display mobile controls properly even when the screen width is below 390px ([8029572](https://github.com/will-moss/isaiah/commit/8029572fd9b1861be71752b69e65cd9a60fa08c4))


### Features

* **client:** added full support for custom themes with theme picker, help, and theme save ([f8fefe8](https://github.com/will-moss/isaiah/commit/f8fefe8ce0b255f60fbe777d07370606cbd86ff6))
* **images:** added support for bulk-pulling latest images ([0629435](https://github.com/will-moss/isaiah/commit/062943575f3d8f22cdbe9af8cdededfdf6dcf635))
* **theming:** added two ready-to-use themes (dawn, and moon) based on Rosé Pine ([fd86de5](https://github.com/will-moss/isaiah/commit/fd86de5b96ff03c7b448795cc501683e469ff028))

# [1.3.0](https://github.com/will-moss/isaiah/compare/1.2.1...1.3.0) (2024-01-05)


### Features

* **containers:** added rename feature ([4e2ecd6](https://github.com/will-moss/isaiah/commit/4e2ecd61fb048ace9eff6935e8a6223343dc1896))

## [1.2.1](https://github.com/will-moss/isaiah/compare/1.2.0...1.2.1) (2024-01-05)


### Bug Fixes

* **client:** added missing mobile control for initiating a system shell ([9b0267d](https://github.com/will-moss/isaiah/commit/9b0267d4415016736a8687fb91043700627395ab))

# [1.2.0](https://github.com/will-moss/isaiah/compare/1.1.0...1.2.0) (2024-01-04)


### Features

* **client:** added full support for mobile responsiveness with adapted controls ([cf8ed79](https://github.com/will-moss/isaiah/commit/cf8ed79cfa3f91270aa2cbccb83298e5aba94832))

# [1.1.0](https://github.com/will-moss/isaiah/compare/1.0.0...1.1.0) (2024-01-04)


### Features

* **client:** added support for custom theming via custom.css file provided at runtime ([6e53323](https://github.com/will-moss/isaiah/commit/6e53323b2c08238e6181813e960ca4babc09992e))

# 1.0.0 (2024-01-04)


### Features

* **project:** first release ([fc04f02](https://github.com/will-moss/isaiah/commit/fc04f02880daac8d0a4acd4ed9f7670ce154ab99))
