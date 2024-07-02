plugins {
    kotlin("jvm") version "1.9.22"
}

group = "org.example"
version = "1.0-SNAPSHOT"

repositories {
    mavenCentral()
}

dependencies {
    implementation(platform("com.azure:azure-sdk-bom:1.2.10"))
    implementation("com.azure:azure-identity")
    implementation("com.azure:azure-storage-blob")
    implementation("io.tus.java.client:tus-java-client:0.5.0")
    implementation("com.github.doyaaaaaken:kotlin-csv-jvm:1.9.3")
    implementation("com.squareup.okhttp3:okhttp:4.12.0")
    implementation("com.fasterxml.jackson.core:jackson-core:2.9.9")
    implementation("com.fasterxml.jackson.core:jackson-annotations:2.9.9")
    implementation("com.fasterxml.jackson.core:jackson-databind:2.14.0-rc1")
    implementation("com.fasterxml.jackson.module:jackson-module-kotlin:2.9.9")
    implementation("org.jetbrains.kotlin:kotlin-test")
}

tasks.test {
    useJUnitPlatform()
}
kotlin {
    jvmToolchain(15)
}