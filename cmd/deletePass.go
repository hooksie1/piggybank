/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// deletePassCmd represents the deletePass command
var deletePassCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a credential from an application",
	Run: func(cmd *cobra.Command, args []string) {
		deleteCred()
	},
}

func init() {
}

func deleteCred() {
	host := viper.GetString("server")
	piggyUser, piggyPass := viper.GetString("piggy_user"), viper.GetString("piggy_pass")

	if application == "" || user == "" {
		log.Println("You must supply an application and username")
		os.Exit(1)
	}

	client := http.Client{}

	query := map[string]string{
		"application": application,
		"username":    user,
	}

	req, err := NewRequest(
		SetMethod("DELETE"),
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

	fmt.Println("Credential deleted")
}
