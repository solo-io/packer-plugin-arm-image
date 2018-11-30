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
	Kali  KnownImageType = "kali"
)

func GuessImageType(url string) KnownImageType {
	if strings.Contains(url, "raspbian") {
		return RaspberryPi
	}

	if strings.Contains(url, "bone") {
		return BeagleBone
	}

	if strings.Contains(url, "kali") {
		return Kali
	}

	return ""

}

func GetImageFilesInCurrentDir() []string {
	files, err := ioutil.ReadDir(".")
	if err != nil {
		return nil
	}

	// optimisitic output dir for packer
	outputFiles, err := ioutil.ReadDir("./output/")
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
