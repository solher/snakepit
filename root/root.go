package root

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var Viper = viper.New()

var (
	cfgFile string
)

// This represents the base command when called without any subcommands
var Cmd = &cobra.Command{
	Use:   "app",
	Short: "An amazing web service.",
}

func init() {
	cobra.OnInitialize(initConfig)
	// Here you will define your flags and configuration settings
	// Cobra supports Persistent Flags which if defined here will be global for your application

	Cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (defaults are $HOME/config.yaml and ./config.yaml)")
	Viper.BindPFlag("config", Cmd.PersistentFlags().Lookup("config"))
}

// Read in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		Viper.SetConfigFile(cfgFile)
	}

	Viper.SetConfigName("config") // name of config file (without extension)
	Viper.AddConfigPath("$HOME")  // adding home directory as first search path
	Viper.AddConfigPath("./")     // adding local directory as second search path
	Viper.AutomaticEnv()          // read in environment variables that match

	// If a config file is found, read it in.
	if err := Viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", Viper.ConfigFileUsed())
	}
}
