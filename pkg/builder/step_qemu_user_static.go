package builder

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

const wrapped = "-wrapped"

type Args struct {
	Args               []string
	PathToQemuInChroot string
}

type stepQemuUserStatic struct {
	ChrootKey             string
	PathToQemuInChrootKey string

	Args        Args
	destQemu    string
	destWrapper string
}

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

func (s *stepQemuUserStatic) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	// Read our value and assert that it is they type we want
	chrootDir := state.Get(s.ChrootKey).(string)
	config := state.Get("config").(*Config)

	ui := state.Get("ui").(packer.Ui)
	ui.Say("Installing qemu-user-static in the chroot")
	srcqemu := config.QemuBinary
	// TODO: maybe put qemu in the temporary dir in the root of the chroot? to guarantee
	// existance of directory and easy cleanup
	s.destQemu = filepath.Join(chrootDir, srcqemu)
	s.Args.PathToQemuInChroot = srcqemu
	state.Put(s.PathToQemuInChrootKey, s.Args.PathToQemuInChroot)

	err := run(state, fmt.Sprintf("cp %s %s", srcqemu, s.destQemu))
	if err != nil {
		return multistep.ActionHalt
	}

	err = s.makeWrapper(ui, state)
	if err != nil {
		return multistep.ActionHalt
	}
	return multistep.ActionContinue
}

func (s *stepQemuUserStatic) makeWrapper(ui packer.Ui, state multistep.StateBag) error {
	if len(s.Args.Args) == 0 {
		return nil
	}

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

	destWrapper := s.destQemu
	s.destQemu += wrapped

	err = run(state, fmt.Sprintf("mv %s %s", destWrapper, s.destQemu))
	if err != nil {
		s.destQemu = destWrapper
		return err
	}

	ui.Say("compiling arguments wrapper")
	err = run(state, fmt.Sprintf("gcc -g -static %s -o %s", tmpfn, destWrapper))
	if err != nil {
		return err
	}

	s.destWrapper = destWrapper
	return nil
}

func (s *stepQemuUserStatic) Cleanup(state multistep.StateBag) {
	if s.destQemu != "" {
		os.Remove(s.destQemu)
	}
	if s.destWrapper != "" {
		os.Remove(s.destWrapper)
	}
}
