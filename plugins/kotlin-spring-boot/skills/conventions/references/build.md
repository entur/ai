# Build

Follow the `build_tool` axis strictly.

- `build_tool=gradle`: use only Gradle snippets.
- `build_tool=maven`: use only Maven snippets.

Never emit Gradle blocks in a Maven repo, and never emit Maven XML in a Gradle repo.

## Pin versions, never invent them

Read versions from repo sources:

- Gradle: `gradle/libs.versions.toml`
- Maven: `pom.xml` (`<properties>` and `<dependencyManagement>`)

For new Entur libraries, fetch artifact names and BOM coordinates from:

| Library | Repo |
|---|---|
| Cloud logging (structured logs, GCP severity, request/response, on-demand) | https://github.com/entur/cloud-logging |
| OIDC auth (resource server, JWT, scopes) | https://github.com/entur/oidc-auth-resource-server |
| Kafka starter (producers, consumers, Avro, schema registry) | https://github.com/entur/entur-kafka-spring-starter |

Other Entur libraries: search https://github.com/entur for `*-spring-boot-starter` or `*-spring-starter`.

## Base build configuration

### `build_tool=gradle`

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

### `build_tool=maven`

```xml
<properties>
  <java.version>25</java.version>
  <kotlin.version><!-- pin --></kotlin.version>
  <spring-boot.version><!-- pin --></spring-boot.version>
</properties>

<dependencyManagement>
  <dependencies>
    <dependency>
      <groupId>org.springframework.boot</groupId>
      <artifactId>spring-boot-dependencies</artifactId>
      <version>${spring-boot.version}</version>
      <type>pom</type>
      <scope>import</scope>
    </dependency>
  </dependencies>
</dependencyManagement>

<build>
  <plugins>
    <plugin>
      <groupId>org.jetbrains.kotlin</groupId>
      <artifactId>kotlin-maven-plugin</artifactId>
      <version>${kotlin.version}</version>
      <configuration>
        <args>
          <arg>-Xjsr305=strict</arg>
        </args>
        <compilerPlugins>
          <plugin>spring</plugin>
        </compilerPlugins>
      </configuration>
      <executions>
        <execution>
          <id>compile</id>
          <goals>
            <goal>compile</goal>
          </goals>
        </execution>
        <execution>
          <id>test-compile</id>
          <goals>
            <goal>test-compile</goal>
          </goals>
        </execution>
      </executions>
    </plugin>

    <plugin>
      <groupId>org.springframework.boot</groupId>
      <artifactId>spring-boot-maven-plugin</artifactId>
    </plugin>
  </plugins>
</build>
```

## Plugin additions

### `api_approach=contract-first`

#### Gradle

```kotlin
plugins {
    alias(libs.plugins.openapi.generator)
}

openApiGenerate {
    generatorName.set("kotlin-spring")
    inputSpec.set("$rootDir/specs/api.yaml")
    outputDir.set(layout.buildDirectory.dir("generated").get().asFile.absolutePath)
    apiPackage.set("org.entur.myapp.api")
    modelPackage.set("org.entur.myapp.model")
    configOptions.set(mapOf(
        "interfaceOnly" to "true",
        "useSpringBoot3" to "true",
        "useTags" to "true",
    ))
}

sourceSets.main {
    kotlin.srcDir(layout.buildDirectory.dir("generated/src/main/kotlin"))
}

tasks.compileKotlin {
    dependsOn(tasks.openApiGenerate)
}
```

#### Maven

```xml
<plugin>
  <groupId>org.openapitools</groupId>
  <artifactId>openapi-generator-maven-plugin</artifactId>
  <version><!-- pin --></version>
  <executions>
    <execution>
      <id>generate-openapi</id>
      <phase>generate-sources</phase>
      <goals>
        <goal>generate</goal>
      </goals>
      <configuration>
        <generatorName>kotlin-spring</generatorName>
        <inputSpec>${project.basedir}/specs/api.yaml</inputSpec>
        <output>${project.build.directory}/generated</output>
        <apiPackage>org.entur.myapp.api</apiPackage>
        <modelPackage>org.entur.myapp.model</modelPackage>
        <configOptions>
          <interfaceOnly>true</interfaceOnly>
          <useSpringBoot3>true</useSpringBoot3>
          <useTags>true</useTags>
        </configOptions>
      </configuration>
    </execution>
  </executions>
</plugin>
```

Set reactive mode when `spring_stack=webflux`:

- Gradle: add `"reactive" to "true"` in `configOptions`.
- Maven: add `<reactive>true</reactive>` in `<configOptions>`.

### `database=jpa`

#### Gradle

