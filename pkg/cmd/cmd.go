package cmd

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/klog/v2"
	"os"
	"token-exchange-cli/pkg/cognito"
	"token-exchange-cli/pkg/logs"
	"token-exchange-cli/pkg/util"
)

var Port string

var cfgFile string

var EntrypointCmd = &cobra.Command{
	Use: "tx",

	Short: "token-exchange - a simple CLI to authenticate via browser",
	Long: `token-exchange - a simple CLI to authenticate via browser
launches browser url with redirect url as parameter to return data
`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
	},
	Run: func(cmd *cobra.Command, args []string) {
		util.CheckErr(cmd.Help())
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		// save config on successful run
		util.CheckErr(viper.WriteConfigAs(viper.ConfigFileUsed()))
	},
}

func Execute() {
	if err := EntrypointCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "token-exchange: failed to start: '%s'", err)
		os.Exit(1)
	}
}

func init() {
	logs.AddFlags(EntrypointCmd.PersistentFlags())

	EntrypointCmd.AddCommand(cognito.Cmd)

	EntrypointCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.tx-config.yml)")
	EntrypointCmd.PersistentFlags().StringVarP(&Port, "port", "p", "8080", "overwrite ports to use, can add multiple")
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	// todo: defaults
	//viper.SetDefault("ContentDir", "content")

	// viper will prefer flags from command line rather than file
	if err := viper.BindPFlags(EntrypointCmd.Flags()); err != nil {
		klog.ErrorS(err, "Failed bind flags")
	}
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		util.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigName(".tx-config")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		klog.V(50).InfoS("Using config", "file", viper.ConfigFileUsed())
	} else {
		klog.V(50).InfoS(err.Error())
	}

	if !EntrypointCmd.Flags().Changed("v") {
		util.CheckErr(EntrypointCmd.Flags().Set("v", viper.GetString("v")))
	}
	klog.V(1).InfoS("verbosity", "v", viper.GetString("v"))
	err := viper.WriteConfigAs(viper.ConfigFileUsed())
	if err != nil {
		klog.V(50).InfoS("Using err", "err", err)
	}

	klog.V(50).InfoS("Using config", "ConfigFileUsed", viper.ConfigFileUsed())

}