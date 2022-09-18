package arch

import (
	"runtime"
)

type KnownArchType string

const (
	Unknown KnownArchType = ""
	Arm     KnownArchType = "arm"
	ArmBE   KnownArchType = "armbe"
	Arm64   KnownArchType = "arm64"
	Arm64BE KnownArchType = "arm64be"
)

var knownValues = map[KnownArchType]string{
	Arm:     string(Arm),
	ArmBE:   string(ArmBE),
	Arm64:   string(Arm64),
	Arm64BE: string(Arm64BE),
}

func Values() []string {
	var values []string
	for _, value := range knownValues {
		values = append(values, value)
	}
	return values
}

func (arch KnownArchType) Valid() bool {
	_, ok := knownValues[arch]
	return ok
}

func (arch KnownArchType) IsNative() bool {
	return string(arch) == runtime.GOARCH
}