```kotlin
plugins {
    alias(libs.plugins.kotlin.jpa)
}
```

#### Maven

Enable the Kotlin JPA compiler plugin in `kotlin-maven-plugin`:

```xml
<configuration>
  <compilerPlugins>
    <plugin>spring</plugin>
    <plugin>jpa</plugin>
  </compilerPlugins>
</configuration>
```

### Layered Boot JAR (all Spring Boot projects)

#### Gradle

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

#### Maven

```xml
<plugin>
  <groupId>org.springframework.boot</groupId>
  <artifactId>spring-boot-maven-plugin</artifactId>
  <configuration>
    <layers>
      <enabled>true</enabled>
    </layers>
  </configuration>
</plugin>
```

Maven uses Spring Boot's default layer order. For custom ordering matching the Gradle setup above, define a `layers.xml` and reference it via `<configuration>` — see the Spring Boot Maven plugin docs.

## Version pinning layouts

### `build_tool=gradle`: `gradle/libs.versions.toml`

Versions below are placeholders. Pin to real released versions from repo state or upstream sources.

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
```

### `build_tool=maven`: `<properties>` + `<dependencyManagement>`

```xml
<properties>
  <kotlin.version><!-- pin --></kotlin.version>
  <spring-boot.version><!-- pin --></spring-boot.version>
  <entur-cloud-logging.version><!-- pin --></entur-cloud-logging.version>
  <openapi-generator.version><!-- pin --></openapi-generator.version>
  <flyway.version><!-- pin --></flyway.version>
</properties>

<dependencyManagement>
  <dependencies>
    <dependency>
      <groupId>no.entur.logging.cloud</groupId>
      <artifactId>bom</artifactId>
      <version>${entur-cloud-logging.version}</version>
      <type>pom</type>
      <scope>import</scope>
    </dependency>
  </dependencies>
</dependencyManagement>
```

## Dependencies by configuration

Use dependencies that match the active axes. AssertJ and Mockito core are transitively available from `spring-boot-starter-test`.

### Always

- `org.springframework.boot:spring-boot-starter-actuator`
- `org.springframework.boot:spring-boot-starter-validation`
- `io.micrometer:micrometer-registry-prometheus`
- Entur cloud-logging BOM + matching runtime/test starter artifacts
- `org.springframework.boot:spring-boot-starter-test`
- `org.testcontainers:junit-jupiter`

### `spring_stack=mvc`

- `org.springframework.boot:spring-boot-starter-web`

### `spring_stack=webflux`

- `org.springframework.boot:spring-boot-starter-webflux`
- `org.jetbrains.kotlinx:kotlinx-coroutines-reactor`

### `database=exposed`

- `org.jetbrains.exposed:exposed-java-time`
- `org.jetbrains.exposed:exposed-spring-boot-starter`
- `org.flywaydb:flyway-core`
- `org.flywaydb:flyway-database-postgresql`
- `org.postgresql:postgresql` (runtime)
- `org.testcontainers:postgresql` (test)

### `database=spring-data-jdbc`

- `org.springframework.boot:spring-boot-starter-data-jdbc`
- `org.flywaydb:flyway-core`
- `org.flywaydb:flyway-database-postgresql`
- `org.postgresql:postgresql` (runtime)
- `org.testcontainers:postgresql` (test)

### `database=jpa`

- `org.springframework.boot:spring-boot-starter-data-jpa`
- `org.flywaydb:flyway-core`
- `org.flywaydb:flyway-database-postgresql`
- `org.postgresql:postgresql` (runtime)
- `org.testcontainers:postgresql` (test)

### Test libraries

| Config | Dependency |
|---|---|
| `test_mocking=mockk` | `com.ninja-squad:springmockk` |
| `test_mocking=mockito-kotlin` | `org.mockito.kotlin:mockito-kotlin` |
| `test_assertions=kotest` | `io.kotest:kotest-assertions-core` |
| `test_assertions=assertj` | (transitive from `spring-boot-starter-test`) |

For OIDC auth and Kafka, add the matching Entur starters from their source repos and pin versions via catalog/properties.

## Artifactory (JFrog)

Entur-internal release artifacts live in JFrog. Treat the URL as environment config; do not hardcode internal URLs in generated output.

### `build_tool=gradle`

Read URL/credentials from `gradle.properties` and/or environment variables:

- local: `~/.gradle/gradle.properties`
- CI: `ARTIFACTORY_AUTH_USER`, `ARTIFACTORY_AUTH_TOKEN`

### `build_tool=maven`

Read repository URL from project POM/profiles, and credentials from `~/.m2/settings.xml` (local) or CI secrets.
