// the consumer command subscribe for cisco ISE pxGrid service.
//watches the ISE session events and takes required actions to add or remove users ip address using FUID API

package cmd

import (
	"github.com/Forcepoint/fp-bd-fuid-cisco-pxgrid/lib"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"time"
)

// consumerRestCmd represents the consumer command
var consumerRestCmd = &cobra.Command{
	Use:   "consumer",
	Short: "subscribe for sessions events using the REST API",
	Long:  `watch session events and take action for AUTHENTICATED and DISCONNECT events`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := lib.ValidateUsernamePassword(); err != nil {
			logrus.Error(err)
			logrus.Exit(1)
		}
		fuidController, err := lib.NewFUIDController()
		if err != nil {
			logrus.Error(err)
			logrus.Exit(1)
		}
		createClient := lib.CreateClient{NodeName: viper.GetString("PXGRID_CLIENT_ACCOUNT_NAME")}
		controller, err := lib.GetController()
		if err != nil {
			logrus.Error(err)
			os.Exit(1)
		}
		accountActivate, err := createClient.AccountActivate(controller)
		if err != nil {
			logrus.Error(err)
			logrus.Exit(1)
		}
		if accountActivate.AccountState != lib.Enabled {
			logrus.Errorf("the status of the client account is %s, please contact your Cisco ISE Administrator to aprove or enable it", accountActivate.AccountState)
			logrus.Exit(1)
		}
		if DisplayProcess {
			logrus.Infof("PxGrid API Client Account %s is Activated and Enabled", createClient.NodeName)
		}
		//do service lookup
		serviceLookupOutput, err := lib.ServiceLookupRequest(lib.ServiceLookupSessions, controller)
		if err != nil {
			logrus.Error(err)
			logrus.Exit(1)
		}
		if DisplayProcess {
			logrus.Infof("service lookup: found %d services", len(serviceLookupOutput.Services))
		}
		restBaseUrl, nodeName, err := lib.GetSessionRestUrl(serviceLookupOutput.Services)
		if err != nil {
			logrus.Error(err)
			logrus.Exit(1)
		}
		if DisplayProcess {
			logrus.Infof("Using restBaseUrl: %s", restBaseUrl)
		}
		accessSecretOutput, err := lib.AccessSecret(nodeName, controller)
		if err != nil {
			logrus.Error(err)
			logrus.Exit(1)
		}
		lib.SetupCloseHandler()
		for {
			if err := lib.SessionListener(accessSecretOutput.Secret, restBaseUrl, viper.GetString("SESSION_LATEST_TIMESTAMP_PATH"), controller, fuidController, DisplayProcess); err != nil {
				logrus.Error(err)
				logrus.Exit(1)
			}
			time.Sleep(time.Duration(viper.GetInt("SESSION_LISTENER_INTERVAL_TIME")) * time.Second)
		}
	},
}

func init() {
	pxgridCmd.AddCommand(consumerRestCmd)
}
