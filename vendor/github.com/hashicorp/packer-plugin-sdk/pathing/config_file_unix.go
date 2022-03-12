//go:build darwin || freebsd || linux || netbsd || openbsd || solaris
// +build darwin freebsd linux netbsd openbsd solaris

package pathing

import (
	"errors"
	"log"
	"os"
	"path/filepath"
)

const (
	defaultConfigFile = ".packerconfig"
	defaultConfigDir  = ".packer.d"
)

func configDir() (path string, err error) {

	if cd := os.Getenv("PACKER_CONFIG_DIR"); cd != "" {
		log.Printf("Detected config directory from env var: %s", cd)
		return filepath.Join(cd, defaultConfigDir), nil
	}

	var dir string
	homedir := os.Getenv("HOME")

	if homedir == "" {
		return "", errors.New("No $HOME environment variable found, required to set Config Directory")
	}

	if hasDefaultConfigFileLocation(homedir) {
		dir = filepath.Join(homedir, defaultConfigDir)
		log.Printf("Old default config directory found: %s", dir)
	} else if xdgConfigHome := os.Getenv("XDG_CONFIG_HOME"); xdgConfigHome != "" {
		log.Printf("Detected xdg config directory from env var: %s", xdgConfigHome)
		dir = filepath.Join(xdgConfigHome, "packer")
	} else {
		dir = filepath.Join(homedir, ".config", "packer")
	}

	return dir, nil
}

func hasDefaultConfigFileLocation(homedir string) bool {
	if _, err := os.Stat(filepath.Join(homedir, defaultConfigDir)); err != nil {
		return false
	}
	return true
}
