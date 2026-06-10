/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package mvn

import "github.com/mercedes-benz/gitflow-cli/core/plugin"

var E2ETestConfig = plugin.TestConfig{
	Name:             "mvn",
	DockerImage:      dockerImage,
	VersionQualifier: "SNAPSHOT",
	VersionFileName:  "pom.xml",
	Template: `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 https://maven.apache.org/xsd/maven-4.0.0.xsd">
    <modelVersion>4.0.0</modelVersion>

    <groupId>com.mercedes-benz</groupId>
    <artifactId>dummy</artifactId>
    <version>{{.Version}}</version>

</project>
`,
	EmptyFileContent:   nil,
	HasBeforeStartHook: false,
}
