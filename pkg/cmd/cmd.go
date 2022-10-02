package cmd

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/klog/v2"
	"os"
	"tx/pkg/cognito"
	"tx/pkg/logs"
	"tx/pkg/util"
)

var Port string

var configFile string

var defaultConfigFilename = ".tx.yaml"

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

	EntrypointCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file (default is ~/.tx.yml)")
	EntrypointCmd.PersistentFlags().StringVarP(&Port, "port", "p", "8080", "which port to listen on localhost")
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	viper.SetConfigType("yaml")
	// viper will prefer flags from command line rather than file
	if err := viper.BindPFlags(EntrypointCmd.Flags()); err != nil {
		klog.ErrorS(err, "Failed bind flags")
	}
	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		home, err := homedir.Dir()
		util.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigName(defaultConfigFilename)
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		klog.V(50).InfoS("Using config", "file", viper.ConfigFileUsed())
	} else {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {

			filepath, err := homedir.Expand(fmt.Sprintf("~/%s", defaultConfigFilename))
			util.CheckErr(err)
			util.CheckErr(viper.WriteConfigAs(filepath))
			klog.V(50).InfoS("wrote new config file", "filepath", filepath)

		}
		klog.V(50).InfoS(err.Error())
	}

	if !EntrypointCmd.Flags().Changed("v") {
		util.CheckErr(EntrypointCmd.Flags().Set("v", viper.GetString("v")))
	}

	err := viper.WriteConfigAs(viper.ConfigFileUsed())
	if err != nil {
		klog.V(50).InfoS("Using err", "err", err)
	}

	klog.V(50).InfoS("Using config file", "ConfigFileUsed", viper.ConfigFileUsed())

}
