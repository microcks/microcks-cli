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
	"github.com/microcks/microcks-cli/pkg/config"
	"github.com/microcks/microcks-cli/pkg/connectors"
	"github.com/microcks/microcks-cli/pkg/errors"
)

func newCommandClient(globalClientOpts *connectors.ClientOptions) (connectors.MicrocksClient, string, error) {
	config.InsecureTLS = globalClientOpts.InsecureTLS
	config.CaCertPaths = globalClientOpts.CaCertPaths
	config.Verbose = globalClientOpts.Verbose

	if globalClientOpts.ServerAddr != "" {
		mc, err := connectors.NewMicrocksClient(globalClientOpts.ServerAddr)
		if err != nil {
			return nil, "", err
		}

		if globalClientOpts.ClientId != "" && globalClientOpts.ClientSecret != "" {
			keycloakURL, err := mc.GetKeycloakURL()
			if err != nil {
				return nil, "", err
			}

			oauthToken := "unauthenticated-token"
			if keycloakURL != "null" {
				kc, err := connectors.NewKeycloakClient(keycloakURL, globalClientOpts.ClientId, globalClientOpts.ClientSecret)
				if err != nil {
					return nil, "", err
				}

				oauthToken, err = kc.ConnectAndGetToken()
				if err != nil {
					return nil, "", err
				}
			}
			mc.SetOAuthToken(oauthToken)
		}
		return mc, globalClientOpts.ServerAddr, nil
	}

	localConfig, err := config.ReadLocalConfig(globalClientOpts.ConfigPath)
	if err != nil {
		return nil, "", err
	}
	if localConfig == nil {
		return nil, "", errors.Wrapf(errors.KindUsage, "please login to perform this operation")
	}

	clientOpts := *globalClientOpts
	if clientOpts.Context == "" {
		clientOpts.Context = localConfig.CurrentContext
	}

	mc, err := connectors.NewClient(clientOpts)
	if err != nil {
		return nil, "", err
	}

	ctx, err := localConfig.ResolveContext(clientOpts.Context)
	if err != nil {
		return nil, "", errors.Wrap(errors.KindNotFound, err)
	}
	return mc, ctx.Server.Server, nil
}
