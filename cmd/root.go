package cmd

import (
	"os"

	"github.com/janstuemmel/oli/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	config  internal.Config
	rootCmd = &cobra.Command{
		Use:   "oli",
		Short: "openrouter cli",
		Run:   internal.Run(&config),
	}
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// online
	rootCmd.PersistentFlags().BoolVarP(&config.Online, "online", "o", false, "use online model")

	// api key
	rootCmd.PersistentFlags().String("apikey", "", "api key")
	viper.BindPFlag("apikey", rootCmd.PersistentFlags().Lookup("apikey"))
	viper.SetDefault("apikey", "")
	viper.BindEnv("apikey", "OPENROUTER_API_KEY")

	// pipe
	rootCmd.PersistentFlags().String("pipe", "", "pipe bin")
	viper.BindPFlag("pipe", rootCmd.PersistentFlags().Lookup("pipe"))
	viper.SetDefault("pipe", "")
	viper.BindEnv("pipe", "OLI_PIPE")

	// model
	rootCmd.PersistentFlags().StringP("model", "m", "", "modle name")
	viper.BindPFlag("model", rootCmd.PersistentFlags().Lookup("model"))
	viper.SetDefault("model", "google/gemini-2.5-flash")
	viper.BindEnv("model", "OLI_MODEL")

	// system prompt
	rootCmd.PersistentFlags().String("system", "", "system prompt")
	viper.BindPFlag("system", rootCmd.PersistentFlags().Lookup("system"))
	viper.SetDefault("system", "")
	viper.BindEnv("system", "OLI_SYSTEM")
}

func initConfig() {
	home, err := os.UserHomeDir()
	cobra.CheckErr(err)

	viper.AddConfigPath(home)
	viper.AddConfigPath(".")

	viper.SetConfigType("yaml")
	viper.SetConfigName(".oli")
	viper.SetEnvPrefix("OLI")
	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	cobra.CheckErr(err)

	err = viper.Unmarshal(&config)
	cobra.CheckErr(err)
}
