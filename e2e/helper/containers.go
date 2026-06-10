/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package helper

import (
	"context"
	"io"
	"log"
	"testing"

	pluginpkg "github.com/mercedes-benz/gitflow-cli/core/plugin"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// SetupPluginContainer starts a Docker container for the given plugin TestConfig
// and registers it with the executor so that CLI commands use "docker exec" on the running container.
// The container mounts the test's local repository path at /work.
// The container is automatically terminated when the test completes.
// If the TestConfig has no DockerImage, this is a no-op.
func SetupPluginContainer(t *testing.T, tc pluginpkg.TestConfig, localRepoPath string) {
	t.Helper()

	if tc.DockerImage == "" {
		return
	}

	registrationName := tc.PluginName
	if registrationName == "" {
		registrationName = tc.Name
	}

	ctx := context.Background()

	log.Printf("[testcontainer] Starting container: plugin=%s image=%s mount=%s", registrationName, tc.DockerImage, localRepoPath)

	req := testcontainers.ContainerRequest{
		Image: tc.DockerImage,
		Cmd:   []string{"tail", "-f", "/dev/null"},
		Mounts: testcontainers.ContainerMounts{
			testcontainers.BindMount(localRepoPath, "/work"),
		},
		WaitingFor: wait.ForExec([]string{"true"}),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("[testcontainer] Failed to start container for %s: %v", tc.Name, err)
	}

	containerID := container.GetContainerID()
	log.Printf("[testcontainer] Container ready: plugin=%s containerID=%s", tc.Name, containerID[:12])

	for _, setupCmd := range tc.SetupCommands {
		log.Printf("[testcontainer] Running setup command: %v", setupCmd)
		code, reader, err := container.Exec(ctx, setupCmd)
		if err != nil {
			t.Fatalf("[testcontainer] Setup command failed for %s: %v", registrationName, err)
		}
		if code != 0 {
			output, _ := io.ReadAll(reader)
			t.Fatalf("[testcontainer] Setup command exited with code %d for %s: %s", code, registrationName, output)
		}
		log.Printf("[testcontainer] Setup command completed: %v", setupCmd)
	}

	pluginpkg.RegisterContainer(registrationName, containerID)

	t.Cleanup(func() {
		log.Printf("[testcontainer] Stopping container: plugin=%s containerID=%s", registrationName, containerID[:12])
		pluginpkg.UnregisterContainer(registrationName)
		if err := container.Terminate(ctx); err != nil {
			t.Logf("[testcontainer] Warning: failed to terminate container %s: %v", containerID[:12], err)
		}
		log.Printf("[testcontainer] Container removed: plugin=%s", registrationName)
	})
}
