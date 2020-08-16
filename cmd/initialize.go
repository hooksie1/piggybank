package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// initializeCmd represents the initialize command
var initializeCmd = &cobra.Command{
	Use:   "initialize",
	Short: "Initialize the database",
	Run: func(cmd *cobra.Command, args []string) {
		initialize()
	},
}

func init() {
	rootCmd.AddCommand(initializeCmd)
}

func initialize() {
	host := viper.GetString("server")

	resp, err := http.Post(host+"/init/initialize", "application/json", nil)
	if err != nil {
		log.Printf("Error initializing database: %s", err)
		os.Exit(1)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response: %s", err)
		os.Exit(1)
	}

	fmt.Println(string(body))

}
