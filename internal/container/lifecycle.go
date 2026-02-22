//go:build linux

package container

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/sevladev/minic/internal/cgroup"
	"github.com/sevladev/minic/internal/filesystem"
	"github.com/sevladev/minic/internal/image"
	"github.com/sevladev/minic/internal/namespace"
	"github.com/sevladev/minic/internal/network"
	"github.com/sevladev/minic/pkg/idgen"
)

func Run(cfg Config) error {
	if !image.Exists(cfg.Image) {
		return fmt.Errorf("image %q not found\nRun: minic pull %s", cfg.Image, cfg.Image)
	}

	containerID := idgen.New()

	overlay, err := filesystem.SetupOverlay(containerID, image.RootfsPath(cfg.Image))
	if err != nil {
		return fmt.Errorf("setup overlay: %w", err)
	}

	hostname := cfg.Hostname
	if hostname == "" {
		hostname = "minic"
	}

	meta := &Metadata{
		ID:        containerID,
		Image:     cfg.Image,
		Command:   cfg.Command,
		Status:    StateCreated,
		Hostname:  hostname,
		CreatedAt: time.Now(),
	}

	args := append([]string{"init"}, cfg.Command...)
	cmd := exec.Command("/proc/self/exe", args...)
	cmd.SysProcAttr = namespace.NewSysProcAttr()
	cmd.Env = append(os.Environ(),
		"MINIC_HOSTNAME="+hostname,
		"MINIC_ROOTFS="+overlay.Merged,
	)

	stdoutLog, _ := os.Create(StdoutLogPath(containerID))
	stderrLog, _ := os.Create(StderrLogPath(containerID))

	if cfg.Detach {
		cmd.Stdout = stdoutLog
		cmd.Stderr = stderrLog
	} else {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Start(); err != nil {
		filesystem.RemoveOverlay(overlay)
		return fmt.Errorf("start container: %w", err)
	}

	pid := cmd.Process.Pid
	meta.PID = pid

	if cfg.NetMode != "none" {
		net, err := network.Setup(containerID, pid)
		if err != nil {
			cmd.Process.Kill()
			filesystem.RemoveOverlay(overlay)
			return fmt.Errorf("setup network: %w", err)
		}
		meta.IP = net.IP.String()
	}

	hasCgroup := false
	limits := cgroup.Limits{
		MemoryBytes: cfg.Resources.MemoryBytes,
		CPUQuota:    cfg.Resources.CPUQuota,
		PidsMax:     cfg.Resources.PidsMax,
	}
	if limits.MemoryBytes > 0 || limits.CPUQuota > 0 || limits.PidsMax > 0 {
		if err := cgroup.Apply(containerID, pid, limits); err != nil {
			cmd.Process.Kill()
			filesystem.RemoveOverlay(overlay)
			return fmt.Errorf("apply cgroups: %w", err)
		}
		hasCgroup = true
	}

	meta.Status = StateRunning
	SaveState(meta)

	if cfg.Detach {
		fmt.Printf("%s\n", containerID)
		go func() {
			cmd.Wait()
			stdoutLog.Close()
			stderrLog.Close()
			if hasCgroup {
				cgroup.Remove(containerID)
			}
			network.Cleanup(containerID)
			filesystem.RemoveOverlay(overlay)
			meta.Status = StateStopped
			meta.ExitCode = cmd.ProcessState.ExitCode()
			SaveState(meta)
		}()
		return nil
	}

	err = cmd.Wait()
	stdoutLog.Close()
	stderrLog.Close()

	if hasCgroup {
		cgroup.Remove(containerID)
	}
	network.Cleanup(containerID)
	filesystem.RemoveOverlay(overlay)

	meta.Status = StateStopped
	if cmd.ProcessState != nil {
		meta.ExitCode = cmd.ProcessState.ExitCode()
	}
	SaveState(meta)

	if err != nil {
		return fmt.Errorf("container exited: %w", err)
	}

	return nil
}

func Stop(id string) error {
	meta, err := FindByPrefix(id)
	if err != nil {
		return err
	}

	if meta.Status != StateRunning {
		return fmt.Errorf("container %s is not running", meta.ID)
	}

	process, err := os.FindProcess(meta.PID)
	if err != nil {
		return fmt.Errorf("find process: %w", err)
	}

	if err := process.Signal(syscall.SIGTERM); err != nil {
		process.Signal(syscall.SIGKILL)
	}

	meta.Status = StateStopped
	SaveState(meta)
	fmt.Println(meta.ID)
	return nil
}

func Remove(id string) error {
	meta, err := FindByPrefix(id)
	if err != nil {
		return err
	}

	if meta.Status == StateRunning {
		return fmt.Errorf("container %s is still running, stop it first", meta.ID)
	}

	RemoveState(meta.ID)
	fmt.Println(meta.ID)
	return nil
}
