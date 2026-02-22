//go:build e2e

package tests

import (
	"os/exec"
	"strings"
	"testing"
	"time"
)

const binary = "/usr/local/bin/minic"

func TestRunEcho(t *testing.T) {
	out, err := exec.Command(binary, "run", "--net", "none", "alpine", "echo", "hello-minic").CombinedOutput()
	if err != nil {
		t.Fatalf("run echo failed: %s: %v", string(out), err)
	}
	if !strings.Contains(string(out), "hello-minic") {
		t.Fatalf("expected 'hello-minic' in output, got: %s", string(out))
	}
}

func TestRunHostname(t *testing.T) {
	out, err := exec.Command(binary, "run", "--net", "none", "--hostname", "testbox", "alpine", "hostname").CombinedOutput()
	if err != nil {
		t.Fatalf("run hostname failed: %s: %v", string(out), err)
	}
	if !strings.Contains(string(out), "testbox") {
		t.Fatalf("expected 'testbox' in output, got: %s", string(out))
	}
}

func TestRunPidIsolation(t *testing.T) {
	out, err := exec.Command(binary, "run", "--net", "none", "alpine", "sh", "-c", "echo $$").CombinedOutput()
	if err != nil {
		t.Fatalf("run pid check failed: %s: %v", string(out), err)
	}
	trimmed := strings.TrimSpace(string(out))
	if trimmed != "1" {
		t.Fatalf("expected PID 1 inside container, got: %q", trimmed)
	}
}

func TestRunFilesystemIsolation(t *testing.T) {
	out, err := exec.Command(binary, "run", "--net", "none", "alpine", "cat", "/etc/os-release").CombinedOutput()
	if err != nil {
		t.Fatalf("run cat os-release failed: %s: %v", string(out), err)
	}
	if !strings.Contains(string(out), "Alpine") {
		t.Fatalf("expected Alpine in os-release, got: %s", string(out))
	}
}

func TestRunOverlayIsolation(t *testing.T) {
	exec.Command(binary, "run", "--net", "none", "alpine", "sh", "-c", "echo test > /overlay-test.txt").CombinedOutput()

	out, err := exec.Command(binary, "run", "--net", "none", "alpine", "ls", "/overlay-test.txt").CombinedOutput()
	if err == nil {
		t.Fatalf("expected file to not exist in new container, got: %s", string(out))
	}
}

func TestDetachAndPs(t *testing.T) {
	out, err := exec.Command(binary, "run", "-d", "--net", "none", "alpine", "sleep", "30").CombinedOutput()
	if err != nil {
		t.Fatalf("run detach failed: %s: %v", string(out), err)
	}

	containerID := strings.TrimSpace(string(out))
	time.Sleep(1 * time.Second)

	psOut, err := exec.Command(binary, "ps").CombinedOutput()
	if err != nil {
		t.Fatalf("ps failed: %s: %v", string(psOut), err)
	}
	if !strings.Contains(string(psOut), containerID) {
		t.Fatalf("expected container %s in ps output, got: %s", containerID, string(psOut))
	}

	stopOut, err := exec.Command(binary, "stop", containerID).CombinedOutput()
	if err != nil {
		t.Fatalf("stop failed: %s: %v", string(stopOut), err)
	}

	time.Sleep(1 * time.Second)

	rmOut, err := exec.Command(binary, "rm", containerID).CombinedOutput()
	if err != nil {
		t.Fatalf("rm failed: %s: %v", string(rmOut), err)
	}
}

func TestPullAndImages(t *testing.T) {
	out, err := exec.Command(binary, "pull", "alpine").CombinedOutput()
	if err != nil {
		t.Fatalf("pull failed: %s: %v", string(out), err)
	}

	imgOut, err := exec.Command(binary, "images").CombinedOutput()
	if err != nil {
		t.Fatalf("images failed: %s: %v", string(imgOut), err)
	}
	if !strings.Contains(string(imgOut), "alpine") {
		t.Fatalf("expected alpine in images output, got: %s", string(imgOut))
	}
}
