import org.jetbrains.kotlin.gradle.tasks.KotlinCompile

plugins {
    kotlin("jvm") version "1.8.10"
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
    // Detect if suite param was passed in
    val hasEnv = project.hasProperty("env")
    val hasSuites = project.hasProperty("suites")

    useTestNG {
        if (hasEnv or hasSuites) {
            val env = project.properties["env"] ?: "dev"
            val allSuites = listOf("aims-celr-csv-$env.xml", "aims-celr-hl7-$env.xml")
            val suitesToRun: List<String> = project.properties["suites"]?.toString()?.split(',') ?: allSuites
            val fullyQualifiedSuites = suitesToRun.map { file("src/test/resources/$it") }
            println(suitesToRun)
            suiteXmlFiles = fullyQualifiedSuites
        }
    }
}


tasks.withType<KotlinCompile> {
    kotlinOptions.jvmTarget = "11"
}

application {
    mainClass.set("MainKt")
}