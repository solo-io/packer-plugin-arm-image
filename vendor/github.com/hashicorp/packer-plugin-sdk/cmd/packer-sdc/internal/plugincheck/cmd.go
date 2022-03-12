package plugincheck

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

var (
	//go:embed README.md
	readme string
)

type Command struct {
}

func (cmd *Command) Help() string {
	return "\n" + readme
}

func (cmd *Command) Run(args []string) int {
	if err := cmd.run(args); err != nil {
		log.Printf("%v", err)
		return 1
	}
	return 0
}

func (cmd *Command) run(args []string) error {

	if len(args) != 1 {
		cmd.Help()
		return errors.New("plugin-check requires a plugin binary name as an argument.\n" +
			"ex: 'packer-plugin-happycloud'. Check will be run on the binary.")
	}

	pluginName := args[0]

	if isOldPlugin(pluginName) {
		fmt.Printf("\n[WARNING] Plugin is named with old prefix `packer-[builder|provisioner|post-processor]-{name})`. " +
			"These will be detected but Packer cannot install them automatically. " +
			"The plugin must be a multi-component plugin named packer-plugin-{name} to be installable through the `packer init` command.\n" +
			"See docs at: https://www.packer.io/docs/plugins.\n")
		return nil
	}

	if err := checkPluginName(pluginName); err != nil {
		return err
	}

	path, err := filepath.Abs(pluginName)
	if err != nil {
		return err
	}

	output, err := exec.Command(path, "describe").Output()
	if err != nil {
		return errors.Wrap(err, "failed to describe plugin")
	}

	desc := pluginDescription{}
	err = json.Unmarshal(output, &desc)
	if err != nil {
		return errors.Wrap(err, "failed to json.Unmarshal plugin description")
	}
	if len(desc.Version) == 0 {
		return errors.New("Version needs to be set")
	}
	if len(desc.SDKVersion) == 0 {
		return errors.New("SDKVersion needs to be set")
	}
	if len(desc.APIVersion) == 0 {
		return errors.New("APIVersion needs to be set")
	}

	if len(desc.Builders) == 0 && len(desc.PostProcessors) == 0 && len(desc.Datasources) == 0 && len(desc.Provisioners) == 0 {
		return errors.New("this plugin defines no component.")
	}
	return nil
}

type pluginDescription struct {
	Version        string   `json:"version"`
	SDKVersion     string   `json:"sdk_version"`
	APIVersion     string   `json:"api_version"`
	Builders       []string `json:"builders"`
	PostProcessors []string `json:"post_processors"`
	Datasources    []string `json:"datasources"`
	Provisioners   []string `json:"provisioners"`
}

func isOldPlugin(pluginName string) bool {
	return strings.HasPrefix(pluginName, "packer-builder-") ||
		strings.HasPrefix(pluginName, "packer-provisioner-") ||
		strings.HasPrefix(pluginName, "packer-post-processor-")
}

// checkPluginName checks for the possible valid names for a plugin,
// packer-plugin-* or packer-[builder|provisioner|post-processor]-*. If the name
// is prefixed with `packer-[builder|provisioner|post-processor]-`, packer won't
// be able to install it, therefore a WARNING will be shown.
func checkPluginName(name string) error {
	if strings.HasPrefix(name, "packer-plugin-") {
		return nil
	}

	return fmt.Errorf("plugin name is not valid")
}

func (cmd *Command) Synopsis() string {
	return "Tell wether a plugin release looks valid for Packer."
}
