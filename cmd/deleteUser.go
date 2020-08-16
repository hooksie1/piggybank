package cmd

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// deleteUserCmd represents the deleteUser command
var deleteUserCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an existing user",
	Run: func(cmd *cobra.Command, args []string) {
		deleteUser()
	},
}

func init() {
}

func deleteUser() {
	host := viper.GetString("server")
	manPass := viper.GetString("manager_pass")

	client := http.Client{}

	req, err := NewRequest(
		SetMethod("DELETE"),
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

	err = checkResponse(resp)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("User deleted\n")
}
