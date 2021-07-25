package utils

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type Mountable interface {
	DevicePath() string
	UUID() string
}

type MountEntry struct {
	Device, Mountpoint, Type, Options string
}
type MountTable struct {
	Entries []MountEntry
}

func (mt *MountTable) Find(m Mountable) *MountEntry {
	for _, e := range mt.Entries {
		if e.Device == m.DevicePath() {
			return &e
		}
		if m.UUID() != "" && e.Device == filepath.Join("/dev/disk/by-uuid/", m.UUID()) {
			return &e
		}
	}
	return nil
}

func NewMountTable() (*MountTable, error) {
	data, err := ioutil.ReadFile("/proc/mounts")
	if err != nil {
		return nil, err
	}
	return ParseMountTable(data)
}
func ParseMountTable(data []byte) (*MountTable, error) {
	var mt MountTable
	stringdata := string(data)
	lines := strings.Split(stringdata, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		entries := strings.Split(line, " ")
		if len(entries) < 4 {
			return nil, errors.New("unexpected table format")
		}
		mt.Entries = append(mt.Entries, MountEntry{
			Device:     entries[0],
			Mountpoint: unescape(entries[1]),
			Type:       entries[2],
			Options:    entries[3],
		})
	}
	return &mt, nil
}

func unescape(s string) string {
	news, err := strconv.Unquote(`"` + s + `"`)
	if err != nil {
		//this should never happen
		panic(err)
	}
	return news
}

type UdevAdm struct {
	Values map[string]string
}

func NewUdevAdm(name string) (*UdevAdm, error) {

	data, err := exec.Command("udevadm", "info", "--query=property", "--name="+name).Output()
	if err != nil {
		return nil, err
	}
	return ParseUdevAdm(data)

}

func ParseUdevAdm(data []byte) (*UdevAdm, error) {
	var udevAdm UdevAdm
	udevAdm.Values = make(map[string]string)
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		entries := strings.SplitN(line, "=", 2)
		if len(entries) < 2 {
			return nil, errors.New("unexpected table format")
		}
		udevAdm.Values[entries[0]] = entries[1]
	}
	return &udevAdm, nil
}

type StringOrBool struct {
	Value bool
}

func (sb *StringOrBool) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		sb.Value = (s == "1") || (strings.ToLower(s) == "true")
		return nil
	}

	var bool_ bool
	if err := json.Unmarshal(b, &bool_); err != nil {
		return err
	}

	sb.Value = bool_
	return nil
}

type StringOrInt struct {
	Value int64
}

func (sb *StringOrInt) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		v, err := strconv.Atoi(s)
		sb.Value = int64(v)
		return err
	}

	if err := json.Unmarshal(b, &sb.Value); err != nil {
		return err
	}
	return nil
}

type LSBLKDevice struct {
	Name       string        `json:"name"`
	Model      string        `json:"model"`
	Size       StringOrInt   `json:"size"`
	Ro         StringOrBool  `json:"ro"`
	Rm         StringOrBool  `json:"rm"`
	DeviceUUID string        `json:"uuid"`
	Children   []LSBLKDevice `json:"children"`

	udevInfo *UdevAdm
}

func (l *LSBLKDevice) DevicePath() string {
	return "/dev/" + l.Name
}

func (l *LSBLKDevice) UUID() string {
	return l.DeviceUUID
}

func (l *LSBLKDevice) UDevInfo() (*UdevAdm, error) {
	if l.udevInfo != nil {
		return l.udevInfo, nil
	}
	udevinfo, err := NewUdevAdm(l.DevicePath())
	l.udevInfo = udevinfo
	return udevinfo, err
}

func (l *LSBLKDevice) Readonly() bool {
	return l.Ro.Value
}

func (l *LSBLKDevice) Removable() bool {
	return l.Rm.Value
}

type LSBLKDevices struct {
	Devices []LSBLKDevice `json:"blockdevices"`
}

func GetLSBLKDevices() (*LSBLKDevices, error) {
	data, err := exec.Command("lsblk", "-b", "--output", "NAME,SIZE,RO,RM,MODEL,UUID", "--json").Output()
	if err != nil {
		return nil, err
	}
	return ParseLSBLKDevices(data)
}

func ParseLSBLKDevices(data []byte) (*LSBLKDevices, error) {
	var l LSBLKDevices
	err := json.Unmarshal(data, &l)
	if err != nil {
		return nil, err
	}
	return &l, nil
}

type Device struct {
	Device      string
	Removable   bool
	ReadOnly    bool
	Name        string
	Mountpoints []string
}

func isReadonly(dev *LSBLKDevice) (bool, error) {
	return dev.Readonly(), nil
}

func isRemovable(dev *LSBLKDevice) (bool, error) {
	if dev.Removable() {
		return true, nil
	}

	udev, err := dev.UDevInfo()
	if err != nil {
		return false, err
	}
	if udev.Values["ID_DRIVE_FLASH_SD"] == "1" {
		return true, nil
	}
	if udev.Values["ID_DRIVE_MEDIA_FLASH_SD"] == "1" {
		return true, nil
	}
	return false, nil
}

func GetDetachableDevices() ([]Device, error) {
	alldevices, err := GetDevices()
	if err != nil {
		return nil, err
	}
	var detachable []Device
	for _, d := range alldevices {
		if d.Removable {
			detachable = append(detachable, d)
		}
	}
	return detachable, nil
}

func GetDevices() ([]Device, error) {
	devices, err := GetLSBLKDevices()
	if err != nil {
		return nil, err
	}
	mntTable, err := NewMountTable()
	if err != nil {
		return nil, err
	}

	var alldevices []Device

	for _, dev := range devices.Devices {
		// get all mount points
		rdev, err := GetDevice(&dev, mntTable)
		if err != nil {
			return nil, err
		}
		alldevices = append(alldevices, *rdev)
	}

	return alldevices, nil
}

func GetDevice(dev *LSBLKDevice, mntTable *MountTable) (*Device, error) {

	mntponts := []string{}

	var getMntPoints func(dev *LSBLKDevice)
	getMntPoints = func(dev *LSBLKDevice) {

		entry := mntTable.Find(dev)
		if entry != nil {
			mntponts = append(mntponts, entry.Mountpoint)
		}
		for _, childdev := range dev.Children {
			getMntPoints(&childdev)
		}
	}
	getMntPoints(dev)
	isro, err := isReadonly(dev)
	if err != nil {
		return nil, err
	}
	isrem, err := isRemovable(dev)
	if err != nil {
		return nil, err
	}
	rdev := Device{
		Device:      dev.DevicePath(),
		Name:        dev.Model,
		Mountpoints: mntponts,
		ReadOnly:    isro,
		Removable:   isrem,
	}
	if rdev.Name == "" {
		udev, err := dev.UDevInfo()

		if err != nil {
			return nil, err
		}
		rdev.Name = udev.Values["ID_NAME"]
	}
	return &rdev, nil
}
