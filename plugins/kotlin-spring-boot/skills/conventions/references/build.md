# Build

Gradle Kotlin DSL is the default. Maven uses equivalent plugins (`kotlin-maven-plugin`, `spring-boot-maven-plugin`, `openapi-generator-maven-plugin`).

## Pin versions, never invent them

Read versions from `gradle/libs.versions.toml` (Gradle) or `pom.xml` `<dependencyManagement>` (Maven). For new Entur libraries, fetch artifact list and BOM coordinates from the source repo:

| Library | Repo |
|---|---|
| Cloud logging (structured logs, GCP severity, request/response, on-demand) | https://github.com/entur/cloud-logging |
| OIDC auth (resource server, JWT, scopes) | https://github.com/entur/oidc-auth-resource-server |
| Kafka starter (producers, consumers, Avro, schema registry) | https://github.com/entur/entur-kafka-spring-starter |

Other Entur libraries: search https://github.com/entur for `*-spring-boot-starter` or `*-spring-starter`.

## Base build.gradle.kts

```kotlin
plugins {
    alias(libs.plugins.kotlin.jvm)
    alias(libs.plugins.kotlin.spring)
    alias(libs.plugins.spring.boot)
    alias(libs.plugins.spring.dependency.mgmt)
}

group = "org.entur"

java {
    toolchain {
        languageVersion = JavaLanguageVersion.of(25)
    }
}

kotlin {
    compilerOptions {
        freeCompilerArgs.add("-Xjsr305=strict")
    }
}

tasks.withType<Test> {
    useJUnitPlatform()
}
```

## Plugin additions

### `api_approach=contract-first`

```kotlin
plugins {
    alias(libs.plugins.openapi.generator)
}

openApiGenerate {
    generatorName.set("kotlin-spring")
    inputSpec.set("$rootDir/specs/api.yaml")
    outputDir.set("$buildDir/generated")
    apiPackage.set("org.entur.myapp.api")
    modelPackage.set("org.entur.myapp.model")
    configOptions.set(mapOf(
        "interfaceOnly" to "true",
        "useSpringBoot3" to "true",
        "useTags" to "true",
    ))
}

sourceSets.main {
    kotlin.srcDir("$buildDir/generated/src/main/kotlin")
}

tasks.compileKotlin {
    dependsOn(tasks.openApiGenerate)
}
```

Add `"reactive" to "true"` to `configOptions` when also `spring_stack=webflux`.

### `database=jpa`

```kotlin
plugins {
    alias(libs.plugins.kotlin.jpa)   // generates no-arg constructors for @Entity
}
```

### Layered Boot JAR (all Spring Boot projects)

```kotlin
tasks {
    bootJar {
        layered {
            enabled = true
            application {
                intoLayer("spring-boot-loader") {
                    include("org/springframework/boot/loader/**")
                }
                intoLayer("application")
            }
            dependencies {
                intoLayer("internal-dependencies") { include("org.entur*:*:*") }
                intoLayer("dependencies")
            }
            layerOrder = listOf("dependencies", "internal-dependencies", "spring-boot-loader", "application")
        }
    }
}
```

## gradle/libs.versions.toml

Versions below are placeholders. Pin to the current stable release per library — never invent a version. Check the project's existing catalog first; for new libraries, check the source repo (Entur libs above; upstream on Maven Central or GitHub releases).

```toml
[versions]
kotlin                  = "<pin>"
spring-boot             = "<pin>"
spring-dependency-mgmt  = "<pin>"
exposed                 = "<pin>"
flyway                  = "<pin>"
entur-cloud-logging     = "<pin>"
kotest                  = "<pin>"
testcontainers          = "<pin>"
openapi-generator       = "<pin>"
spring-mockk            = "<pin>"
mockito-kotlin          = "<pin>"

[libraries]
exposed-java-time       = { group = "org.jetbrains.exposed", name = "exposed-java-time",           version.ref = "exposed" }
exposed-spring-boot     = { group = "org.jetbrains.exposed", name = "exposed-spring-boot-starter", version.ref = "exposed" }

flyway-core             = { group = "org.flywaydb",          name = "flyway-core",                version.ref = "flyway" }
flyway-postgres         = { group = "org.flywaydb",          name = "flyway-database-postgresql", version.ref = "flyway" }

# Entur cloud-logging — exact artifact names: see repo README
entur-logging-spring    = { group = "no.entur.logging.cloud", name = "<starter-name>",      version.ref = "entur-cloud-logging" }
entur-logging-test      = { group = "no.entur.logging.cloud", name = "<test-starter-name>", version.ref = "entur-cloud-logging" }

kotest-assertions-core  = { group = "io.kotest",          name = "kotest-assertions-core", version.ref = "kotest" }
spring-mockk            = { group = "com.ninja-squad",    name = "springmockk",            version.ref = "spring-mockk" }
mockito-kotlin          = { group = "org.mockito.kotlin", name = "mockito-kotlin",         version.ref = "mockito-kotlin" }

[bundles]
exposed             = ["exposed-java-time", "exposed-spring-boot"]
flyway              = ["flyway-core", "flyway-postgres"]
entur-cloud-logging = ["entur-logging-spring"]

[plugins]
kotlin-jvm              = { id = "org.jetbrains.kotlin.jvm",           version.ref = "kotlin" }
kotlin-spring           = { id = "org.jetbrains.kotlin.plugin.spring", version.ref = "kotlin" }
kotlin-jpa              = { id = "org.jetbrains.kotlin.plugin.jpa",    version.ref = "kotlin" }
openapi-generator       = { id = "org.openapi.generator",              version.ref = "openapi-generator" }
spring-boot             = { id = "org.springframework.boot",           version.ref = "spring-boot" }
spring-dependency-mgmt  = { id = "io.spring.dependency-management",    version.ref = "spring-dependency-mgmt" }
```

