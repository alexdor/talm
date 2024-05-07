// Code generated by go run tools/import_commands.go --talos-version v1.7.1 reboot
// DO NOT EDIT.

// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package commands

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/siderolabs/talos/cmd/talosctl/pkg/talos/action"
	"github.com/siderolabs/talos/cmd/talosctl/pkg/talos/helpers"
	"github.com/siderolabs/talos/pkg/machinery/client"
)

var rebootCmdFlags struct {
	trackableActionCmdFlags
	mode        string
	configFiles []string
}

// rebootCmd represents the reboot command.
var rebootCmd = &cobra.Command{
	Use:   "reboot",
	Short: "Reboot a node",
	Long:  ``,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if rebootCmdFlags.debug {
			rebootCmdFlags.wait = true
		}

		var opts []client.RebootMode

		switch rebootCmdFlags.mode {
		// skips kexec and reboots with power cycle
		case "powercycle":
			opts = append(opts, client.WithPowerCycle)
		case "default":
		default:
			return fmt.Errorf("invalid reboot mode: %q", rebootCmdFlags.mode)
		}

		if !rebootCmdFlags.wait {
			return WithClient(func(ctx context.Context, c *client.Client) error {
				if err := helpers.ClientVersionCheck(ctx, c); err != nil {
					return err
				}

				if err := c.Reboot(ctx, opts...); err != nil {
					return fmt.Errorf("error executing reboot: %s", err)
				}

				return nil
			})
		}

		return action.NewTracker(
			&GlobalArgs,
			action.MachineReadyEventFn,
			rebootGetActorID(opts...),
			action.WithPostCheck(action.BootIDChangedPostCheckFn),
			action.WithDebug(rebootCmdFlags.debug),
			action.WithTimeout(rebootCmdFlags.timeout),
		).Run()
	},
}

func rebootGetActorID(opts ...client.RebootMode) func(ctx context.Context, c *client.Client) (string, error) {
	return func(ctx context.Context, c *client.Client) (string, error) {
		resp, err := c.RebootWithResponse(ctx, opts...)
		if err != nil {
			return "", err
		}

		if len(resp.GetMessages()) == 0 {
			return "", errors.New("no messages returned from action run")
		}

		return resp.GetMessages()[0].GetActorId(), nil
	}
}

func init() {
	rebootCmd.Flags().StringSliceVarP(&rebootCmdFlags.configFiles,
		"file", "f", nil, "specify config files or patches in a YAML file (can specify multiple)",
	)
	rebootCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		nodesFromArgs := len(
			GlobalArgs.
				Nodes) > 0
		endpointsFromArgs := len(GlobalArgs.Endpoints) > 0
		for _, configFile := range rebootCmdFlags.configFiles {
			if err := processModelineAndUpdateGlobals(configFile, nodesFromArgs,

				endpointsFromArgs,
				false); err !=
				nil {
				return err
			}
		}
		return nil
	}

	rebootCmd.Flags().StringVarP(&rebootCmdFlags.mode, "mode", "m", "default", "select the reboot mode: \"default\", \"powercycle\" (skips kexec)")
	rebootCmdFlags.addTrackActionFlags(rebootCmd)
	addCommand(rebootCmd)
}
