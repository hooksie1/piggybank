package cmd

import (
	"fmt"
	"strings"

	"github.com/hooksie1/piggybank/service"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var secretsCmd = &cobra.Command{
	Use:          "secrets",
	Short:        "Interact with piggybank secrets",
	RunE:         secrets,
	Args:         cobra.MatchAll(cobra.MinimumNArgs(1), cobra.OnlyValidArgs),
	ValidArgs:    []string{"add", "get", "delete"},
	SilenceUsage: true,
}

func init() {
	clientCmd.AddCommand(secretsCmd)
	secretsCmd.Flags().StringP("id", "i", "", "Secret ID")
	viper.BindPFlag("id", secretsCmd.Flags().Lookup("id"))
	secretsCmd.MarkFlagRequired("id")
	secretsCmd.Flags().StringP("value", "v", "", "Secret value")
	viper.BindPFlag("value", secretsCmd.Flags().Lookup("value"))
}

func getSubject(verb string, id string) string {
	return fmt.Sprintf("nex.piggybank.secrets.%s.%s", strings.ToUpper(verb), id)
}

func secrets(cmd *cobra.Command, args []string) error {
	opts := natsOpts{
		name:   "piggy-client",
		prefix: viper.GetString("inbox_prefix"),
	}
	nc, err := newNatsConnection(opts)
	if err != nil {
		return err
	}
	id := viper.GetString("id")

	client := service.Client{Conn: nc}

	switch args[0] {
	case "get":
		msg, err := client.Get(id)
		if err != nil {
			return err
		}

		fmt.Println(msg)
	case "add":
		val := viper.GetString("value")
		if val == "" {
			return fmt.Errorf("value flag is required to add a secret")
		}
		msg, err := client.Post(id, []byte(val))
		if err != nil {
			return err
		}

		fmt.Println(msg)
	case "delete":
		msg, err := client.Delete(id)
		if err != nil {
			return err
		}

		fmt.Println(msg)
	}

	return nil
}
