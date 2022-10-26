package renderdocs

import (
	"bytes"
	"embed"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	fs "github.com/hashicorp/packer-plugin-sdk/cmd/packer-sdc/internal/fs"
	"github.com/pkg/errors"
)

var (
	cmdPrefix = "renderdocs"

	//go:embed README.md
	readme string
)

type Command struct {
	SrcDir      string
	PartialsDir string
	DstDir      string
}

func (cmd *Command) Flags() *flag.FlagSet {
	fs := flag.NewFlagSet(cmdPrefix, flag.ExitOnError)
	fs.StringVar(&cmd.SrcDir, "src", "docs-src", "folder to copy docs from.")
	fs.StringVar(&cmd.PartialsDir, "partials", "docs-partials", "folder containing all mdx partials.")
	fs.StringVar(&cmd.DstDir, "dst", "docs-rendered", "output folder.")
	return fs
}

func (cmd *Command) Help() string {
	cmd.Flags().Usage()
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
	f := cmd.Flags()
	err := f.Parse(args)
	if err != nil {
		return errors.Wrap(err, "unable to parse flags")
	}
	log.Printf("Copying %q to %q", cmd.SrcDir, cmd.DstDir)
	if err := fs.SyncDir(cmd.SrcDir, cmd.DstDir); err != nil {
		return errors.Wrap(err, "SyncDir failed")
	}
	log.Printf("Replacing @include '...' calls in %s", cmd.DstDir)

	return RenderDocsFolder(cmd.DstDir, cmd.PartialsDir)
}

func RenderDocsFolder(folder, partials string) error {
	entries, err := ioutil.ReadDir(folder)
	if err != nil {
		return errors.Wrapf(err, "cannot read directory %s", folder)
	}

	for _, entry := range entries {
		entryPath := filepath.Join(folder, entry.Name())
		if entry.IsDir() {
			if err = RenderDocsFolder(entryPath, partials); err != nil {
				return err
			}
		} else {
			if err = renderDocsFile(entryPath, partials); err != nil {
				return errors.Wrap(err, "renderDocsFile")
			}
		}
	}
	return nil
}

var (
	includeStr = []byte("@include '")
)

func renderDocsFile(filePath, partialsDir string) error {
	f, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	for i := 0; i+len(includeStr) < len(f); i++ {
		if f[i] != '@' {
			continue
		}
		if diff := bytes.Compare(f[i:i+len(includeStr)], includeStr); diff != 0 {
			continue
		}
		ii := i + len(includeStr)
		for ; ii < len(f); ii++ {
			if f[ii] == '\'' {
				break
			}
		}
		if ii == len(includeStr) || f[ii] != '\'' {
			log.Printf("Unclosed @include quote at %d in %s", ii, filePath)
		}
		partialPath := string(f[i+len(includeStr) : ii])
		partial, err := getPartial(partialsDir, partialPath)
		if err != nil {
			return err
		}
		f = append(f[:i], append(partial, f[ii+1:]...)...)
	}

	return os.WriteFile(filePath, f, 0)
}

//go:embed docs-partials/*
var partialFiles embed.FS

// getPartial will first try to look for partials in the
// renderdocs/docs-partials dir. This makes common/shared partials available to
// all docs with for example:
//  @include 'packer-plugin-sdk/communicator/Config.mdx'
// Otherwise it tries to find a partial in/ the actual filesystem.
func getPartial(partialsDir, partialPath string) ([]byte, error) {
	if partial, err := partialFiles.ReadFile(strings.Join([]string{"docs-partials", partialPath}, "/")); err == nil {
		return partial, nil
	}
	return os.ReadFile(filepath.Join(partialsDir, partialPath))
}

func (cmd *Command) Synopsis() string {
	return "From a src directory and a partials directory, generate the end result docs into a folder."
}
