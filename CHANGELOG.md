# [1.8.0](https://github.com/will-moss/isaiah/compare/1.7.0...1.8.0) (2024-03-29)


### Bug Fixes

* **client:** added more constraints to prevent unfortunate multi-layered popups when using shortcuts ([7203e2f](https://github.com/will-moss/isaiah/commit/7203e2f6d9af01499dd99b4558305251df22fa91))
* **client:** changed the shortcut for Parameters because of collision with Pull, and changed helpers ([71dc832](https://github.com/will-moss/isaiah/commit/71dc8320c248370c99b53e0bf7df0eaa6286141e))
* **client:** fixed an annoying reset while in prompt / menu triggered by logs received in background ([5e6307f](https://github.com/will-moss/isaiah/commit/5e6307f0532d1df54b05f0d5f0ab7f707e5b7a3e))
* **client:** fixed the client's state not associating agent and host picking to being inside a menu ([2208068](https://github.com/will-moss/isaiah/commit/22080681def4c950d474e6dd0e4b85708e6542a6))
* **client:** on mobile display, removed the control to copy logs because it can't fit ([71f0daa](https://github.com/will-moss/isaiah/commit/71f0daa6d7c284a39e5498d9c598f99952e2df35))
* **logs:** fixed a styling issue causing stripped log lines' background to be cut ([6b3437f](https://github.com/will-moss/isaiah/commit/6b3437f077d7ae11a8f8fafa0c66511c128d3de3))


### Features

* **client:** added control to copy logs to clipboard ([2183060](https://github.com/will-moss/isaiah/commit/2183060324ed031f0debe1e81ec688346a0ac96e))
* **client:** added full support for client-side persistent parameters management ([ac03b03](https://github.com/will-moss/isaiah/commit/ac03b0318af454307369fe06af2e9aed2a3d2f6e))
* **client:** added prompt before container pause when using keyboard shortcuts ([eb3c9fe](https://github.com/will-moss/isaiah/commit/eb3c9fe0769609b91a8f73773362e391ef04cc7d))
* **client:** added search feature for Docker objects in the client ([4cc6708](https://github.com/will-moss/isaiah/commit/4cc670817c9d4c3150d5b1d35058ed09961fdd13))
* **client:** added support for log lines wrap ([fa23f0e](https://github.com/will-moss/isaiah/commit/fa23f0ed05d3b09d5d654d01dd3f6d2d7b26ee73))
* **client:** added support for prompts on keyboard shortcuts, and persistent user settings (WIP) ([6f1012a](https://github.com/will-moss/isaiah/commit/6f1012a93a0e73e77a688ad7042e9764bfa8db31))
* **client:** added support for toggling log lines' timestamp display ([32b9c1a](https://github.com/will-moss/isaiah/commit/32b9c1a883ea36a0228d57fa70391a45d7fcbe14))

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
