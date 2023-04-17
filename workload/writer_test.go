package workload

import (
	"os/exec"
	"testing"
)

func testRemovePath(t *testing.T, path string) {
	t.Helper()
	cmd := exec.Command("rm", "-rf", path)
	err := cmd.Run()
	if err != nil {
		t.Error("Failed executing command", cmd, err)
	}
}

func TestWriterDirectory(t *testing.T) {
	dir := "."
	err := checkDirectory(dir)
	if err != nil {
		t.Error("Expected to check dir", dir, err)
	}

	dir = "/non-existent"
	err = checkDirectory(dir)
	if err == nil {
		t.Error("Expected to not check dir", dir, err)
	}

	dir = "non-existent"
	defer testRemovePath(t, dir)

	err = checkDirectory(dir)
	if err == nil {
		t.Error("Expected to not check dir", dir, err)
	}
	err = createDirectory(dir)
	if err != nil {
		t.Error("Unexpected error creating dir", dir, err)
	}
	err = checkDirectory(dir)
	if err != nil {
		t.Error("Expected to check dir", dir, err)
	}
}
