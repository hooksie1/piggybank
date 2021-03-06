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

// backupCmd represents the backup command
var (
	backupCmd = &cobra.Command{
		Use:   "backup",
		Short: "Backup the database (must use the manager account)",
		Run:   backup,
	}
)

func init() {
	rootCmd.AddCommand(backupCmd)

	backupCmd.Flags().BoolP("local", "l", false, "Writes the backup locally")
	backupCmd.Flags().StringP("path", "p", "", "Path for local backup")
	viper.BindPFlag("backLocal", backupCmd.Flags().Lookup("local"))
	viper.BindPFlag("backPath", backupCmd.Flags().Lookup("path"))
}

func backup(cmd *cobra.Command, args []string) {
	host := viper.GetString("server")
	managerPass := viper.GetString("manager_pass")
	backupType := "local"

	local := viper.GetBool("backLocal")
	path := viper.GetString("backPath")

	if local && path == "" {
		log.Println("You must supply a path for the local backup")
		os.Exit(1)
	}

	if local {
		backupType = "http"
	}

	client := http.Client{}

	query := map[string]string{
		"type": backupType,
	}

	req, err := NewRequest(
		SetMethod("POST"),
		SetURL(host+"/api/backup"),
		SetCredentials("manager", managerPass),
		SetQuery(query),
	)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending reqeust to server: %s", err)
		os.Exit(1)
	}

	err = checkResponse(resp)
	if err != nil {
		fmt.Println(err)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response: %s", err)
		os.Exit(1)
	}

	if local {
		err = writeBackup(body, path)
		if err != nil {
			log.Printf("Error writing backup to file: %s", err)
			os.Exit(1)
		}
	}

	fmt.Println("Backup completed")

}

func writeBackup(data []byte, path string) error {

	err := ioutil.WriteFile(path, data, 0600)
	if err != nil {
		return fmt.Errorf("Error writing data to file: %s", err)
	}

	return nil

}
