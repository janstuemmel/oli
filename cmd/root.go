package cmd

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"slices"

	"github.com/janstuemmel/oli/internal/app"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Model struct {
	Id   string `json:"id"`
	Name string `json:"name"`

	Architecture struct {
		InputModalities  []string `json:"input_modalities"`
		OutputModalities []string `json:"output_modalities"`
	} `json:"architecture"`

	Pricing struct {
		Prompt            string `json:"prompt"`
		Completion        string `json:"completion"`
		Request           string `json:"request"`
		Image             string `json:"image"`
		WebSearch         string `json:"web_search"`
		InternalReasoning string `json:"internal_reasoning"`
	} `json:"pricing"`
}

type modelJson struct {
	Data []Model `json:"data"`
}

//go:embed models.json
var modelsFile string

var (
	config  app.Config
	rootCmd = &cobra.Command{
		Use:   "oli",
		Short: "openrouter cli",
		Run:   app.Run(&config),
		Args:  cobra.MinimumNArgs(0),
	}
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	var modelJson modelJson
	var models []string
	err := json.Unmarshal([]byte(modelsFile), &modelJson)
	cobra.CheckErr(err)
	for _, model := range modelJson.Data {
		if slices.Contains(model.Architecture.InputModalities, "text") && slices.Contains(model.Architecture.OutputModalities, "text") {
			models = append(models, model.Id)
		}
	}

	cobra.OnInitialize(initConfig)

	// online
	rootCmd.Flags().BoolVarP(&config.Online, "online", "o", false, "use online model")

	// api key
	rootCmd.Flags().String("apikey", "", "api key")
	viper.BindPFlag("apikey", rootCmd.Flags().Lookup("apikey"))
	viper.SetDefault("apikey", "")
	viper.BindEnv("apikey", "OPENROUTER_API_KEY")

	// pipe
	rootCmd.Flags().String("pipe", "", "pipe bin")
	viper.BindPFlag("pipe", rootCmd.Flags().Lookup("pipe"))
	viper.SetDefault("pipe", "")
	viper.BindEnv("pipe", "OLI_PIPE")

	// model
	rootCmd.Flags().StringP("model", "m", "", "model name")
	viper.BindPFlag("model", rootCmd.Flags().Lookup("model"))
	viper.SetDefault("model", "openrouter/free")

	viper.BindEnv("model", "OLI_MODEL")
	rootCmd.RegisterFlagCompletionFunc("model", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return models, cobra.ShellCompDirectiveNoFileComp
	})

	// system prompt
	rootCmd.Flags().String("system", "", "system prompt")
	viper.BindPFlag("system", rootCmd.Flags().Lookup("system"))
	viper.SetDefault("system", "")
	viper.BindEnv("system", "OLI_SYSTEM")

	rootCmd.AddCommand(completion)
}

func initConfig() {
	home, err := os.UserHomeDir()
	cobra.CheckErr(err)

	viper.AddConfigPath(fmt.Sprintf("%s/.config/oli", home))
	viper.AddConfigPath(".")

	viper.SetConfigType("yaml")
	viper.SetConfigName("config.yaml")

	viper.ReadInConfig()
	viper.Unmarshal(&config)
}
