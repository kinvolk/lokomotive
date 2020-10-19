// Copyright 2020 The Lokomotive Authors
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

// Package cmd has code for all the subcommands of lokoctl.
package cmd

import (
	"os"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var RootCmd = &cobra.Command{
	Use:   "lokoctl",
	Short: "Manage Lokomotive clusters",
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

var kubeconfigFlag string

func addKubeconfigFileFlag(pf *flag.FlagSet) {
	pf.StringVar(
		&kubeconfigFlag,
		"kubeconfig-file",
		"", // Special empty default, use getKubeconfig()
		"Path to a kubeconfig file. If empty, the following precedence order is used:\n"+
			"  1. Cluster asset dir when a lokocfg file is present in the current directory.\n"+
			"  2. KUBECONFIG environment variable.\n"+
			"  3. ~/.kube/config file.")

	if err := viper.BindPFlag(kubeconfigFlag, pf.Lookup(kubeconfigFlag)); err != nil {
		panic("failed registering kubeconfig flag")
	}
}

func init() { //nolint:gochecknoinits
	cobra.OnInitialize(cobraInit)

	RootCmd.DisableAutoGenTag = true

	RootCmd.PersistentFlags().String("lokocfg", "./", "Path to lokocfg directory or file")
	viper.BindPFlag("lokocfg", RootCmd.PersistentFlags().Lookup("lokocfg"))
	RootCmd.PersistentFlags().String("lokocfg-vars", "./lokocfg.vars", "Path to lokocfg.vars file")
	viper.BindPFlag("lokocfg-vars", RootCmd.PersistentFlags().Lookup("lokocfg-vars"))
}

func cobraInit() {
	viper.AutomaticEnv()
}
