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

// createUserCmd represents the createUser command
var createUserCmd = &cobra.Command{
	Use:   "create",
	Short: "Creates a system user (must be used by the manager account)",
	Run: func(cmd *cobra.Command, args []string) {
		createUser()
	},
}

func init() {

}

func createUser() {
	host := viper.GetString("server")
	manPass := viper.GetString("manager_pass")
	sysUser := viper.GetString("sysUser")

	user, client := &server.User{}, http.Client{}

	req, err := NewRequest(
		SetMethod("POST"),
		SetURL(host+"/api/user/"+sysUser),
		SetCredentials("manager", manPass),
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

	if err = checkResponse(resp); err != nil {
		fmt.Println(err)
		return
	}

	if err := PrintData(user, resp.Body); err != nil {
		fmt.Printf("Error printing data: %s", err)
		os.Exit(1)
	}
}
