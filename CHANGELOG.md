## [1.36.8](https://github.com/will-moss/isaiah/compare/1.36.7...1.36.8) (2026-02-28)


### Bug Fixes

* **publishing:** attempt at fixing the release process since the update of Docker Engine ([a0147c6](https://github.com/will-moss/isaiah/commit/a0147c61ea52b118d342d9352601c740f7e5ba48))

## [1.36.7](https://github.com/will-moss/isaiah/compare/1.36.6...1.36.7) (2026-02-28)


### Bug Fixes

* **dummy:** forcefully creating a new version ([86a9ee7](https://github.com/will-moss/isaiah/commit/86a9ee700e78e6864f6039170760be4fe8943112))

## [1.36.6](https://github.com/will-moss/isaiah/compare/1.36.5...1.36.6) (2026-02-28)


### Bug Fixes

* **dummy:** forcefully creating a new version ([53eba6b](https://github.com/will-moss/isaiah/commit/53eba6b9c06297b24659d8ce4310fc7bf039898b))

## [1.36.5](https://github.com/will-moss/isaiah/compare/1.36.4...1.36.5) (2026-02-28)


### Bug Fixes

* **dummy:** forcefully creating a new version ([8760cda](https://github.com/will-moss/isaiah/commit/8760cdaa14fe064b7ae153bcad30fb41ea614288))

## [1.36.4](https://github.com/will-moss/isaiah/compare/1.36.3...1.36.4) (2026-02-28)


### Bug Fixes

* **dummy:** forcefully creating a new version ([e274471](https://github.com/will-moss/isaiah/commit/e274471dc68b95fa126900b60afc268633aa2eb1))

## [1.36.3](https://github.com/will-moss/isaiah/compare/1.36.2...1.36.3) (2026-02-28)


### Bug Fixes

* **compatibility:** updated the Docker SDK for Go tu work with Docker 29+, and adjusted the code ([a3b8ed6](https://github.com/will-moss/isaiah/commit/a3b8ed6ccf849a0a71336f9430a0650ddc976d66)), closes [#32](https://github.com/will-moss/isaiah/issues/32)

## [1.36.2](https://github.com/will-moss/isaiah/compare/1.36.1...1.36.2) (2025-07-23)


### Bug Fixes

* **client:** bypass Ctrl key to prevent clash with native browser shortcuts, thanks @Lumberj3ck ([22908df](https://github.com/will-moss/isaiah/commit/22908df87fe0a4c792ef7fddff3f338ba9c2a95b))

## [1.36.1](https://github.com/will-moss/isaiah/compare/1.36.0...1.36.1) (2025-07-13)


### Bug Fixes

* **agents:** added agents sort by name for easier navigation and selection ([f53a344](https://github.com/will-moss/isaiah/commit/f53a344b95f6a562d4a2e7fb08c3323c94b9defb)), closes [#24](https://github.com/will-moss/isaiah/issues/24)
* **overview:** added native scroll mechanism to the overview to accomodate for a list of many agents ([c5addb9](https://github.com/will-moss/isaiah/commit/c5addb969d5fb3c46b53d9576f6d3440972951c3)), closes [#23](https://github.com/will-moss/isaiah/issues/23)

# [1.36.0](https://github.com/will-moss/isaiah/compare/1.35.1...1.36.0) (2025-05-07)


### Bug Fixes

* **containers:** added a note about updating Isaiah container from Isaiah itself, to prevent errors ([57b0f02](https://github.com/will-moss/isaiah/commit/57b0f02149905d9375243fafee55ea1ee62f14a6))


### Features

* **containers:** made the containers bulk operations show the containers' name rather than their ID ([8e3656a](https://github.com/will-moss/isaiah/commit/8e3656a8fc6cbe250905a88f4e4cdd3688cdc91b))

## [1.35.1](https://github.com/will-moss/isaiah/compare/1.35.0...1.35.1) (2025-05-05)


### Bug Fixes

* **containers:** added a check to prevent Isaiah from stopping/restarting/updating itself in bulk op ([8f77097](https://github.com/will-moss/isaiah/commit/8f77097d7c442921894072f951bd11b09d71c7e8))
* **style:** added new CSS rules to try to prevent scrollbars from showing forcefully on all browsers ([bc04ad4](https://github.com/will-moss/isaiah/commit/bc04ad48bc315094be8124a93bc0a7bc64517a81))

# [1.35.0](https://github.com/will-moss/isaiah/compare/1.34.0...1.35.0) (2025-03-27)


### Features

* **containers:** added the ability to bulk restart and bulk update containers ([410fa62](https://github.com/will-moss/isaiah/commit/410fa62060928f412be1f01400b16bf56137972b)), closes [#22](https://github.com/will-moss/isaiah/issues/22)

# [1.34.0](https://github.com/will-moss/isaiah/compare/1.33.0...1.34.0) (2025-02-15)


### Features

* **images:** added ability to see whether an image was used or not, with associated containers ([e2fb66a](https://github.com/will-moss/isaiah/commit/e2fb66a72d0652f16cdcbd2e8932e89ce90ff1bb)), closes [#21](https://github.com/will-moss/isaiah/issues/21)

# [1.33.0](https://github.com/will-moss/isaiah/compare/1.32.1...1.33.0) (2024-11-29)


### Features

* **cli:** added -v / --version switch on the commandline, to retrieve the version of the binary ([5bc5935](https://github.com/will-moss/isaiah/commit/5bc5935320715032f39dbc74d61be8691ac2132e))

## [1.32.1](https://github.com/will-moss/isaiah/compare/1.32.0...1.32.1) (2024-10-21)


### Bug Fixes

* **server:** fixed the code responsible from browsing a volume inside a shell, to prevent hanging ([3b63a85](https://github.com/will-moss/isaiah/commit/3b63a852241cf0afd5688e57fad14a99996830ab))

# [1.32.0](https://github.com/will-moss/isaiah/compare/1.31.1...1.32.0) (2024-10-18)


### Features

* **server:** added the ability to browse volumes while inside a Docker container, using Busybox ([973d195](https://github.com/will-moss/isaiah/commit/973d19514ab3ce2311118bd51c36dc2ede7d4d66)), closes [#16](https://github.com/will-moss/isaiah/issues/16)

## [1.31.1](https://github.com/will-moss/isaiah/compare/1.31.0...1.31.1) (2024-10-18)


### Bug Fixes

* **server:** fixed container environment retrieval crashing the app when no value is provided ([824d1e9](https://github.com/will-moss/isaiah/commit/824d1e9f4ddf361078231a7fb917e799578a80cb)), closes [#18](https://github.com/will-moss/isaiah/issues/18)

# [1.31.0](https://github.com/will-moss/isaiah/compare/1.30.0...1.31.0) (2024-10-06)


### Bug Fixes

* **mobile:** improved styling on mobile when the resource's name is too large to fit the screen ([bf24d94](https://github.com/will-moss/isaiah/commit/bf24d945f84ffef490ed0783ebcd6ba179c8463b))


### Features

* **mobile:** replaced the scrolling menu and associated notification with a hamburger + popup menu ([b2814b9](https://github.com/will-moss/isaiah/commit/b2814b97b7aaba661aef8bcfb30726a1a954ce31)), closes [#17](https://github.com/will-moss/isaiah/issues/17)

# [1.30.0](https://github.com/will-moss/isaiah/compare/1.29.1...1.30.0) (2024-09-30)


### Bug Fixes

* **mobile:** added an initial popup on mobile to instruct users about the scrollable menu ([c510000](https://github.com/will-moss/isaiah/commit/c510000642a331c0c0e80f0f408cdfa2e840c66d)), closes [#17](https://github.com/will-moss/isaiah/issues/17)
* **mobile:** improved styling on mobile to prevent tab header buttons to be cut out ([c71f866](https://github.com/will-moss/isaiah/commit/c71f8662bd0af60c54014c84d70d333554d217d0))


### Features

* **images:** added fallback to RepoDigests when RepoTags is empty, to construct image's name ([80e8e86](https://github.com/will-moss/isaiah/commit/80e8e86c90376d40ca5aabba955c024cf54fdbe4)), closes [#15](https://github.com/will-moss/isaiah/issues/15)

## [1.29.1](https://github.com/will-moss/isaiah/compare/1.29.0...1.29.1) (2024-09-16)


### Bug Fixes

* **multi-node:** improved auto-login, agent-switching, and overview when in a multi-node deployment ([944e1d7](https://github.com/will-moss/isaiah/commit/944e1d791f9c18988a13f4db75551415f25bf6af))
* **multi-node:** removed port binding check when current node is an Agent ([3e5de37](https://github.com/will-moss/isaiah/commit/3e5de37534969facfdd3ce09f7350bae1cac3036))

# [1.29.0](https://github.com/will-moss/isaiah/compare/1.28.0...1.29.0) (2024-09-16)


### Features

* **stacks:** added new bulk operations for stacks ([5278982](https://github.com/will-moss/isaiah/commit/5278982363312fa1631f9758a454f851a4967ad3))

# [1.27.0](https://github.com/will-moss/isaiah/compare/1.26.0...1.27.0) (2024-09-05)


### Features

* **client:** in a multi-node environment, Agents are now shown in the first overview for better UX ([6c304af](https://github.com/will-moss/isaiah/commit/6c304af1cdcbde700254aecee7f94b0649b1b612)), closes [#12](https://github.com/will-moss/isaiah/issues/12)

# [1.26.0](https://github.com/will-moss/isaiah/compare/1.25.0...1.26.0) (2024-08-29)


### Bug Fixes

* **client:** now the stack icon is displayed correctly in Safari when showing the Overview panel ([c7844de](https://github.com/will-moss/isaiah/commit/c7844de8232cc0d61a25af3c0e800adb92614a8c))


### Features

* **stacks:** added support for stopping a stack (run "docker compose stop") ([dc98921](https://github.com/will-moss/isaiah/commit/dc9892153482f334bb66d1b8dfa8f80a21e90425))

# [1.25.0](https://github.com/will-moss/isaiah/compare/1.24.0...1.25.0) (2024-08-28)


### Features

* **client:** added support for native browser password autocompletion during authentication ([5a4af9f](https://github.com/will-moss/isaiah/commit/5a4af9fbdd586b93187229312ff62af9ef05d613))
* **multi-node:** added a reconnection strategy so agents auto-reconnect to master when disconnected ([f5395d1](https://github.com/will-moss/isaiah/commit/f5395d1c155157992264e3d0821049b6d2c1c86a))
* **multi-node:** added auto-login to agents from the client when master and agent share same secret ([3ef2cbd](https://github.com/will-moss/isaiah/commit/3ef2cbdb82a27ebc492d13f5ed7654fbadb1a1df))

# [1.24.0](https://github.com/will-moss/isaiah/compare/1.23.2...1.24.0) (2024-08-26)


### Bug Fixes

* **containers:** added a pop-up block to prevent trying to edit a container on a remote host ([552cb6e](https://github.com/will-moss/isaiah/commit/552cb6e7f45d3619d10c50f32149dc3078374176))
* **ui:** fixed a typo in the authentication process success notification ([683a9b9](https://github.com/will-moss/isaiah/commit/683a9b9bc3a34fb93a37bfc48c98ff1600d207e8))


### Features

* **project:** added full support for chunked communication (local + multi-node + multi-host) ([fb7261e](https://github.com/will-moss/isaiah/commit/fb7261e261c524c8d0ebc5d0595e34ab8f5d10d3)), closes [#9](https://github.com/will-moss/isaiah/issues/9)
* **stacks:** added support for managing Docker Compose stacks on remote hosts over Docker socket ([c40a468](https://github.com/will-moss/isaiah/commit/c40a4686f09aaf872dddfcf7d2c05edd9c6599b5))

## [1.23.2](https://github.com/will-moss/isaiah/compare/1.23.1...1.23.2) (2024-08-25)


### Bug Fixes

* **client:** numbers are now properly colored with syntax highlight in docker-compose.yml files ([872260e](https://github.com/will-moss/isaiah/commit/872260e391218e94453634178113e12e43f7667d))

## [1.23.1](https://github.com/will-moss/isaiah/compare/1.23.0...1.23.1) (2024-08-18)


### Bug Fixes

* **security:** added a check to allow only docker run commands in edit container feature ([1928ef5](https://github.com/will-moss/isaiah/commit/1928ef5c2e6f4c9167f12f5f1fe417f2641096bc))

# [1.23.0](https://github.com/will-moss/isaiah/compare/1.22.0...1.23.0) (2024-08-16)


### Features

* **project:** added full support for Docker stacks (docker-compose projects) ([4becb31](https://github.com/will-moss/isaiah/commit/4becb31a62fa8f3d048298c50c49dce263daa444))
* **project:** added support for editing containers (down, edit the docker run command, recreate) ([834d024](https://github.com/will-moss/isaiah/commit/834d0246f44352431e2a2010b65205b87dfeca79))

# [1.22.0](https://github.com/will-moss/isaiah/compare/1.21.1...1.22.0) (2024-08-09)


### Features

* **project:** added support for updating a container (down, pull, up) ([d42ca7b](https://github.com/will-moss/isaiah/commit/d42ca7bc780cca040878e7281030e0e5d3c339c6))

## [1.21.1](https://github.com/will-moss/isaiah/compare/1.21.0...1.21.1) (2024-07-20)


### Bug Fixes

* **shell:** disabled system shell feature when inside a Docker container, and added documentation ([af61d3b](https://github.com/will-moss/isaiah/commit/af61d3bf38dc5a1ce7686c57ed9c8d2e90f93132))

# [1.21.0](https://github.com/will-moss/isaiah/compare/1.20.1...1.21.0) (2024-07-16)


### Features

* **preferences:** added support for server-based preferences, rather than client-based ([b4c6aad](https://github.com/will-moss/isaiah/commit/b4c6aade5e822ff1e8d433f5866e9ab952373fc7)), closes [#4](https://github.com/will-moss/isaiah/issues/4)

## [1.20.1](https://github.com/will-moss/isaiah/compare/1.20.0...1.20.1) (2024-07-02)


### Bug Fixes

* **client:** added a check to prevent Javascript errors while the Inspector is still not loaded ([ba83ceb](https://github.com/will-moss/isaiah/commit/ba83ceb606cd2add630c9f1dbe199a0246fefc27))

# [1.20.0](https://github.com/will-moss/isaiah/compare/1.19.0...1.20.0) (2024-06-04)


### Bug Fixes

* **server:** the client won't hang anymore when trying to retrieve stats from a Created Container ([d89bd8a](https://github.com/will-moss/isaiah/commit/d89bd8a272104679db9e84381ceddba25896077f)), closes [#3](https://github.com/will-moss/isaiah/issues/3)


### Features

* **client:** extended the hover render-blocking mechanism to all of the inspector's tabs ([520118f](https://github.com/will-moss/isaiah/commit/520118f5ea864146474a295069e40c416dd596c9)), closes [#3](https://github.com/will-moss/isaiah/issues/3)

# [1.19.0](https://github.com/will-moss/isaiah/compare/1.18.0...1.19.0) (2024-05-25)


### Features

* **style:** increased the line-break's height in the version popup ([ba3e17b](https://github.com/will-moss/isaiah/commit/ba3e17b843c1afa412273262dbe47f0dcdc795e3))

# [1.18.0](https://github.com/will-moss/isaiah/compare/1.17.1...1.18.0) (2024-05-10)


### Features

* **client:** added a check against Github to display the latest version available when V is pressed ([c9d8e15](https://github.com/will-moss/isaiah/commit/c9d8e15df87e157c8783240d43d73e7073890674))

## [1.17.1](https://github.com/will-moss/isaiah/compare/1.17.0...1.17.1) (2024-05-10)


### Bug Fixes

* **client:** added a control to prevent interferences with logs' text selection and mouse clicks ([ede576e](https://github.com/will-moss/isaiah/commit/ede576ee36743f27926010cdcc549cabbc5353bc))

# [1.17.0](https://github.com/will-moss/isaiah/compare/1.16.0...1.17.0) (2024-05-10)


### Features

* **client:** you can now select and copy logs with your mouse, without losing selection on refresh ([1aa3c53](https://github.com/will-moss/isaiah/commit/1aa3c53089d72313925080a00927541ddbf57a6b))

# [1.16.0](https://github.com/will-moss/isaiah/compare/1.15.0...1.16.0) (2024-05-08)


### Features

* **client:** you can now pull the highlighted image without having to type anything ([16b9186](https://github.com/will-moss/isaiah/commit/16b918626af34d546f9e4301f644e9798e69c752))
* **jump:** added support for fuzzy-searching, with option to enable/disable it ([db4b2f7](https://github.com/will-moss/isaiah/commit/db4b2f70c2b068c108e86080885d6d04ebe923e7))
* **jump:** you can now cancel the jump action or confirm it without waiting for remote hosts search ([08b4bf9](https://github.com/will-moss/isaiah/commit/08b4bf9d55f71630cce4798671f519ce6a21822d))
* **server:** enabled native Goroutines to improve performance and enable anticipated cancels ([e20e02f](https://github.com/will-moss/isaiah/commit/e20e02f56cfac7376c8175f56fcea2ed8b13ef60))

# [1.15.0](https://github.com/will-moss/isaiah/compare/1.14.1...1.15.0) (2024-05-08)


### Features

* **client:** added a Version command to display the current version of Isaiah in the browser ([93725d1](https://github.com/will-moss/isaiah/commit/93725d1327fb595fb0c316ba4bdb270fb7c2dff0))

## [1.14.1](https://github.com/will-moss/isaiah/compare/1.14.0...1.14.1) (2024-05-07)


### Bug Fixes

* **client:** fixed the Jump feature by fixing the code responsible for sorting resources ([5b67363](https://github.com/will-moss/isaiah/commit/5b6736376925da5f48c6b909564c91901d37a331))

# [1.14.0](https://github.com/will-moss/isaiah/compare/1.13.0...1.14.0) (2024-05-06)


### Bug Fixes

* **jump:** rewrote part of the jump feature to prevent loss of key presses and improve search ([f28c33a](https://github.com/will-moss/isaiah/commit/f28c33a8c613ad4893df9a957c7c15305d9ab93c))


### Features

* **security:** added support for forward proxy authentication (e.g. with Authelia) ([6485734](https://github.com/will-moss/isaiah/commit/6485734aaded6de388a5c1482efe13bc99b3fce0))

# [1.12.0](https://github.com/will-moss/isaiah/compare/1.11.1...1.12.0) (2024-05-03)


### Bug Fixes

* **overview:** added mobile control for the overview feature ([ccdf2e5](https://github.com/will-moss/isaiah/commit/ccdf2e5d5446ed36963adf8c5e0e8c45b5914261))


### Features

* **client:** added an option to enable the user to choose between raw log lines and stripped ones ([b31102b](https://github.com/will-moss/isaiah/commit/b31102bdfd5c1b06d5b84678a213c8ad076ad08d))
* **client:** added jump feature (type the name of any resource, and quickly jump to it) ([92fea15](https://github.com/will-moss/isaiah/commit/92fea15cb7e5b86add01bebb83e3263016749352))

## [1.11.1](https://github.com/will-moss/isaiah/compare/1.11.0...1.11.1) (2024-04-04)


### Bug Fixes

* **authentication:** fixed broken authentication due to default authentication hash set ([f784f79](https://github.com/will-moss/isaiah/commit/f784f79d3e5d7bc7823ea6f436976e6b4c2ccbab))

# [1.11.0](https://github.com/will-moss/isaiah/compare/1.10.0...1.11.0) (2024-04-04)


### Bug Fixes

* **client:** fixed a bug causing the shell to be non-exitable on mobile due to too-strict controls ([ec7b8e3](https://github.com/will-moss/isaiah/commit/ec7b8e32a5f4c426e48800ef1c17e23635cec169))
* **client:** fixed a case when the menu tab's header would show undefined (during a remove action) ([9386db3](https://github.com/will-moss/isaiah/commit/9386db371229687593fca34e8625b80dcb0350c4))
* **client:** fixed app rendering when absolutely no data was gathered from the Docker server daemon ([f2adfc7](https://github.com/will-moss/isaiah/commit/f2adfc71213b66359f82e50780a6b40ac6f4f6f3))
* **client:** improved mouse navigation while searching, and fixed an infinite loop scenario ([ce71f59](https://github.com/will-moss/isaiah/commit/ce71f59deab088f891dd9575328df5172b456172))
* **server:** fixed a bug causing Isaiah to believe a remote host is accessible, while Docker isn't ([69e8495](https://github.com/will-moss/isaiah/commit/69e8495936ee7fcb0870f868b7fc91c9a46b19de))


### Features

* **client:** added ability to click on the agent's name to trigger the agent picker ([99dc276](https://github.com/will-moss/isaiah/commit/99dc276815d32756b3d53291fe320e63dc3064e8))
* **client:** added ability to pick a host by clicking on the host's name in the lower right corner ([f6bd54c](https://github.com/will-moss/isaiah/commit/f6bd54c7c835e951506902bdaf5aa5c071a475b3))
* **project:** added support for an Overview panel showing the server, hosts, and agents all at once ([5e831e5](https://github.com/will-moss/isaiah/commit/5e831e5ad2dba3880d204c99320cc519acf2fd4f))

# [1.10.0](https://github.com/will-moss/isaiah/compare/1.9.0...1.10.0) (2024-03-30)


### Bug Fixes

* **client:** fixed mouse navigation when performing search on logs, with minor refactoring ([85707f4](https://github.com/will-moss/isaiah/commit/85707f419a37400d4d79d91523c73df5d9a58245))


### Features

* **containers:** added "Created At" to the list of supported fields for Containers ([28eb6c3](https://github.com/will-moss/isaiah/commit/28eb6c3e8a563a508f296ebc7ea572807e778de3))
* **project:** added full support for client-side sorting of all the Docker resource lists ([071758d](https://github.com/will-moss/isaiah/commit/071758d6c984fbc2c9af2dc725d18acce71551b3))

# [1.9.0](https://github.com/will-moss/isaiah/compare/1.8.0...1.9.0) (2024-03-30)


### Bug Fixes

* **client:** added check to prevent running logs' copy when logs are empty ([ccf9ff7](https://github.com/will-moss/isaiah/commit/ccf9ff76af3ee360c0db823a3684d2faac73a431))
* **client:** added missing automatic inspector refresh on search while typing ([d6a25f9](https://github.com/will-moss/isaiah/commit/d6a25f9031ea9012edcd8d0b63757626f5121afa))
* **client:** fixed mouse navigation between different tabs while performing search ([558aaf4](https://github.com/will-moss/isaiah/commit/558aaf45c2a2545246f87a9f81aaaedf421aa156))
* **client:** fixed mouse navigation when performing search ([93d706c](https://github.com/will-moss/isaiah/commit/93d706c645d4de6e72522f23eeafcc8180caeabd))


### Features

* **client:** added support for searching through log lines ([c5d3cf8](https://github.com/will-moss/isaiah/commit/c5d3cf823ee44a9f21f3ba317e96ca975b55ba13))

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
