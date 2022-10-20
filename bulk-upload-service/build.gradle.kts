import org.jetbrains.kotlin.gradle.tasks.KotlinCompile

plugins {
	id("org.springframework.boot") version "2.7.4"
	id("io.spring.dependency-management") version "1.0.14.RELEASE"
	id("com.microsoft.azure.azurewebapp") version "1.6.0"
	kotlin("jvm") version "1.6.21"
	kotlin("plugin.spring") version "1.6.21"
}

group = "gov.cdc"
version = "0.0.1-SNAPSHOT"
java.sourceCompatibility = JavaVersion.VERSION_11

repositories {
	mavenCentral()
}

dependencies {
	implementation("org.springframework.boot:spring-boot-starter-web")
	implementation("org.jetbrains.kotlin:kotlin-reflect")
	implementation("org.jetbrains.kotlin:kotlin-stdlib-jdk8")
	implementation("com.azure:azure-storage-blob:12.7.0")
	implementation("com.google.code.gson:gson")
//	runtimeOnly("com.microsoft.azure:azure-keyvault-secrets-spring-boot-starter:2.3.5")
	implementation("com.azure:azure-security-keyvault-secrets:4.5.0")
	implementation("com.azure:azure-identity:1.6.1")
	testImplementation("org.springframework.boot:spring-boot-starter-test")
}

tasks.withType<KotlinCompile> {
	kotlinOptions {
		freeCompilerArgs = listOf("-Xjsr305=strict")
		jvmTarget = "11"
	}
}

tasks.withType<Test> {
	useJUnitPlatform()
}

azurewebapp {
	subscription = "01f9b1b1-a130-4031-ba25-71771007d3ca"//'OCIO-EDEDEV-C1'
	resourceGroup = "ocio-ede-dev-moderate-hl7-rg"
	appName = "as-bulk-upload"
	pricingTier = "P1v2"
	region = "eastus"
	setRuntime(closureOf<com.microsoft.azure.gradle.configuration.GradleRuntimeConfig> {
		os("Linux")
		webContainer("Java SE")
		javaVersion("Java 11")
	})
//	setAppSettings(closureOf<MutableMap<String, String>> {
//		put("key", "value")
//	})
	setAuth(closureOf<com.microsoft.azure.gradle.auth.GradleAuthConfig> {
		type = "azure_cli"
	})
}