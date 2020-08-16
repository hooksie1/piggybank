package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type MasterPass struct {
	Password string `json:"master_password"`
}

var unlockPass string

// unlockCmd represents the unlock command
var unlockCmd = &cobra.Command{
	Use:   "unlock",
	Short: "Unlock the database with the master password",
	Run: func(cmd *cobra.Command, args []string) {
		unlock()
	},
}

func init() {
	rootCmd.AddCommand(unlockCmd)
	unlockCmd.Flags().StringVarP(&unlockPass, "password", "p", "", "the unlock password to use")

}

func unlock() {
	host := viper.GetString("server")

	pass := MasterPass{
		Password: unlockPass,
	}

	data, err := json.Marshal(pass)
	if err != nil {
		log.Printf("Error marshaling data: %s", err)
		os.Exit(1)
	}

	client := http.Client{}

	req, err := NewRequest(
		SetMethod("POST"),
		SetURL(host+"/init/unlock"),
		SetBody(bytes.NewBuffer(data)),
	)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	resp, err := client.Do(req)

	if err := checkResponse(resp); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Database unlocked")

}
