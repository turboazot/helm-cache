package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/turboazot/helm-cache/pkg/services"
	"go.uber.org/zap"
)

var cfgFile string
var rootCmd = &cobra.Command{
	Use:   "helm-cache",
	Short: "Helm chart cache daemon",
	Long:  "A cache daemon that caching Helm v3 charts in Kubernetes cluster",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return initConfig(cmd)
	},
	Run: func(cmd *cobra.Command, args []string) {
		chartmuseumUrl, err := cmd.Flags().GetString("chartmuseumUrl")
		if err != nil {
			zap.L().Sugar().Fatalf("Fail to get chartmuseum url: %v", err)
		}
		chartmuseumUsername, err := cmd.Flags().GetString("chartmuseumUsername")
		if err != nil {
			zap.L().Sugar().Fatalf("Fail to get chartmuseum username: %v", err)
		}
		chartmuseumPassword, err := cmd.Flags().GetString("chartmuseumPassword")
		if err != nil {
			zap.L().Sugar().Fatalf("Fail to get chartmuseum password: %v", err)
		}

		scanningInterval, err := cmd.Flags().GetInt("scanningInterval")
		if err != nil {
			zap.L().Sugar().Fatalf("Fail to get scanning interval: %v", err)
		}

		c, err := services.NewCollector(chartmuseumUrl, chartmuseumUsername, chartmuseumPassword)
		if err != nil {
			zap.L().Sugar().Fatalf("Fail to initialize collector: %v", err)
		}

		for {
			zap.L().Sugar().Info("Checking all helm secrets...")
			err = c.CheckAllSecrets()
			if err != nil {
				zap.L().Sugar().Fatalf("Fail to check helm secrets: %v", err)
			}
			zap.L().Sugar().Info("Checking finished!")
			time.Sleep(time.Millisecond * time.Duration(scanningInterval*1000))
		}
	},
}

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
		zap.L().Sugar().Infof("Using config file: %s", v.ConfigFileUsed())
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
