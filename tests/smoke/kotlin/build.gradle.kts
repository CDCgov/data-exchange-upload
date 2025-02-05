import org.gradle.api.tasks.testing.logging.TestExceptionFormat
import org.jetbrains.kotlin.gradle.tasks.KotlinCompile

plugins {
    kotlin("jvm") version "1.8.10"
    id("com.adarshr.test-logger") version "4.0.0"
    application
}

kotlin {
    jvmToolchain(17)

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
    testImplementation("org.slf4j:slf4j-simple:1.6.1")
}

tasks.test {
    testLogging.showStandardStreams = true
    testLogging.exceptionFormat = TestExceptionFormat.FULL

    useTestNG {
        // This is needed otherwise System.getProperty returns null for the custom properties.
        if (project.hasProperty("manifestFilter")) {
            systemProperties["manifestFilter"] = project.properties["manifestFilter"]
        }
    }
}

tasks.withType<KotlinCompile> {
    kotlinOptions.jvmTarget = "17"
}

application {
    mainClass.set("MainKt")
}