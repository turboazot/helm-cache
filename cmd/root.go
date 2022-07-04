package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/turboazot/helm-cache/pkg/services"
)

var (
	cfgFile string

	rootCmd = &cobra.Command{
		Use:   "helm-cache",
		Short: "Helm chart cache daemon",
		Long:  "A cache daemon that caching Helm v3 charts in Kubernetes cluster",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initConfig(cmd)
		},
		Run: func(cmd *cobra.Command, args []string) {
			chartmuseumUrl, err := cmd.Flags().GetString("chartmuseumUrl")
			if err != nil {
				panic(err.Error())
			}
			chartmuseumUsername, err := cmd.Flags().GetString("chartmuseumUsername")
			if err != nil {
				panic(err.Error())
			}
			chartmuseumPassword, err := cmd.Flags().GetString("chartmuseumPassword")
			if err != nil {
				panic(err.Error())
			}

			scanningInterval, err := cmd.Flags().GetInt("scanningInterval")
			if err != nil {
				panic(err.Error())
			}

			c := services.NewCollector(chartmuseumUrl, chartmuseumUsername, chartmuseumPassword)
			for {
				fmt.Println("Checking all helm secrets...")
				c.CheckAllSecrets()
				fmt.Println("Checking finished!")
				time.Sleep(time.Millisecond * time.Duration(scanningInterval*1000))
			}
		},
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.helm-cache.yaml)")
	rootCmd.PersistentFlags().StringP("chartmuseumUrl", "c", "", "Chartmuseum URL")
	rootCmd.MarkPersistentFlagRequired("chartmuseumUrl")
	rootCmd.PersistentFlags().StringP("chartmuseumUsername", "u", "", "Chartmuseum username")
	rootCmd.MarkPersistentFlagRequired("chartmuseumUsername")
	rootCmd.PersistentFlags().StringP("chartmuseumPassword", "p", "", "Chartmuseum password")
	rootCmd.MarkPersistentFlagRequired("scanningInterval")
	rootCmd.PersistentFlags().IntP("scanningInterval", "s", 10, "Interval between scanning helm release secrets")
	viper.BindPFlag("chartmuseumUrl", rootCmd.PersistentFlags().Lookup("chartmuseumUrl"))
	viper.BindPFlag("chartmuseumUsername", rootCmd.PersistentFlags().Lookup("chartmuseumUsername"))
	viper.BindPFlag("chartmuseumPassword", rootCmd.PersistentFlags().Lookup("chartmuseumPassword"))
	viper.BindPFlag("scanningInterval", rootCmd.PersistentFlags().Lookup("scanningInterval"))
}

func initConfig(cmd *cobra.Command) error {
	v := viper.New()
	if cfgFile != "" {
		// Use config file from the flag.
		v.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		v.AddConfigPath(home)
		v.SetConfigType("yaml")
		v.SetConfigName(".helm-cache")
	}

	v.AutomaticEnv()

	if err := v.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", v.ConfigFileUsed())
	}

	bindFlags(cmd, v)

	return nil
}

func bindFlags(cmd *cobra.Command, v *viper.Viper) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if !f.Changed && v.IsSet(f.Name) {
			val := v.Get(f.Name)
			cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
		}
	})
}
