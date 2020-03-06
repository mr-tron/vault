package command

import (
	"fmt"
	"strings"

	"github.com/mitchellh/cli"
	"github.com/posener/complete"
)

var _ cli.Command = (*OperatorRaftConfigurationCommand)(nil)
var _ cli.CommandAutocomplete = (*OperatorRaftConfigurationCommand)(nil)

type OperatorRaftConfigurationCommand struct {
	*BaseCommand
}

func (c *OperatorRaftConfigurationCommand) Synopsis() string {
	return "Returns the raft cluster configuration"
}

func (c *OperatorRaftConfigurationCommand) Help() string {
	helpText := `
Usage: vault operator raft configuration

  Provides the details of all the peers in the raft cluster.

	  $ vault operator raft configuration

` + c.Flags().Help()

	return strings.TrimSpace(helpText)
}

func (c *OperatorRaftConfigurationCommand) Flags() *FlagSets {
	set := c.flagSet(FlagSetHTTP | FlagSetOutputFormat)

	return set
}

func (c *OperatorRaftConfigurationCommand) AutocompleteArgs() complete.Predictor {
	return complete.PredictAnything
}

func (c *OperatorRaftConfigurationCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *OperatorRaftConfigurationCommand) Run(args []string) int {
	f := c.Flags()

	if err := f.Parse(args); err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	client, err := c.Client()
	if err != nil {
		c.UI.Error(err.Error())
		return 2
	}

	secret, err := client.Logical().Read("sys/storage/raft/configuration")
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error reading the raft cluster configuration: %s", err))
		return 2
	}
	if secret == nil {
		return 2
	}

	if Format(c.UI) != "table" {
		return OutputSecret(c.UI, secret)
	}

	config := secret.Data["config"].(map[string]interface{})

	servers := config["servers"].([]interface{})
	out := []string{"Node | Address | State | Voter"}
	for _, serverRaw := range servers {
		server := serverRaw.(map[string]interface{})
		state := "follower"
		if server["leader"].(bool) {
			state = "leader"
		}

		out = append(out, fmt.Sprintf("%s | %s | %s | %t", server["node_id"].(string), server["address"].(string), state, server["voter"].(bool)))
	}

	c.UI.Output(tableOutput(out, nil))
	return 0
}
