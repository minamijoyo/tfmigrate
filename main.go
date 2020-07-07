package main

import (
	"context"
	"fmt"
	"os"

	"github.com/minamijoyo/tfmigrate/tfexec"
)

func main() {
	e := tfexec.NewExecutor(".", os.Environ())
	terraformCLI := tfexec.NewTerraformCLI(e)
	v, err := terraformCLI.Version(context.Background())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	fmt.Fprintln(os.Stdout, v)
}
