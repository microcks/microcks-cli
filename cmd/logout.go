/*
 * Copyright The Microcks Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmd

import (
	"fmt"

	"github.com/microcks/microcks-cli/pkg/config"
	"github.com/microcks/microcks-cli/pkg/connectors"
	"github.com/microcks/microcks-cli/pkg/errors"
	"github.com/spf13/cobra"
)

func NewLogoutCommand(globalClientOpts *connectors.ClientOptions) *cobra.Command {

	logoutCmd := &cobra.Command{
		Use:   "logout CONTEXT",
		Short: "Log out from Microcks",
		Long:  "Log out from Microcks",
		Example: `# Log out from a Microcks server URL
microcks logout http://localhost:8080

# Log out from a named context
microcks logout dev-context`,

		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.Wrapf(errors.KindUsage, "logout requires a CONTEXT or server URL argument")
			}

			target := args[0]
			if err := logoutContext(target, globalClientOpts.ConfigPath); err != nil {
				return err
			}
			fmt.Printf("Logged out from '%s'\n", target)
			return nil
		},
	}

	return logoutCmd
}

func logoutContext(target, configPath string) error {
	localCfg, err := config.ReadLocalConfig(configPath)
	if err != nil {
		return err
	}
	if localCfg == nil {
		return errors.Wrapf(errors.KindUsage, "nothing to log out from")
	}

	userName := target
	if ctx, err := localCfg.ResolveContext(target); err == nil {
		userName = ctx.User.Name
	}

	if ok := localCfg.RemoveToken(userName); !ok {
		return errors.Wrapf(errors.KindNotFound, "context %q does not exist", target)
	}

	err = config.ValidateLocalConfig(*localCfg)
	if err != nil {
		return errors.Wrap(errors.KindUsage, fmt.Errorf("logging out leaves local config invalid: %w", err))
	}

	return config.WriteLocalConfig(*localCfg, configPath)
}
