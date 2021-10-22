// +build windows

package pathing

import (
	"log"
	"os"
	"path/filepath"
)

const (
	defaultConfigFile = "packer.config"
	defaultConfigDir  = "packer.d"
)

func configDir() (path string, err error) {

	if cd := os.Getenv("PACKER_CONFIG_DIR"); cd != "" {
		log.Printf("Detected config directory from env var: %s", cd)
		return filepath.Join(cd, defaultConfigDir), nil
	}

	homedir, err := homeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homedir, defaultConfigDir), nil
}
