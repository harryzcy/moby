package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/containerd/log"
)

const (
	// StockRuntimeName is used by the 'default-runtime' flag in dockerd as the
	// default value. On Windows keep this empty so the value is auto-detected
	// based on other options.
	StockRuntimeName = ""

	WindowsV1RuntimeName = "com.docker.hcsshim.v1"
	WindowsV2RuntimeName = "io.containerd.runhcs.v1"
)

var builtinRuntimes = map[string]bool{
	WindowsV1RuntimeName: true,
	WindowsV2RuntimeName: true,
}

// BridgeConfig is meant to store all the parameters for both the bridge driver and the default bridge network. On
// Windows: 1. "bridge" in this context reference the nat driver and the default nat network; 2. the nat driver has no
// specific parameters, so this struct effectively just stores parameters for the default nat network.
type BridgeConfig struct {
	DefaultBridgeConfig
}

type DefaultBridgeConfig struct {
	commonBridgeConfig

	// MTU is not actually used on Windows, but the --mtu option has always
	// been there on Windows (but ignored).
	MTU int `json:"mtu,omitempty"`
}

// Config defines the configuration of a docker daemon.
// It includes json tags to deserialize configuration from a file
// using the same names that the flags in the command line uses.
type Config struct {
	CommonConfig

	// Fields below here are platform specific. (There are none presently
	// for the Windows daemon.)
}

// GetExecRoot returns the user configured Exec-root
func (conf *Config) GetExecRoot() string {
	return ""
}

// GetInitPath returns the configured docker-init path
func (conf *Config) GetInitPath() string {
	return ""
}

// IsSwarmCompatible defines if swarm mode can be enabled in this config
func (conf *Config) IsSwarmCompatible() error {
	return nil
}

// ValidatePlatformConfig checks if any platform-specific configuration settings are invalid.
//
// Deprecated: this function was only used internally and is no longer used. Use [Validate] instead.
func (conf *Config) ValidatePlatformConfig() error {
	return validatePlatformConfig(conf)
}

// IsRootless returns conf.Rootless on Linux but false on Windows
func (conf *Config) IsRootless() bool {
	return false
}

func setPlatformDefaults(cfg *Config) error {
	cfg.Root = filepath.Join(os.Getenv("programdata"), "docker")
	cfg.ExecRoot = filepath.Join(os.Getenv("programdata"), "docker", "exec-root")
	cfg.Pidfile = filepath.Join(cfg.Root, "docker.pid")
	return nil
}

// validatePlatformConfig checks if any platform-specific configuration settings are invalid.
func validatePlatformConfig(conf *Config) error {
	if conf.MTU != 0 && conf.MTU != DefaultNetworkMtu {
		log.G(context.TODO()).Warn(`WARNING: MTU for the default network is not configurable on Windows, and this option will be ignored.`)
	}
	if conf.FirewallBackend != "" {
		return errors.New("firewall-backend can only be configured on Linux")
	}
	return nil
}

// validatePlatformExecOpt validates if the given exec-opt and value are valid
// for the current platform.
func validatePlatformExecOpt(opt, value string) error {
	switch opt {
	case "isolation":
		// TODO(thaJeztah): add validation that's currently in Daemon.setDefaultIsolation()
		return nil
	case "native.cgroupdriver":
		return fmt.Errorf("option '%s' is only supported on linux", opt)
	default:
		return fmt.Errorf("unknown option: '%s'", opt)
	}
}
