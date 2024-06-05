import org.gradle.api.tasks.testing.logging.TestExceptionFormat
import org.jetbrains.kotlin.gradle.tasks.KotlinCompile

plugins {
    kotlin("jvm") version "1.8.10"
    id("com.adarshr.test-logger") version "4.0.0"
    application
}

kotlin {
    jvmToolchain(11)
}

group = "me.cfarmer"
version = "1.0-SNAPSHOT"

repositories {
    mavenCentral()
}

dependencies {
    testImplementation(kotlin("test"))
    testImplementation(platform("com.azure:azure-sdk-bom:1.2.10"))
    testImplementation("com.azure:azure-identity")
    testImplementation("com.azure:azure-storage-blob")
    testImplementation("org.testng:testng:7.7.0")
    testImplementation("io.tus.java.client:tus-java-client:0.5.0")
    testImplementation("com.squareup.okhttp3:okhttp:4.12.0")
    testImplementation("com.fasterxml.jackson.core:jackson-core:2.9.9")
    testImplementation("com.fasterxml.jackson.core:jackson-annotations:2.9.9")
    testImplementation("com.fasterxml.jackson.core:jackson-databind:2.14.0-rc1")
    testImplementation("com.fasterxml.jackson.module:jackson-module-kotlin:2.9.9")
    testImplementation("io.rest-assured:rest-assured:5.4.0")
    testImplementation("joda-time:joda-time:2.12.7")
}

tasks.test {
    testLogging.showStandardStreams = true
    testLogging.exceptionFormat = TestExceptionFormat.FULL

    // Detect if suite params were passed in
//    val hasEnv = project.hasProperty("env")
//    val hasSuites = project.hasProperty("useCases")
    useTestNG {
        // This is needed otherwise System.getProperty returns null for the custom properties.
        if (project.hasProperty("useCases")) {
            systemProperties["useCases"] = project.properties["useCases"]
        }
    }
//    useTestNG {
    // If true, we want to test with XML suites.  Otherwise, test directly with Gradle and rely on default parameters.
//        if (hasEnv or hasSuites) {
//            val env = project.properties["env"] ?: "dev" // Default to dev.
//            val allUseCases = File("src/test/resources/$env").listFiles().map { it.nameWithoutExtension } // Collect all use cases from the env-specific suite directory.
//            val useCasesToRun: List<String> = project.properties["useCases"]?.toString()?.split(',') ?: allUseCases // If a set of use cases were passed in, use them.  Otherwise, default to running all.
//            val fullyQualifiedSuites = useCasesToRun.map { file("src/test/resources/$env/$it.xml") }
//            println("Running tests for use cases: $useCasesToRun")
//            suiteXmlFiles = fullyQualifiedSuites
//        }
//    }
}

tasks.withType<KotlinCompile> {
    kotlinOptions.jvmTarget = "11"
}

application {
    mainClass.set("MainKt")
}