## Dependencies by configuration

### Always

```kotlin
dependencies {
    implementation("org.springframework.boot:spring-boot-starter-actuator")
    implementation("org.springframework.boot:spring-boot-starter-validation")
    implementation(platform("no.entur.logging.cloud:bom:${libs.versions.enturCloudLogging.get()}"))
    implementation(libs.enturLoggingSpring)        // exact starter from cloud-logging README
    implementation("io.micrometer:micrometer-registry-prometheus")

    testImplementation("org.springframework.boot:spring-boot-starter-test")
    testImplementation(platform("no.entur.logging.cloud:bom:${libs.versions.enturCloudLogging.get()}"))
    testImplementation(libs.enturLoggingTest)      // exact test starter from cloud-logging README
    testImplementation("org.testcontainers:junit-jupiter")
}
```

For OIDC auth, add the Entur resource-server starter and test starter from https://github.com/entur/oidc-auth-resource-server. For Kafka, use the Entur Kafka Spring starter from https://github.com/entur/entur-kafka-spring-starter. Pin both via the version catalog.

### `spring_stack=mvc`
```kotlin
implementation("org.springframework.boot:spring-boot-starter-web")
```

### `spring_stack=webflux`
```kotlin
implementation("org.springframework.boot:spring-boot-starter-webflux")
implementation("org.jetbrains.kotlinx:kotlinx-coroutines-reactor")
```

### `database=exposed`
```kotlin
implementation(libs.bundles.exposed)
implementation(libs.bundles.flyway)
runtimeOnly("org.postgresql:postgresql")
testImplementation("org.testcontainers:postgresql")
```

### `database=spring-data-jdbc`
```kotlin
implementation("org.springframework.boot:spring-boot-starter-data-jdbc")
implementation(libs.bundles.flyway)
runtimeOnly("org.postgresql:postgresql")
testImplementation("org.testcontainers:postgresql")
```

### `database=jpa`
```kotlin
implementation("org.springframework.boot:spring-boot-starter-data-jpa")
implementation(libs.bundles.flyway)
runtimeOnly("org.postgresql:postgresql")
testImplementation("org.testcontainers:postgresql")
```

### Test libraries

AssertJ and mockito-core come transitively via `spring-boot-starter-test`. Add only the entries matching the active config.

| Config | Dependency |
|---|---|
| `test_mocking=mockk` | `testImplementation(libs.springMockk)` |
| `test_mocking=mockito-kotlin` | `testImplementation(libs.mockitoKotlin)` |
| `test_assertions=kotest` | `testImplementation(libs.kotestAssertionsCore)` |
| `test_assertions=assertj` | (transitive) |

## Artifactory (JFrog)

Entur-internal release artifacts live in JFrog. Treat the URL as deployment config — read it from a project-specific `gradle.properties`, an env var, or an existing repo's build script. Don't hardcode internal URLs in skill output.

```kotlin
repositories {
    val artifactoryUrl: String? by project        // pin in gradle.properties or pass via -P
    val artifactoryUser: String? by project
    val artifactoryPassword: String? by project

    if (!artifactoryUrl.isNullOrBlank()) {
        maven {
            name = "Entur JFrog"
            url = URI(artifactoryUrl)
            credentials {
                username = artifactoryUser ?: System.getenv("ARTIFACTORY_AUTH_USER")
                password = artifactoryPassword ?: System.getenv("ARTIFACTORY_AUTH_TOKEN")
            }
        }
    }
}
```

Credentials: `~/.gradle/gradle.properties` locally, `ARTIFACTORY_AUTH_USER` / `ARTIFACTORY_AUTH_TOKEN` org secrets in CI. Ask the team or check an existing repo for the current Artifactory URL.
