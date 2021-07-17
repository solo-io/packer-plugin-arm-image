package utils

import (
	"io/ioutil"
	"os"
	"strings"
)

type KnownImageType string

const (
	RaspberryPi KnownImageType = "raspberrypi"
	BeagleBone  KnownImageType = "beaglebone"
	Kali        KnownImageType = "kali"
	Ubuntu      KnownImageType = "ubuntu"
	Unknown     KnownImageType = ""
)

func GuessImageType(url string) KnownImageType {
	if strings.Contains(url, "raspbian") {
		return RaspberryPi
	}
	if strings.Contains(url, "raspios") {
		return RaspberryPi
	}

	if strings.Contains(url, "bone") {
		return BeagleBone
	}

	if strings.Contains(url, "kali") {
		return Kali
	}

	if strings.Contains(url, "ubuntu") {
		return Ubuntu
	}

	return ""

}

func GetImageFilesInCurrentDir() []string {
	files, err := ioutil.ReadDir(".")
	if err != nil {
		return nil
	}

	// optimisitic output dir for packer
	outputFiles, err := ioutil.ReadDir("./output-arm-image/")
	if err == nil {
		files = append(files, outputFiles...)
	}

	var potentialFiles []string

	for _, file := range files {
		if hasPotential(file) {
			potentialFiles = append(potentialFiles, file.Name())
		}
	}

	return potentialFiles
}

func hasPotential(info os.FileInfo) bool {
	if info.Name() == "image" {
		// this is the default output name
		return true
	}
	if GuessImageType(info.Name()) != "" {
		return true
	}
	if strings.HasSuffix(info.Name(), ".img") {
		return true
	}
	if strings.HasSuffix(info.Name(), ".iso") {
		return true
	}
	if strings.HasSuffix(info.Name(), ".xz") {
		return true
	}

	return false
}
