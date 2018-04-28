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
)

func GuessImageType(url string) KnownImageType {
	if strings.Contains(url, "raspbian") {
		return RaspberryPi
	}

	if strings.Contains(url, "bone") {
		return BeagleBone
	}

	return ""

}

func GetImageFilesInCurrentDir() []string {
	files, err := ioutil.ReadDir(".")
	if err != nil {
		return nil
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
	return GuessImageType(info.Name()) != ""
}
