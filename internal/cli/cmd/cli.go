/*
Copyright ApeCloud Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cliflag "k8s.io/component-base/cli/flag"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/templates"

	"github.com/apecloud/kubeblocks/internal/cli/cmd/backupconfig"
	"github.com/apecloud/kubeblocks/internal/cli/cmd/bench"
	"github.com/apecloud/kubeblocks/internal/cli/cmd/cluster"
	"github.com/apecloud/kubeblocks/internal/cli/cmd/kubeblocks"
	"github.com/apecloud/kubeblocks/internal/cli/cmd/options"
	"github.com/apecloud/kubeblocks/internal/cli/cmd/playground"
	"github.com/apecloud/kubeblocks/internal/cli/cmd/version"
)

var cfgFile string

func NewCliCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kbcli",
		Short: "KubeBlocks CLI",
		Long: `
=============================================
 __    __ _______   ______  __       ______ 
|  \  /  \       \ /      \|  \     |      \
| ▓▓ /  ▓▓ ▓▓▓▓▓▓▓\  ▓▓▓▓▓▓\ ▓▓      \▓▓▓▓▓▓
| ▓▓/  ▓▓| ▓▓__/ ▓▓ ▓▓   \▓▓ ▓▓       | ▓▓  
| ▓▓  ▓▓ | ▓▓    ▓▓ ▓▓     | ▓▓       | ▓▓  
| ▓▓▓▓▓\ | ▓▓▓▓▓▓▓\ ▓▓   __| ▓▓       | ▓▓  
| ▓▓ \▓▓\| ▓▓__/ ▓▓ ▓▓__/  \ ▓▓_____ _| ▓▓_ 
| ▓▓  \▓▓\ ▓▓    ▓▓\▓▓    ▓▓ ▓▓     \   ▓▓ \
 \▓▓   \▓▓\▓▓▓▓▓▓▓  \▓▓▓▓▓▓ \▓▓▓▓▓▓▓▓\▓▓▓▓▓▓

=============================================
A Command Line Interface for KubeBlocks`,

		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	// From this point and forward we get warnings on flags that contain "_" separators
	// when adding them with hyphen instead of the original name.
	cmd.SetGlobalNormalizationFunc(cliflag.WarnWordSepNormalizeFunc)

	flags := cmd.PersistentFlags()

	// add kubernetes flags like kubectl
	kubeConfigFlags := genericclioptions.NewConfigFlags(true)
	kubeConfigFlags.AddFlags(flags)
	matchVersionKubeConfigFlags := cmdutil.NewMatchVersionFlags(kubeConfigFlags)
	matchVersionKubeConfigFlags.AddFlags(flags)

	f := cmdutil.NewFactory(matchVersionKubeConfigFlags)
	ioStreams := genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}

	// Add subcommands
	cmd.AddCommand(
		playground.NewPlaygroundCmd(ioStreams),
		kubeblocks.NewKubeBlocksCmd(f, ioStreams),
		cluster.NewClusterCmd(f, ioStreams),
		bench.NewBenchCmd(),
		options.NewCmdOptions(ioStreams.Out),
		version.NewVersionCmd(f),
		backupconfig.NewBackupConfigCmd(f, ioStreams),
	)

	filters := []string{"options"}
	templates.ActsAsRootCommand(cmd, filters, []templates.CommandGroup{}...)

	cobra.OnInitialize(initConfig)
	return cmd
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".kbcli" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".kbcli")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}