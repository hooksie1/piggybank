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
	Run:   deleteCred,
}

func init() {
}

func deleteCred(cmd *cobra.Command, args []string) {
	host := viper.GetString("server")
	piggyUser, piggyPass := viper.GetString("piggy_user"), viper.GetString("piggy_pass")
	application := viper.GetString("app")
	user := viper.GetString("appUser")

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
