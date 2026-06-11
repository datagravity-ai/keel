package docker

import (
	"os"
	"testing"
)

func TestGetDigestDockerHub(t *testing.T) {
	if os.Getenv("DOCKER_INTEGRATION") == "" {
		t.Skip("skipping integration test; set DOCKER_INTEGRATION=1 to run")
	}

	client := New("https://index.docker.io", "", "")

	tags, err := client.Tags("karolisr/keel")
	if err != nil {
		t.Fatalf("failed to get tags, error: %s", err)
	}

	if len(tags) == 0 {
		t.Errorf("no tags?")
	}
}
