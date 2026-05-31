plugins {
    id("com.google.protobuf") version "0.9.4"
    application
}

group = "io.github.phi42.ade"
version = "0.1.0"

repositories {
    mavenCentral()
}

dependencies {
    implementation("com.google.protobuf:protobuf-java:4.29.3")
}

application {
    mainClass.set("io.github.phi42.ade.plugin.Main")
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
        attributes["Main-Class"] = "io.github.phi42.ade.plugin.Main"
    }
    from(configurations.runtimeClasspath.get().map { if (it.isDirectory) it else zipTree(it) })
    with(tasks.jar.get())
    dependsOn(tasks.compileJava)
}

tasks.build {
    dependsOn(tasks["fatJar"])
}
