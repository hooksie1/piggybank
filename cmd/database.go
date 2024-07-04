package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hooksie1/piggybank/service"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var databaseCmd = &cobra.Command{
	Use:          "database",
	Short:        "Interact with the piggybank db, valid args are init, lock, unlock",
	RunE:         database,
	Args:         cobra.MatchAll(cobra.MinimumNArgs(1), cobra.OnlyValidArgs),
	ValidArgs:    []string{"init", "lock", "unlock"},
	SilenceUsage: true,
}

func init() {
	clientCmd.AddCommand(databaseCmd)
	databaseCmd.Flags().String("key", "", "Database key")
	viper.BindPFlag("key", databaseCmd.Flags().Lookup("key"))
}

func database(cmd *cobra.Command, args []string) error {
	nc, err := newNatsConnection("piggy-client")
	if err != nil {
		return err
	}
	key := viper.GetString("key")

	switch args[0] {
	case "init":
		msg, err := nc.Request("piggybank.database.initialize", nil, 1*time.Second)
		if err != nil {
			return err
		}

		fmt.Println(string(msg.Data))
		return nil
	case "unlock":
		if key == "" {
			return fmt.Errorf("database key required")
		}

		req := service.DatabaseKey{DBKey: key}

		data, err := json.Marshal(req)
		if err != nil {
			return err
		}
		msg, err := nc.Request("piggybank.database.unlock", data, 1*time.Second)
		if err != nil {
			return err
		}

		fmt.Println(string(msg.Data))

	case "lock":
		msg, err := nc.Request("piggybank.database.lock", nil, 1*time.Second)
		if err != nil {
			return err
		}

		fmt.Println(string(msg.Data))
	}

	return nil
}
