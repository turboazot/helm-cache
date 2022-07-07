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
var homeDirectory string
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

		scanningInterval, err := cmd.Flags().GetDuration("scanningInterval")
		if err != nil {
			zap.L().Sugar().Fatalf("Fail to get scanning interval: %v", err)
		}

		helmClient, err := services.NewHelmClient(homeDirectory)
		if err != nil {
			zap.L().Sugar().Fatalf("Fail to initialize helm client: %v", err)
		}

		chartmuseumClient, err := services.NewChartmuseumClient(chartmuseumUrl, chartmuseumUsername, chartmuseumPassword)
		if err != nil {
			zap.L().Sugar().Fatalf("Fail to initialize chartmuseum client: %v", err)
		}

		c, err := services.NewCollector(helmClient, chartmuseumClient)
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
			time.Sleep(scanningInterval)
		}
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is path to config.yaml under helm-cache home directory)")
	rootCmd.PersistentFlags().StringVar(&homeDirectory, "homedir", "", "Home directory (default is $HOME/.helm-cache)")
	rootCmd.PersistentFlags().StringP("chartmuseumUrl", "c", "", "Chartmuseum URL")
	rootCmd.PersistentFlags().StringP("chartmuseumUsername", "u", "", "Chartmuseum username")
	rootCmd.PersistentFlags().StringP("chartmuseumPassword", "p", "", "Chartmuseum password")
	rootCmd.PersistentFlags().DurationP("scanningInterval", "s", 10*time.Second, "Interval between scanning helm release secrets")
	viper.BindPFlag("chartmuseumUrl", rootCmd.PersistentFlags().Lookup("chartmuseumUrl"))
	viper.BindPFlag("chartmuseumUsername", rootCmd.PersistentFlags().Lookup("chartmuseumUsername"))
	viper.BindPFlag("chartmuseumPassword", rootCmd.PersistentFlags().Lookup("chartmuseumPassword"))
	viper.BindPFlag("scanningInterval", rootCmd.PersistentFlags().Lookup("scanningInterval"))
}

func initConfig(cmd *cobra.Command) error {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	v := viper.New()

	if homeDirectory == "" {
		homeDirectory = fmt.Sprintf("%s/.helm-cache", userHomeDir)
	}

	if cfgFile != "" {
		v.SetConfigFile(cfgFile)
	} else {
		v.AddConfigPath(homeDirectory)
		v.SetConfigType("yaml")
		v.SetConfigName("config")
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
