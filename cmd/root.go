package cmd

import (
	_ "embed"
	"encoding/json"
	"os"
	"slices"

	"github.com/janstuemmel/oli/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

//go:embed models.json
var modelsFile string

var (
	config  internal.Config
	rootCmd = &cobra.Command{
		Use:   "oli",
		Short: "openrouter cli",
		Run:   internal.Run(&config),
	}
)

type Model struct {
	Id           string `json:"id"`
	Name         string `json:"name"`
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
	viper.SetDefault("model", "openrouter/free")

	viper.BindEnv("model", "OLI_MODEL")
	rootCmd.RegisterFlagCompletionFunc("model", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return models, cobra.ShellCompDirectiveNoFileComp
	})

	// system prompt
	rootCmd.PersistentFlags().String("system", "", "system prompt")
	viper.BindPFlag("system", rootCmd.PersistentFlags().Lookup("system"))
	viper.SetDefault("system", "")
	viper.BindEnv("system", "OLI_SYSTEM")

	// completions
	rootCmd.AddCommand(&cobra.Command{
		Use:       "completion [bash|zsh|fish|powershell]",
		Short:     "Generate shell completion script",
		ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
		Run: func(cmd *cobra.Command, args []string) {
			switch args[0] {
			case "bash":
				_ = cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				_ = cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				_ = cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				_ = cmd.Root().GenPowerShellCompletion(os.Stdout)
			}
		},
	})
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
