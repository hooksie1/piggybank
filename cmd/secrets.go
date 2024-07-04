package cmd

import (
	"fmt"
	"strings"
	"time"

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
	return fmt.Sprintf("piggybank.secrets.%s.%s", strings.ToUpper(verb), id)
}

func secrets(cmd *cobra.Command, args []string) error {
	nc, err := newNatsConnection("piggy-client")
	if err != nil {
		return err
	}
	id := viper.GetString("id")

	switch args[0] {
	case "get":
		subject := fmt.Sprintf("piggybank.secrets.GET.%s", id)
		msg, err := nc.Request(subject, nil, 1*time.Second)
		if err != nil {
			return err
		}

		fmt.Println(string(msg.Data))
	case "add":
		val := viper.GetString("value")
		if val == "" {
			return fmt.Errorf("value flag is required to add a secret")
		}
		subject := fmt.Sprintf("piggybank.secrets.POST.%s", id)
		msg, err := nc.Request(subject, []byte(val), 1*time.Second)
		if err != nil {
			return err
		}

		fmt.Println(string(msg.Data))
	case "delete":
		subject := fmt.Sprintf("piggybank.secrets.DELETE.%s", id)
		msg, err := nc.Request(subject, nil, 1*time.Second)
		if err != nil {
			return err
		}

		fmt.Println(string(msg.Data))
	}

	return nil
}
