package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// databaseCmd represents the database command
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

	switch args[0] {
	case "init":
		msg, err := nc.Request("piggybank.database.initialize", nil, 1*time.Second)
		if err != nil {
			return err
		}

		fmt.Println(string(msg.Data))
		return nil
	}

	return nil
}
