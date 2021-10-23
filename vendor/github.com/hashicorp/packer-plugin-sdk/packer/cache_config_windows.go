// +build windows

package packer

const (
	defaultConfigFile = "packer_cache"
)

func getDefaultCacheDir() string {
	return defaultConfigFile
}
