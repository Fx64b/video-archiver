# [1.2.0](https://github.com/Fx64b/video-archiver/compare/0.1.1.0-BETA...0.1.2.0-BETA) (2025-04-19)


### Bug Fixes

* **deps:** attempt to fix pnpm-workspace.yaml ([b9d4a23](https://github.com/Fx64b/video-archiver/commit/b9d4a235854da0285bba6d2d6f8c2cd39a2e33b6))
* **docker:** update port mapping from 3001 to 3000 ([3c80f06](https://github.com/Fx64b/video-archiver/commit/3c80f0659cd8978dd54a2c0b7ef5c2ac3caa5e13))


### Features

* **backend/statistics:** Add statistics service ([2becb68](https://github.com/Fx64b/video-archiver/commit/2becb68df845d3075d19d725c99feb276b796c24))
* **backend/statistics:** enhance statistics with top videos and other storage metrics ([64540e7](https://github.com/Fx64b/video-archiver/commit/64540e7beca1853adf70a18ee6cc3102e21c8852))
* **scripts:** add `--backend-only` flag ([a6a627a](https://github.com/Fx64b/video-archiver/commit/a6a627a166b6490e253a8ba1bf914ac9fc661759))
* **web/statistics:** add DownloadsChart and StorageChart components for statistics visualization ([02f9df3](https://github.com/Fx64b/video-archiver/commit/02f9df352889fd8fec60de166a19e4bdc7c108a8))
* **web/statistics:** load and display basic download statistics ([b2b8c63](https://github.com/Fx64b/video-archiver/commit/b2b8c63e262ddba227cde7d767cc0d411d5cc43d))

# [1.1.0](https://github.com/Fx64b/video-archiver/compare/0.1.0.0-BETA...0.1.1.0-BETA) (2025-03-09)


### Bug Fixes

* **web/metadata:** isChannel not working propperly ([7f7f814](https://github.com/Fx64b/video-archiver/commit/7f7f81431305fd602c0d5d190b193772f7d41f7a))


### Features

* **web/websocket:** add automatic websocket reconnect ([c7edda6](https://github.com/Fx64b/video-archiver/commit/c7edda6c73b32a4326a5cbc2d0864295d5ab734e))

# 1.0.0 (2025-03-04)


### Bug Fixes

* add more files to gitignore and add known issues to readme ([23aeb6e](https://github.com/Fx64b/video-archiver/commit/23aeb6e388c6ca6c3c28d977224679e76eb1a6b5))
* **config:** add missing comma ([8e6a91e](https://github.com/Fx64b/video-archiver/commit/8e6a91e273190905d63f76431115121fa452660f))
* **config:** invalid tag format ([b0f0549](https://github.com/Fx64b/video-archiver/commit/b0f05491b298df671ca02599ef25a28d9964e108))
* **config:** missed comma error ([4cc5b45](https://github.com/Fx64b/video-archiver/commit/4cc5b453c93b5f4a48ee2641103a8c6877c30f92))
* **docker:** permission issued that only allowed root user or docker to delete data files ([01c67f8](https://github.com/Fx64b/video-archiver/commit/01c67f8e5970d9c9d0f1f39e52fd241b7e26326f))
* **env:** incorrect env variable for websocket ([5350f95](https://github.com/Fx64b/video-archiver/commit/5350f95926bed132230f57ba24eb3ea8c8476e7a))
* **env:** no websocket env variable ([fcc9e35](https://github.com/Fx64b/video-archiver/commit/fcc9e35fbaaef07ea5be3495ac7bfdbc682dbc5f))
* **handlers:** cors issues ([0f90af5](https://github.com/Fx64b/video-archiver/commit/0f90af5d8a09f026bd2ccb34f5433ce431c6969a))
* **jobs:** minor fixes ([1bbce34](https://github.com/Fx64b/video-archiver/commit/1bbce34b7c76d2b25f417b368927fca1853c07b9))
* **metadat:** incorrect metadata path ([83f3a50](https://github.com/Fx64b/video-archiver/commit/83f3a50cccc84012c64bb9d3f11a9cc3cb1bb0b5))
* minor improvements ([724109f](https://github.com/Fx64b/video-archiver/commit/724109f31bc09b093b984cb9d674c69e1a3b0a59))
* **progress:** tracking of playlist/channels and videos not working as expected ([353d726](https://github.com/Fx64b/video-archiver/commit/353d72612cab82ba38ae73bf6e01e47fabb5f98a))
* **progress:** video already downloaded message not working properly when downloading a playlist twice ([6bfc878](https://github.com/Fx64b/video-archiver/commit/6bfc8784f198841aba3dce2e21cdd68679330ea7))
* re-add npm dependency but set package.json to private ([e1b8e22](https://github.com/Fx64b/video-archiver/commit/e1b8e226e4a623ae79da3524b7f0d35e0ab59e28))
* remove semantic-release npm ([0acab8a](https://github.com/Fx64b/video-archiver/commit/0acab8ae4827a007ad4af1d42c906f0d6df03aea))
* semantic-release error ([4761dfb](https://github.com/Fx64b/video-archiver/commit/4761dfb9f9dbddd387059f72724aa0d1da9ca702))
* **theme-provider:** type import causing error ([061a180](https://github.com/Fx64b/video-archiver/commit/061a1808c5a78f237dcd682bebec582c7c7c5c9a))
* upgrade next from 13.5.4 to 13.5.5 ([23829e4](https://github.com/Fx64b/video-archiver/commit/23829e40f63f0896ff6e3aff37c73f1e9021d17b))
* upgrade next from 13.5.5 to 13.5.6 ([9d980ee](https://github.com/Fx64b/video-archiver/commit/9d980ee4b8954ba966b7ede054fe759f8a1f1448))
* **url-input:** small spacing issue ([1c6c5f9](https://github.com/Fx64b/video-archiver/commit/1c6c5f983e27070129564ff527e3ab2b199d18de))
* **workflows:** actions improvement ([41e0515](https://github.com/Fx64b/video-archiver/commit/41e0515a23570c5b7a88d8452b0fc406813e2f05))
* **workflows:** pnpm not installed ([7277e01](https://github.com/Fx64b/video-archiver/commit/7277e0140314b7fe494a588dd30b325e9881419f))
* **workflows:** prevent action from being executed twice ([aefda55](https://github.com/Fx64b/video-archiver/commit/aefda55e250385c7c97f35cc0357ea8746ba563d))
* **workflows:** remove cd ([ef06cf0](https://github.com/Fx64b/video-archiver/commit/ef06cf026a9921db3dbf7c6d620779f885443ab7))


### Features

* add basic metadata service ([6fae963](https://github.com/Fx64b/video-archiver/commit/6fae9636a17112cbdf477f225b2384e5a2b76b7c))
* add run.sh script ([de56f19](https://github.com/Fx64b/video-archiver/commit/de56f1962acc248cb3b3d7af9e3a960148f1348b))
* add unversioned files ([8ac4eb8](https://github.com/Fx64b/video-archiver/commit/8ac4eb8e296b993bbd5e00f191959384108e42d6))
* add websocket channel for download status and display it in the frontend ([6a7a2b0](https://github.com/Fx64b/video-archiver/commit/6a7a2b0d32a5698178add9d1617ed443e3fb0f1b))
* **api/poc:** example that executes lscpu and sends it to the frontend ([dc5e911](https://github.com/Fx64b/video-archiver/commit/dc5e91138d52f74ee2c8adeaf97b7bcdf67fa78d))
* **app:** add basic examples ([16cf19c](https://github.com/Fx64b/video-archiver/commit/16cf19ccd544b97bdc0520a82329d51dee7c0fb5))
* **app:** switch from app router to normal pages ([fea971e](https://github.com/Fx64b/video-archiver/commit/fea971e25e10476a46ab6036a6c6379f0b7288f8))
* **backend/metadata:** improve channel metadata handling and send recent downloads with metadata ([f627045](https://github.com/Fx64b/video-archiver/commit/f627045b85c416ac937cf631b4a27c6ae1fc7816))
* **backend/metadata:** several improvements for metadata handling ([83adde1](https://github.com/Fx64b/video-archiver/commit/83adde1694fcaa64805c8987fbe189359751d355))
* **backend/recent:** add endpoint for recent jobs and do refactoring ([cd6f653](https://github.com/Fx64b/video-archiver/commit/cd6f6538db875b858cfe59c979552f3498cfbd08))
* **backend:** add automatically generated typescript types based on go code ([b537a11](https://github.com/Fx64b/video-archiver/commit/b537a1113e692aa2ba00dd876172f09691308fca))
* basic frontend setup with shadcn sidebar ([791e4fc](https://github.com/Fx64b/video-archiver/commit/791e4fcf7d8775260d2f24297b13c39c6c10d8e1))
* copy stuff from fx64b.dev to hopefully get semantic-release working properly ([50bbcb3](https://github.com/Fx64b/video-archiver/commit/50bbcb33730186420ad4111f0eace6d17fe766c6))
* improve queue and stream logs ([9ae8f6f](https://github.com/Fx64b/video-archiver/commit/9ae8f6f21e0b585f313dd5344d707a984eaeaff3))
* **metadata:** basic metadata display for POC ([3915108](https://github.com/Fx64b/video-archiver/commit/39151089706bb2094d3236a18e5b45514766265f))
* **metadata:** improve metadata handling, basic metadata display in frontend (WIP) [BROKEN] ([75d6aea](https://github.com/Fx64b/video-archiver/commit/75d6aeadea7b389acd024dfd891581f4f567c055))
* **progress:** add jobType ([8ccac29](https://github.com/Fx64b/video-archiver/commit/8ccac293a08f750bb0744f2951bcc6cb3cd7e914))
* **progress:** metadata download detection general improvements, performance and bug fixes ([90c71ab](https://github.com/Fx64b/video-archiver/commit/90c71ab79657dd1bf3f0273ee3e081730a9da2e5))
* **progress:** several progress tracking features ([5b573ea](https://github.com/Fx64b/video-archiver/commit/5b573ea6b5181c92c6c5687c7a87f039d406bc0b))
* **README:** add categorization checkbox ([ac6a2b2](https://github.com/Fx64b/video-archiver/commit/ac6a2b2ac40d38d8c1ce6c0e8ffbe28e0228223f))
* **README:** add command (to trigger workflow) ([8f85172](https://github.com/Fx64b/video-archiver/commit/8f85172134dbdd7fe4078154096305d9f88a6148))
* **README:** add currently working on notice ([4324211](https://github.com/Fx64b/video-archiver/commit/4324211d623f158c7b856aa061d8b28eddefb424))
* **README:** trigger release workflow ([b73583b](https://github.com/Fx64b/video-archiver/commit/b73583bfe8fca9b17f4df0f4026dd9f56c429bf8))
* **README:** update project description ([159f609](https://github.com/Fx64b/video-archiver/commit/159f609cadc8c1912dddb97a93093d74e877c2da))
* remove everything and start anew ([7728d38](https://github.com/Fx64b/video-archiver/commit/7728d38d93f66d1a68019ad94b419e3eb5461f76))
* several fixes regarding permissions and database ([3b18845](https://github.com/Fx64b/video-archiver/commit/3b1884592fa3e1dc1babeab0f9c042c632493fe5))
* updates ([f8abbf6](https://github.com/Fx64b/video-archiver/commit/f8abbf6bf424ddd65eaff7f6e7e7d19064fb282d))
* **web/dashboard:** add placeholder pages ([56a080c](https://github.com/Fx64b/video-archiver/commit/56a080c2390301c6ca8b60c88f95dd96a15b1f68))
* **web/job-progress:** improve job progress card ([271f979](https://github.com/Fx64b/video-archiver/commit/271f9791af17f73d6d34bc260927782d831be966))
* **web/metadata:** add metadata display for recent jobs ([f50f1d6](https://github.com/Fx64b/video-archiver/commit/f50f1d645474af1b7ec5b74de209420e6020344a))
* **web/metadata:** improve metadata display and state ([3bb5d56](https://github.com/Fx64b/video-archiver/commit/3bb5d562b9b4efa0b6c4fa8d1a4f5a9985ba05fd))
* **web/metadata:** improve metadata display for job-progress cards ([ef0b3e1](https://github.com/Fx64b/video-archiver/commit/ef0b3e127d0df220b6e8d76eb4d0d345934e1f93))
* **web/recent:** display recently downloaded jobs ([da435f0](https://github.com/Fx64b/video-archiver/commit/da435f047a7e59f4e91fa2b58d93ce030c237ff5))
* **web/url-input:** improve url input state management and add settings button ([61de520](https://github.com/Fx64b/video-archiver/commit/61de520987b1a9f10743d05b77b11393c4c021e0))
* **web:** minor styling fixes ([b962749](https://github.com/Fx64b/video-archiver/commit/b962749167c4073983f335eb6c530f2432a0af70))
* **workflows:** add step to update version in web/package.json after release is done ([a8a7b8e](https://github.com/Fx64b/video-archiver/commit/a8a7b8e4710c1809f9458759f58a1dc5dc046ec6))
