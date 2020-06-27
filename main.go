package main

import (
	"fmt"
	"os"

	"github.com/minamijoyo/tfmigrate/tfexec"
)

func main() {
	e := tfexec.NewDefaultExecutor()
	terraformCLI := tfexec.NewTerraformCLI(e)
	err := terraformCLI.Version([]string{"-json"})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
