// Copyright 2023 The frp Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sub

import (
	"fmt"
	"os"

	"github.com/samber/lo"
	"github.com/spf13/cobra"

	v1 "github.com/fatedier/frp/pkg/config/v1"
	"github.com/fatedier/frp/pkg/config/v1/validation"
	"github.com/fatedier/frp/pkg/consts"
)

var proxyTypes = []string{
	consts.TCPProxy,
	consts.UDPProxy,
	consts.TCPMuxProxy,
	consts.HTTPProxy,
	consts.HTTPSProxy,
	consts.STCPProxy,
	consts.SUDPProxy,
	consts.XTCPProxy,
}

var visitorTypes = []string{
	consts.STCPProxy,
	consts.SUDPProxy,
	consts.XTCPProxy,
}

func init() {
	for _, typ := range proxyTypes {
		c := v1.NewProxyConfigurerByType(typ)
		if c == nil {
			panic("proxy type: " + typ + " not support")
		}
		clientCfg := v1.ClientCommonConfig{}
		cmd := NewProxyCommand(typ, c, &clientCfg)
		RegisterClientCommonConfigFlags(cmd, &clientCfg)
		RegisterProxyFlags(cmd, c)

		// add sub command for visitor
		if lo.Contains(visitorTypes, typ) {
			vc := v1.NewVisitorConfigurerByType(typ)
			if vc == nil {
				panic("visitor type: " + typ + " not support")
			}
			visitorCmd := NewVisitorCommand(typ, vc, &clientCfg)
			RegisterVisitorFlags(visitorCmd, vc)
			cmd.AddCommand(visitorCmd)
		}
		rootCmd.AddCommand(cmd)
	}
}

func NewProxyCommand(name string, c v1.ProxyConfigurer, clientCfg *v1.ClientCommonConfig) *cobra.Command {
	return &cobra.Command{
		Use:   name,
		Short: fmt.Sprintf("Run frpc with a single %s proxy", name),
		Run: func(cmd *cobra.Command, args []string) {
			clientCfg.Complete()
			if _, err := validation.ValidateClientCommonConfig(clientCfg); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			c.Complete(clientCfg.User)
			c.GetBaseConfig().Type = name
			if err := validation.ValidateProxyConfigurerForClient(c); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			err := startService(clientCfg, []v1.ProxyConfigurer{c}, nil, "")
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}
}

func NewVisitorCommand(name string, c v1.VisitorConfigurer, clientCfg *v1.ClientCommonConfig) *cobra.Command {
	return &cobra.Command{
		Use:   "visitor",
		Short: fmt.Sprintf("Run frpc with a single %s visitor", name),
		Run: func(cmd *cobra.Command, args []string) {
			clientCfg.Complete()
			if _, err := validation.ValidateClientCommonConfig(clientCfg); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			c.Complete(clientCfg)
			c.GetBaseConfig().Type = name
			if err := validation.ValidateVisitorConfigurer(c); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			err := startService(clientCfg, nil, []v1.VisitorConfigurer{c}, "")
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}
}
