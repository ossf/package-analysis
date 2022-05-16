package sandbox

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/ossf/package-analysis/internal/log"
)

const (
	ipBin           = "/usr/sbin/ip"
	iptablesLoadBin = "/usr/sbin/iptables-restore"
	iptablesRules   = "/usr/local/etc/iptables.rules"
	dummyInterface  = "cnidummy0"

	// bridgeInterface is the name of the podman bridge defined in
	// tools/network/podman-analysis.conflist. This bridge is used by the
	// sandbox during analysis to separate the sandbox traffic from the host.
	bridgeInterface = "cni-analysis"
)

const (
	// NetworkInterface is the name of a network interface that has access to
	// the sandbox network traffic.
	NetworkInterface = bridgeInterface
)

func loadIptablesRules() error {
	log.Debug("Loading iptable rules")

	// Open the iptables-restore configuration
	f, err := os.Open(iptablesRules)
	if err != nil {
		return err
	}
	defer f.Close()

	logOut := log.Writer(log.InfoLevel)
	defer logOut.Close()
	logErr := log.Writer(log.WarnLevel)
	defer logErr.Close()

	cmd := exec.Command(iptablesLoadBin)
	cmd.Stdout = logOut
	cmd.Stderr = logErr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	defer stdin.Close()
	if err := cmd.Start(); err != nil {
		return err
	}
	// Send the iptables rules to the command via stdin
	if _, err := io.Copy(stdin, f); err != nil {
		return err
	}
	stdin.Close()
	return cmd.Wait()
}

// createBridgeNetwork ensures that NetworkInterface and the bridge network
// exists prior to the sandbox.
//
// podman would create this bridge interface anyway, but doing it early allows
// a packet capture to be started on the interface prior to the sandbox
// starting.
func createBridgeNetwork() error {
	log.Debug("Creating bridge network")

	// Create the bridge
	cmd := exec.Command(ipBin, "link", "add", "name", bridgeInterface, "type", "bridge")
	if err := cmd.Run(); err != nil {
		// If the error is not an ExitError, or the Exit Code is not 2, then abort.
		if exitErr, ok := err.(*exec.ExitError); !ok || exitErr.ExitCode() != 2 {
			return fmt.Errorf("failed to add bridge interface: %w", err)
		}
	}

	// Bring the bridge up.
	cmd = exec.Command(ipBin, "link", "set", bridgeInterface, "up")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to bring up bridge interface: %w", err)
	}

	// Add a dummy device so the bridge stays up
	cmd = exec.Command(ipBin, "link", "add", "dev", dummyInterface, "type", "dummy")
	if err := cmd.Run(); err != nil {
		// If the error is not an ExitError, or the Exit Code is not 2, then abort.
		if exitErr, ok := err.(*exec.ExitError); !ok || exitErr.ExitCode() != 2 {
			return fmt.Errorf("failed to create dummy inteface: %w", err)
		}
	}

	// Add the dummy device to the bridge network
	cmd = exec.Command(ipBin, "link", "set", "dev", dummyInterface, "master", bridgeInterface)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add dummy interface to bridge: %w", err)
	}

	// Bring the dummy device up.
	cmd = exec.Command(ipBin, "link", "set", dummyInterface, "up")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to bring up dummy interface: %w", err)
	}

	return nil
}

// InitEnv initializes the host for running sandboxes.
//
// It will ensure that the network interface exists, and any firewall
// rules are configured.
//
// This function is idempotent and is safe to be called more than once.
//
// This function must be called after logging is complete, and may exit if
// any of the commands fail.
func InitEnv() {
	// Create the bridge network
	if err := createBridgeNetwork(); err != nil {
		log.Fatal("Failed to create bridge network", "error", err)
	}
	// Load iptables rules to further isolate the sandbox
	if err := loadIptablesRules(); err != nil {
		log.Fatal("Failed restoring iptables rules", "error", err)
	}
}
