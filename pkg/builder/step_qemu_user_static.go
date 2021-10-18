package builder

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

const wrapped = "-wrapped"

type Args struct {
	Args               []string
	PathToQemuInChroot string
}

type stepQemuUserStatic struct {
	ChrootKey             string
	PathToQemuInChrootKey string

	Args                    Args
	qemuDestinationInChroot string
	destWrapper             string
}

// if we need to pass args to qemu, we need to compile a static wrapper
const argsTemplate = `
// for malloc
#include <stdlib.h>
// for execve
#include <unistd.h>
// for memcpy
#include <string.h>

int main(int argc, char **argv, char **envp) {
	unsigned int length = {{len .Args}}; 
	char **qemuargs = malloc(sizeof(argv[0]) * (argc + length + 1));
	int index = 0;

	qemuargs[index++] = argv[0];

	{{range .Args}}qemuargs[index++] = "{{.}}";
	{{end}}
	memcpy(qemuargs + index, argv + 1, sizeof(argv[0]) * (argc - 1));
	qemuargs[argc + length] = NULL;
	return execve("{{.PathToQemuInChroot}}", qemuargs, envp);
}
`

func (s *stepQemuUserStatic) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	// Read our value and assert that it is they type we want
	chrootDir := state.Get(s.ChrootKey).(string)
	config := state.Get("config").(*Config)

	ui := state.Get("ui").(packer.Ui)
	ui.Say(fmt.Sprintf("Installing qemu-user-static (%s) in the chroot", config.QemuBinary))
	qemuInHostPath := config.QemuBinary
	_, qemuFilename := filepath.Split(qemuInHostPath)

	// place qemu in the root dir in the chroot, as it is guaranteed to exist
	s.Args.PathToQemuInChroot = "/" + qemuFilename

	s.qemuDestinationInChroot = filepath.Join(chrootDir, s.Args.PathToQemuInChroot)
	state.Put(s.PathToQemuInChrootKey, s.Args.PathToQemuInChroot)

	err := run(ctx, state, fmt.Sprintf("cp %s %s", qemuInHostPath, s.qemuDestinationInChroot))
	if err != nil {
		return multistep.ActionHalt
	}

	err = s.makeWrapper(ctx, ui, state)
	if err != nil {
		return multistep.ActionHalt
	}
	return multistep.ActionContinue
}

func (s *stepQemuUserStatic) makeWrapper(ctx context.Context, ui packer.Ui, state multistep.StateBag) error {
	if len(s.Args.Args) == 0 {
		return nil
	}

	// prepare source file for wrapper
	t := template.Must(template.New("qemu-wrapper").Parse(argsTemplate))

	s.Args.PathToQemuInChroot += wrapped

	var buffer bytes.Buffer
	t.Execute(&buffer, s.Args)

	dir, err := ioutil.TempDir("", "compile")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir) // clean up

	tmpfn := filepath.Join(dir, "main.c")
	if err := ioutil.WriteFile(tmpfn, buffer.Bytes(), 0666); err != nil {
		return err
	}

	// move original qemu. keep track so we can clean up
	destWrapper := s.qemuDestinationInChroot
	s.qemuDestinationInChroot += wrapped

	err = run(ctx, state, fmt.Sprintf("mv %s %s", destWrapper, s.qemuDestinationInChroot))
	if err != nil {
		s.qemuDestinationInChroot = destWrapper
		return err
	}

	// compile wrapper to the location of the original qemu
	ui.Say("compiling arguments wrapper")
	err = run(ctx, state, fmt.Sprintf("gcc -g -static %s -o %s", tmpfn, destWrapper))
	if err != nil {
		return err
	}

	// keep track so we can clean up
	s.destWrapper = destWrapper
	return nil
}

func (s *stepQemuUserStatic) Cleanup(state multistep.StateBag) {
	if s.qemuDestinationInChroot != "" {
		os.Remove(s.qemuDestinationInChroot)
	}
	if s.destWrapper != "" {
		os.Remove(s.destWrapper)
	}
}
