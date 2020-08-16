package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var application string
var user string
var pass string

// creatPassCmd represents the creatPass command
var createPassCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a credential (cannot use the manager account)",
	Run: func(cmd *cobra.Command, args []string) {
		createPass()
	},
}

func init() {
	createPassCmd.Flags().StringVarP(&pass, "pass", "p", "", "the user's password ")
	viper.BindPFlag("pass", createPassCmd.Flags().Lookup("pass"))
}

func createPass() {
	host := viper.GetString("server")
	piggyUser := viper.GetString("piggy_user")
	piggyPass := viper.GetString("piggy_pass")

	if application == "" || user == "" || pass == "" {
		log.Println("You must supply an application, username, and password")
		os.Exit(1)
	}

	appUser := AppUser{
		Application: application,
		Username:    user,
		Password:    pass,
	}

	data, err := json.Marshal(appUser)
	if err != nil {
		log.Printf("Error marshaling data: %s", err)
		os.Exit(1)
	}

	client := http.Client{}

	req, err := NewRequest(
		SetMethod("POST"),
		SetURL(host+"/api/password"),
		SetBody(bytes.NewBuffer(data)),
		SetCredentials(piggyUser, piggyPass),
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

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response: %s", err)
		os.Exit(1)
	}

	fmt.Println("Credential created")

}
