# Changelog

## [0.1.2](https://github.com/entur/ai/compare/v0.1.1...v0.1.2) (2026-06-19)


### Documentation

* **guides:** align reference guidance with golden path ([#66](https://github.com/entur/ai/issues/66)) ([f937eee](https://github.com/entur/ai/commit/f937eee5523da8a97dd917975429059841fc7827))
* standardise guide metadata and fix CI/CD references ([#68](https://github.com/entur/ai/issues/68)) ([30fc89d](https://github.com/entur/ai/commit/30fc89df1d908856e8324f2e9fcc9173b663ce21))

## [0.1.1](https://github.com/entur/ai/compare/v0.1.0...v0.1.1) (2026-06-10)

### Bug Fixes

* add alerting reference ([#62](https://github.com/entur/ai/issues/62)) ([2b66982](https://github.com/entur/ai/commit/2b6698252614b6481a0ec9b6fcf99dcae171fff3))
* add incident response playbook ([#61](https://github.com/entur/ai/issues/61)) ([968a528](https://github.com/entur/ai/commit/968a528228d7666ac58e70a4b3443580b9c4c8a9))
* update PSQL cred's ([#63](https://github.com/entur/ai/issues/63)) ([e4eb933](https://github.com/entur/ai/commit/e4eb933bb285237d9793e0c5111ed3667036666f))

## [0.1.0](https://github.com/entur/ai/compare/v0.0.2...v0.1.0) (2026-06-05)

### Features

* Cloud Run - 'bronze' path ([#52](https://github.com/entur/ai/issues/52)) ([9c9df81](https://github.com/entur/ai/commit/9c9df81debabeecbbcf35ef96e1860ca3eb0f241))

### Documentation

* **cicd:** require per-job permissions on terraform workflow calls ([#50](https://github.com/entur/ai/issues/50)) ([d659076](https://github.com/entur/ai/commit/d659076caae501fe690be4eff3de952ba2450736))
* **iam-roles:** document source of truth and sync procedure ([#56](https://github.com/entur/ai/issues/56)) ([6e3f0ad](https://github.com/entur/ai/commit/6e3f0ad7b97b5866342c5d726576ff92ab3e88ae))
* **iam-roles:** mirror six new roles from upstream allowlist ([#58](https://github.com/entur/ai/issues/58)) ([3363358](https://github.com/entur/ai/commit/3363358ada4527e72883da65ed012ab24b185e2e))
* move contributing guidance to CONTRIBUTING.md ([#48](https://github.com/entur/ai/issues/48)) ([bac629f](https://github.com/entur/ai/commit/bac629f4777712031cce8b9fedc94d1a16a52db4))
* **profiler:** add profiler.md playbook and extract from observability ([#57](https://github.com/entur/ai/issues/57)) ([257b259](https://github.com/entur/ai/commit/257b25920c617ff91857061665b8bed1de92972c))
* Revise review process and documentation testing notes ([#49](https://github.com/entur/ai/issues/49)) ([fde31a5](https://github.com/entur/ai/commit/fde31a50c8b1e6cf557762a38604348fc0f98a05))
* **tracing:** add tracing.md playbook and extract from observability/logging ([#55](https://github.com/entur/ai/issues/55)) ([0f676a4](https://github.com/entur/ai/commit/0f676a42e89c7ca77fe8f72f26fa433daef49642))
* **tracing:** fix Go exporter self-loop and Cloud Run sampler advice ([#59](https://github.com/entur/ai/issues/59)) ([21f1335](https://github.com/entur/ai/commit/21f133546ee289ce772c009121ea3c4c4c425522))

## [0.0.2](https://github.com/entur/ai/compare/v0.0.1...v0.0.2) (2026-05-26)

### Bug Fixes

* switch release-please to node type so package.json is bumped ([#44](https://github.com/entur/ai/issues/44)) ([e6aad5c](https://github.com/entur/ai/commit/e6aad5cd66cec24a35d97ec593fad91d9b99f8d4))

## 0.0.1 (2026-05-26)

### Features

* 1st initial commit ([d670a3a](https://github.com/entur/ai/commit/d670a3a00ff5f614e5fb1d4030d66d04bf1199ba))
* add GCP project naming docs, comprehension tests, and bootstrap… ([#6](https://github.com/entur/ai/issues/6)) ([39f9dde](https://github.com/entur/ai/commit/39f9dde5820d7707a4c3c974ec2df3e498d2523a))
* **plugins:** clean up plugin naming and port remaining 2 skills ([#19](https://github.com/entur/ai/issues/19)) ([f97a6ae](https://github.com/entur/ai/commit/f97a6aed784de254056561fe0498b2e2d1870100))
* **plugins:** Support existing guides through plugin ([#22](https://github.com/entur/ai/issues/22)) ([80d27b7](https://github.com/entur/ai/commit/80d27b76c4ff832e5c321b1524b65b23aaa0d711))
* PoC of plugin marketplace support ([#18](https://github.com/entur/ai/issues/18)) ([d128243](https://github.com/entur/ai/commit/d128243d8016c0c5f8bb426f4257bffe9532c296))
* **tests:** add MCP knowledge-base quality test suite ([#39](https://github.com/entur/ai/issues/39)) ([f4c8a9c](https://github.com/entur/ai/commit/f4c8a9c79498995d4fe1dc2f21e58578ffe5522a))

### Bug Fixes

* **docs:** correct and extend cicd guides ([#28](https://github.com/entur/ai/issues/28)) ([a79cc74](https://github.com/entur/ai/commit/a79cc74b35d8b6a230943ec4f29992275f0c49e2))
* Pink Elephant Theory - improve AI language to use 'Always' ([#12](https://github.com/entur/ai/issues/12)) ([f299a78](https://github.com/entur/ai/commit/f299a78abfd0554f94f2e506c39d5eb7d7b82ebe))
* remove PG_USER/PG_PASSWORD from ExternalSecrets examples ([#10](https://github.com/entur/ai/issues/10)) ([a940209](https://github.com/entur/ai/commit/a940209317b73e58909f7720610e56e643c5ce77))
* specify use entur self-service ([#5](https://github.com/entur/ai/issues/5)) ([30da007](https://github.com/entur/ai/commit/30da00728bbf9609b0689c49ee599fee1c8b8df3))
* **tests:** route agents to setup-cicd-workflows skill + cleanup ([#38](https://github.com/entur/ai/issues/38)) ([7dbc558](https://github.com/entur/ai/commit/7dbc558fbe8b8b2788d0311cd4a60aec18e220e5))
* **tests:** stabilize flaky and failing scenario tests ([#30](https://github.com/entur/ai/issues/30)) ([cd33a0a](https://github.com/entur/ai/commit/cd33a0a1b5ed74fe02012d2ebeac3a272b8df800))
* url for developer guide ([#37](https://github.com/entur/ai/issues/37)) ([836f3a6](https://github.com/entur/ai/commit/836f3a62112ce08ca84d2f1c2e0a27f6a49ceea8))
