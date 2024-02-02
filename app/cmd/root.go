/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"
	"os"

	"github.com/inoth/promcollectr"
	_ "github.com/inoth/promcollectr/exporter/all"
	"github.com/inoth/toybox"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "promclt",
	Short: "prometheus采集器调度",
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		cfg := toybox.SetConfig{
			ConfDir:  "config",
			FileType: "toml",
		}
		if confDir != "" {
			cfg.ConfDir = confDir
		}

		tb := toybox.New(
			toybox.WithLoadConf(cfg),
			promcollectr.NewPromcollectrComponent(promcollectr.WithCfgPath(cfg.ConfDir+"/exporter")),
		)
		if err := tb.Run(); err != nil {
			log.Fatalf("%w\n", err)
			os.Exit(1)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var (
	confDir string // 配置文件地址
)

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.promcollectr.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	rootCmd.Flags().StringVarP(&confDir, "config", "c", "./config", "配置文件夹地址")
}
