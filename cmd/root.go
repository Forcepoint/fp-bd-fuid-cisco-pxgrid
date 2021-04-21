package cmd

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io"
	"os"
)

var (
	cfgFile        string
	DisplayProcess bool
)
var rootCmd = &cobra.Command{
	Use:   "fuid-ise",
	Short: "Integration between Forcepoint User ID service and Cisco ISE",
	Long: `Integration between Forcepoint User Id Service and Cisco ISE.
Consume to pxGrid service to watch session events`,
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	viper.SetDefault("INTERNAL_LOGS_FILE", "")
	viper.SetDefault("SESSION_LATEST_TIMESTAMP_PATH", "")
	//ISE configs
	viper.SetDefault("PXGRID_CLIENT_ACCOUNT_NAME", "")
	viper.SetDefault("PXGRID_CLIENT_ACCOUNT_PASSWORD", "")
	viper.SetDefault("PXGRID_HOST_ADDRESS", "")
	viper.SetDefault("ISE_PORT", 8910)
	//FUID configs
	viper.SetDefault("FUID_IP_ADDRESS", "")
	viper.SetDefault("FUID_API_USERNAME", "")
	viper.SetDefault("FUID_API_PASSWORD", "")
	viper.SetDefault("FUID_PORT", 5000)
	//AD configs
	viper.SetDefault("AD_IP_ADDRESS", "")
	viper.SetDefault("AD_PORT", 389)
	viper.SetDefault("AD_ADMIN_USERNAME", "")
	viper.SetDefault("AD_ADMIN_PASSWORD", "")
	viper.SetDefault("AD_DOMAIN_NAME", "i")
	viper.SetDefault("LDAP_TIMEOUT", 10)
	viper.SetDefault("LDAP_PAGES", 500)
	//other Config
	viper.SetDefault("SESSION_LISTENER_INTERVAL_TIME", 2)
	viper.SetDefault("SAVE_LOGS", false)
	viper.SetDefault("DISPLAY_INFO", false)

	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "YAML config file ")
	rootCmd.PersistentFlags().BoolP("save-logs", "s", true, "Save logs")
	if err := viper.BindPFlag("SAVE_LOGS",
		rootCmd.PersistentFlags().Lookup("save-logs")); err != nil {
		logrus.Fatal(err.Error())
	}
	rootCmd.PersistentFlags().BoolP("info", "i", true, "Display the progress information on the console")
	if err := viper.BindPFlag("DISPLAY_INFO",
		rootCmd.PersistentFlags().Lookup("info")); err != nil {
		logrus.Fatal(err.Error())
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	}
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		_, _ = fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}

	if viper.GetBool("SAVE_LOGS") {
		errorLogFile, err := os.OpenFile(viper.GetString("INTERNAL_LOGS_FILE"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			logrus.Fatalf("Cannot create or open the Error logs file: %s", viper.GetString("INTERNAL_LOGS_FILE"))
		}
		mw := io.MultiWriter(os.Stdout, errorLogFile)
		logrus.SetOutput(mw)
		logrus.SetFormatter(&logrus.TextFormatter{})
	}
	DisplayProcess = viper.GetBool("DISPLAY_INFO")
}