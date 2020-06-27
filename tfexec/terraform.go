package tfexec

// TerraformCLI is an interface for executing the terraform command.
type TerraformCLI interface {
	// Version executes a terraform version command.
	Version(options []string) error
	// Init executes a terraform init command.
	Init(options []string, dir string) error
	// Plan executes a terraform plan command.
	Plan(options []string, dir string) error
	// Show executes a terraform show command.
	Show(options []string, path string) error
	// Import executes a terraform import command.
	Import(options []string, addr string, id string) error
	// StatePull executes a terraform state pull command.
	StatePull(options []string) error
	// StatePush executes a terraform state push command.
	StatePush(options []string, path string) error
	// StateMv executes a terraform state mv command.
	StateMv(options []string, source string, destination string) error
	// StateRm executes a terraform state rm command.
	StateRm(options []string, address ...string) error
	// StateList executes a terraform state list command.
	StateList(options []string, address ...string) error
}

// terraformCLI implements the TerraformCLI interface.
type terraformCLI struct {
	// Executor is a componenet which executes an arbitrary command.
	*Executor
}

var _ TerraformCLI = (*terraformCLI)(nil)

// NewTerraformCLI returns an implementation of the TerraformCLI interface.
func NewTerraformCLI(e *Executor) TerraformCLI {
	return &terraformCLI{
		Executor: e,
	}
}

func (c *terraformCLI) run(args []string) error {
	cmd := c.Executor.NewCommand("terraform", args...)
	return c.Executor.Run(cmd)
}

// Version executes a terraform version command.
func (c *terraformCLI) Version(options []string) error {
	args := []string{"version"}
	args = append(args, options...)
	return c.run(args)
}

// Init executes a terraform init command.
func (c *terraformCLI) Init(options []string, dir string) error {
	args := []string{"init"}
	args = append(args, options...)
	args = append(args, dir)
	return c.run(args)
}

// Plan executes a terraform plan command.
func (c *terraformCLI) Plan(options []string, dir string) error {
	args := []string{"plan"}
	args = append(args, options...)
	args = append(args, dir)
	return c.run(args)
}

// Show executes a terraform show command.
func (c *terraformCLI) Show(options []string, path string) error {
	args := []string{"show"}
	args = append(args, options...)
	args = append(args, path)
	return c.run(args)
}

// Import executes a terraform import command.
func (c *terraformCLI) Import(options []string, addr string, id string) error {
	args := []string{"import"}
	args = append(args, options...)
	args = append(args, addr, id)
	return c.run(args)
}

// StatePull executes a terraform state pull command.
func (c *terraformCLI) StatePull(options []string) error {
	args := []string{"state", "pull"}
	args = append(args, options...)
	return c.run(args)
}

// StatePush executes a terraform state push command.
func (c *terraformCLI) StatePush(options []string, path string) error {
	args := []string{"state", "push"}
	args = append(args, options...)
	args = append(args, path)
	return c.run(args)
}

// StateMv executes a terraform state mv command.
func (c *terraformCLI) StateMv(options []string, source string, destination string) error {
	args := []string{"state", "mv"}
	args = append(args, options...)
	args = append(args, source, destination)
	return c.run(args)
}

// StateRm executes a terraform state rm command.
func (c *terraformCLI) StateRm(options []string, address ...string) error {
	args := []string{"state", "rm"}
	args = append(args, options...)
	args = append(args, address...)
	return c.run(args)
}

// StateList executes a terraform state list command.
func (c *terraformCLI) StateList(options []string, address ...string) error {
	args := []string{"state", "list"}
	args = append(args, options...)
	args = append(args, address...)
	return c.run(args)
}
