package cmd

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/turboazot/helm-cache/pkg/services"
	"go.uber.org/zap"
)

func runRootCommand(cmd *cobra.Command, args []string) {
	var err error
	var kubeconfigPath string

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

	homeDirectory, err := cmd.Flags().GetString("homeDirectory")
	if err != nil {
		zap.L().Sugar().Fatalf("Fail to get home directory value: %v", err)
	}

	inclusterConfig, err := cmd.Flags().GetBool("inclusterConfig")
	if err != nil {
		zap.L().Sugar().Fatalf("Fail to get in-cluster config value: %v", err)
	}
	if inclusterConfig {
		kubeconfigPath = ""
	} else {
		kubeconfigPath, err = cmd.Flags().GetString("kubeconfigPath")
		if err != nil {
			zap.L().Sugar().Fatalf("Fail to get kubeconfig path config value: %v", err)
		}
	}

	helmClient, err := services.NewHelmClient(homeDirectory)
	if err != nil {
		zap.L().Sugar().Fatalf("Fail to initialize helm client: %v", err)
	}

	chartmuseumClient, err := services.NewChartmuseumClient(chartmuseumUrl, chartmuseumUsername, chartmuseumPassword)
	if err != nil {
		zap.L().Sugar().Fatalf("Fail to initialize chartmuseum client: %v", err)
	}

	c, err := services.NewCollector(helmClient, chartmuseumClient, kubeconfigPath)
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
}

func Execute() error {
	var rootCmd = &cobra.Command{
		Use:               "helm-cache",
		Short:             "Helm chart cache daemon",
		Long:              "A cache daemon that caching Helm v3 charts in Kubernetes cluster",
		PersistentPreRunE: initCommand,
		Run:               runRootCommand,
	}
	rootCmd.PersistentFlags().StringP("configFile", "f", "", "config file (default is path to config.yaml under helm-cache home directory)")
	rootCmd.PersistentFlags().BoolP("inclusterConfig", "i", false, "in-cluster config")
	rootCmd.PersistentFlags().StringP("kubeconfigPath", "k", "", "kubeconfig path (default is $HOME/.kube/config)")
	rootCmd.PersistentFlags().StringP("homeDirectory", "d", "", "Home directory (default is $HOME/.helm-cache)")
	rootCmd.PersistentFlags().StringP("chartmuseumUrl", "c", "", "Chartmuseum URL")
	rootCmd.PersistentFlags().StringP("chartmuseumUsername", "u", "", "Chartmuseum username")
	rootCmd.PersistentFlags().StringP("chartmuseumPassword", "p", "", "Chartmuseum password")
	rootCmd.PersistentFlags().DurationP("scanningInterval", "s", 10*time.Second, "Interval between scanning helm release secrets")
	viper.BindPFlag("chartmuseumUrl", rootCmd.PersistentFlags().Lookup("chartmuseumUrl"))
	viper.BindPFlag("chartmuseumUsername", rootCmd.PersistentFlags().Lookup("chartmuseumUsername"))
	viper.BindPFlag("chartmuseumPassword", rootCmd.PersistentFlags().Lookup("chartmuseumPassword"))
	viper.BindPFlag("scanningInterval", rootCmd.PersistentFlags().Lookup("scanningInterval"))

	return rootCmd.Execute()
}

func initCommand(cmd *cobra.Command, args []string) error {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	v := viper.New()

	homeDirectory, err := cmd.Flags().GetString("homeDirectory")
	if err != nil {
		return err
	}
	configFile, err := cmd.Flags().GetString("configFile")
	if err != nil {
		return err
	}

	kubeconfigPath, err := cmd.Flags().GetString("kubeconfigPath")
	if err != nil {
		return err
	}

	inclusterConfig, err := cmd.Flags().GetBool("inclusterConfig")
	if err != nil {
		return err
	}

	if kubeconfigPath == "" {
		defaultKubeconfigPath := fmt.Sprintf("%s/.kube/config", userHomeDir)

		_, err := os.Stat(defaultKubeconfigPath)

		if !errors.Is(err, os.ErrNotExist) && !inclusterConfig {
			cmd.Flags().Set("kubeconfigPath", defaultKubeconfigPath)
		}
	}

	if homeDirectory == "" {
		homeDirectory = fmt.Sprintf("%s/.helm-cache", userHomeDir)
	}

	err = cmd.Flags().Set("homeDirectory", homeDirectory)
	if err != nil {
		return err
	}

	if configFile != "" {
		v.SetConfigFile(configFile)
	} else {
		v.AddConfigPath(homeDirectory)
		v.SetConfigType("yaml")
		v.SetConfigName("config")
	}

	v.AutomaticEnv()

	if err := v.ReadInConfig(); err == nil {
		zap.L().Sugar().Infof("Using config file: %s", v.ConfigFileUsed())
	}

	return bindFlags(cmd, v)
}

func bindFlags(cmd *cobra.Command, v *viper.Viper) error {
	var err error = nil
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if !f.Changed && v.IsSet(f.Name) {
			val := v.Get(f.Name)
			err = cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
		}
	})

	return err
}
