package cmd

import (
	"fmt"

	"github.com/hooksie1/piggybank/service"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var databaseCmd = &cobra.Command{
	Use:          "database",
	Short:        "Interact with the piggybank db, valid args are init, lock, unlock",
	RunE:         database,
	Args:         cobra.MatchAll(cobra.MinimumNArgs(1), cobra.OnlyValidArgs),
	ValidArgs:    service.GetClientDBVerbs(),
	SilenceUsage: true,
}

func init() {
	clientCmd.AddCommand(databaseCmd)
	databaseCmd.Flags().String("key", "", "Database key")
	viper.BindPFlag("key", databaseCmd.Flags().Lookup("key"))
}

func database(cmd *cobra.Command, args []string) error {
	opts := natsOpts{
		name:   "piggy-client",
		prefix: viper.GetString("inbox_prefix"),
	}
	nc, err := newNatsConnection(opts)
	if err != nil {
		return err
	}
	key := viper.GetString("key")

	if args[0] == "unlock" && key == "" {
		return fmt.Errorf("database key required")
	}

	client := service.Client{
		Conn: nc,
	}

	request, err := service.NewDBRequest(service.DBVerb(args[0]), key)
	if err != nil {
		return err
	}

	resp, err := client.Do(request)
	if err != nil {
		return err
	}

	fmt.Println(resp)

	return nil
}
