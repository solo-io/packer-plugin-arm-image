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

const qemyBinary = "/usr/bin/qemu-arm-static"

type Args struct {
	Args          []string
	WrapperSuffix string
}

type stepQemuUserStatic struct {
	ChrootKey   string
	Args        Args
	destQemu    string
	destWrapper string
}

const argsTemplate = `
#include <stdlib.h>
#include <unistd.h>

int main(int argc, char **argv, char **envp) {
 unsigned int length = {{len .Args}}; 
 char *qemuargs = malloc(sizeof(argv[0]) * (argc + length + 1));
 int index = 0;

 qemuargs[index++] = argv[0];

 {{range .Args}}qemuargs[index++] = "{{.}}";
 {{end}}
 qemuargs[argc + length] = NULL;
 memcpy(qemuargs + index, argv + 1, sizeof(argv[0]) * (argc - 1));
 return execve("/usr/bin/qemu-arm-static{{.WrapperSuffix}}", qemuargs, envp);
}
`

func (s *stepQemuUserStatic) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	// Read our value and assert that it is they type we want
	chrootDir := state.Get(s.ChrootKey).(string)
	ui := state.Get("ui").(packer.Ui)
	ui.Say("Installing qemu-arm-static in the chroot")
	srcqemu := qemyBinary
	s.destQemu = filepath.Join(chrootDir, srcqemu)

	err := s.makeWrapper(ui, state)
	if err != nil {
		return multistep.ActionHalt
	}

	err = run(state, fmt.Sprintf("cp %s %s", srcqemu, s.destQemu))
	if err != nil {
		return multistep.ActionHalt
	}
	return multistep.ActionContinue
}

func (s *stepQemuUserStatic) makeWrapper(ui packer.Ui, state multistep.StateBag) error {
	if len(s.Args.Args) == 0 {
		return nil
	}
	s.Args.WrapperSuffix = "-wrapped"

	t := template.Must(template.New("qemu-wrapper").Parse(argsTemplate))

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
	s.destQemu += s.Args.WrapperSuffix

	err = run(state, fmt.Sprintf("mv %s %s", destWrapper, s.destQemu))
	if err != nil {
		s.destQemu = destWrapper
		return err
	}

	ui.Say("compiling arguments wrapper")
	err = run(state, fmt.Sprintf("gcc -static %s -s -o %s", tmpfn, destWrapper))
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
