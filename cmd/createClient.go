// create-client command is used to create a pxGrid client account based username & password
// this command requires to parameters: --server, --username
//--server: the DNS name or IP address of the cisco ISE server.
//--username: a username which will be used as the username for the client account

package cmd

import (
	"fmt"
	"github.com/Forcepoint/fp-bd-fuid-cisco-pxgrid/lib"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"time"
)

var createClientCmd = &cobra.Command{
	Use:   "create-client",
	Short: "Create a pxGrid client account",
	Long: `Creating Username & Password for Client Registration, Once the client is created the ISE
administrator needs to approve the created client account`,
	Run: func(cmd *cobra.Command, args []string) {
		createClient := lib.CreateClient{NodeName: viper.GetString("PXGRID_CLIENT_ACCOUNT_NAME")}
		controller, err := lib.GetController()
		if err != nil {
			logrus.Error(err)
			os.Exit(1)
		}
		iseClient, err := createClient.Create(controller)
		if err != nil {
			logrus.Error(err)
			os.Exit(1)
		}
		if DisplayProcess {
			logrus.Infof("Created  pxGrid client ccount with name '%s'", createClient.NodeName)
		}
		viper.Set("PXGRID_CLIENT_ACCOUNT_NAME", iseClient.UserName)
		viper.Set("PXGRID_CLIENT_ACCOUNT_PASSWORD", iseClient.Password)
		time.Sleep(3 * time.Second)
		accountActivate, err := createClient.AccountActivate(controller)
		if err != nil {
			logrus.Error(err)
			os.Exit(1)
		}
		if DisplayProcess {
			logrus.Infof("pxgrid client account '%s' has been activated", createClient.NodeName)
			logrus.Warnf("Contact your ISE Administrator to approve the created account")
		}
		fmt.Println("********* pxGrid API login Information are ********")
		fmt.Println("UserName: ", iseClient.UserName)
		fmt.Println("Password: ", iseClient.Password)
		fmt.Println()
		fmt.Println("save your ISE API login information in somewhere save.")
		if accountActivate.AccountState == "PENDING" {
			fmt.Printf("Created client account status is %s, the ISE Administrator needs to approve the created client account\n", accountActivate.AccountState)
		} else {
			fmt.Printf("Created client account status is %s\n", accountActivate.AccountState)
		}
	},
}

func init() {
	pxgridCmd.AddCommand(createClientCmd)
	createClientCmd.Flags().StringP("server", "", "", "ISE server DNS name or IP address")
	if err := createClientCmd.MarkFlagRequired("server"); err != nil {
		logrus.Fatal(err.Error())
	}

	createClientCmd.Flags().StringP("username", "u", "", "a username which will be used as pxGrid client name")
	if err := viper.BindPFlag("PXGRID_CLIENT_ACCOUNT_NAME",
		createClientCmd.Flags().Lookup("username")); err != nil {
		logrus.Fatal(err.Error())
	}
	if err := createClientCmd.MarkFlagRequired("username"); err != nil {
		logrus.Fatal(err.Error())
	}

}
