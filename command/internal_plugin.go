package command

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform/plugin"
	"github.com/kardianos/osext"
)

// InternalPluginCommand is a Command implementation that allows plugins to be
// compiled into the main Terraform binary and executed via a subcommand.
type InternalPluginCommand struct {
	Meta
}

const TFSPACE = "-TFSPACE-"

// BuildPluginCommandString builds a special string for executing internal
// plugins. It has the following format:
//
// 	/path/to/terraform-TFSPACE-internal-plugin-TFSPACE-terraform-provider-aws
//
// We split the string on -TFSPACE- to build the command executor. The reason we
// use -TFSPACE- is so we can support spaces in the /path/to/terraform part.
func BuildPluginCommandString(pluginType, pluginName string) (string, error) {
	terraformPath, err := osext.Executable()
	if err != nil {
		return "", err
	}
	parts := []string{terraformPath, "internal-plugin", pluginType, pluginName}
	return strings.Join(parts, TFSPACE), nil
}

// Internal plugins do not support any CLI args, but we do receive flags that
// main.go:mergeEnvArgs has merged in from EnvCLI. Instead of making main.go
// aware of this exception, we strip all flags from our args. Flags are easily
// identified by the '-' prefix, ensured by the cli package used.
func StripArgFlags(args []string) []string {
	argsNoFlags := []string{}
	for i := range args {
		if !strings.HasPrefix(args[i], "-") {
			argsNoFlags = append(argsNoFlags, args[i])
		}
	}
	return argsNoFlags
}

func (c *InternalPluginCommand) Run(args []string) int {
	// strip flags from args, only use subcommands.
	args = StripArgFlags(args)

	if len(args) != 2 {
		log.Printf("Wrong number of args; expected: terraform internal-plugin pluginType pluginName")
		return 1
	}

	pluginType := args[0]
	pluginName := args[1]

	log.SetPrefix(fmt.Sprintf("%s-%s (internal) ", pluginName, pluginType))

	switch pluginType {
	case "provisioner":
		pluginFunc, found := InternalProvisioners[pluginName]
		if !found {
			log.Printf("[ERROR] Could not load provisioner: %s", pluginName)
			return 1
		}
		log.Printf("[INFO] Starting provisioner plugin %s", pluginName)
		plugin.Serve(&plugin.ServeOpts{
			ProvisionerFunc: pluginFunc,
		})
	default:
		log.Printf("[ERROR] Invalid plugin type %s", pluginType)
		return 1
	}

	return 0
}

func (c *InternalPluginCommand) Help() string {
	helpText := `
Usage: terraform internal-plugin pluginType pluginName

  Runs an internally-compiled version of a plugin from the terraform binary.

  NOTE: this is an internal command and you should not call it yourself.
`

	return strings.TrimSpace(helpText)
}

func (c *InternalPluginCommand) Synopsis() string {
	return "internal plugin command"
}
