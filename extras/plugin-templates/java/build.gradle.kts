plugins {
    id("com.google.protobuf") version "0.9.4"
    application
}

// TODO: change group to your organisation's reverse domain (e.g. "com.example").
group = "com.example"
version = "0.1.0"

repositories {
    mavenCentral()
}

dependencies {
    implementation("com.google.protobuf:protobuf-java:4.29.3")
}

application {
    // TODO: update this to match your package after renaming.
    mainClass.set("com.example.plugin.Main")
}

protobuf {
    protoc {
        artifact = "com.google.protobuf:protoc:4.29.3"
    }
}

// Fat JAR so the plugin runs as a single file:
//   java -jar build/libs/plugin.jar --info
tasks.register<Jar>("fatJar") {
    archiveFileName.set("plugin.jar")
    duplicatesStrategy = DuplicatesStrategy.EXCLUDE
    manifest {
        // TODO: update this to match your package after renaming.
        attributes["Main-Class"] = "com.example.plugin.Main"
    }
    from(configurations.runtimeClasspath.get().map { if (it.isDirectory) it else zipTree(it) })
    with(tasks.jar.get())
    dependsOn(tasks.compileJava)
}

tasks.build {
    dependsOn(tasks["fatJar"])
}
