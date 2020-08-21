package cmd

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gitlab.com/hooksie1/piggybank/server"
)

// lookupCmd represents the lookup command
var lookupCmd = &cobra.Command{
	Use:   "lookup",
	Short: "Lookup a credential",
	Run:   lookupCred,
}

func init() {
	lookupCmd.Flags().StringP("app", "a", "", "the application to look up")
	lookupCmd.Flags().StringP("user", "u", "", "the user in the application to retrieve")
	viper.BindPFlag("credApp", lookupCmd.Flags().Lookup("app"))
	viper.BindPFlag("credUser", lookupCmd.Flags().Lookup("user"))
}

func lookupCred(cmd *cobra.Command, args []string) {
	host := viper.GetString("server")
	piggyUser := viper.GetString("piggy_user")
	piggyPass := viper.GetString("piggy_pass")
	credUser := viper.GetString("credUser")
	credApp := viper.GetString("credApp")

	app := &server.Application{}

	if credApp == "" || credUser == "" {
		log.Println("You must supply an application and username")
		os.Exit(1)
	}

	client := http.Client{}

	query := map[string]string{
		"application": credApp,
		"username":    credUser,
	}

	req, err := NewRequest(
		SetMethod("GET"),
		SetURL(host+"/api/password"),
		SetCredentials(piggyUser, piggyPass),
		SetQuery(query),
	)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending request to server: %s", err)
		os.Exit(1)
	}

	err = checkResponse(resp)
	if err != nil {
		fmt.Println(err)
		return
	}

	if err := PrintData(app, resp.Body); err != nil {
		fmt.Printf("Error printing data: %s", err)
		os.Exit(1)
	}

}
