package cmd

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"gitlab.com/hooksie1/piggybank/server"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the piggy bank webserver",
	RunE:  start,
}

func init() {
	rootCmd.AddCommand(startCmd)

}

func start(cmd *cobra.Command, args []string) error {
	opts, err := cfg.Config.getOptions()
	if err != nil {
		return err
	}

	s := server.NewServer()
	n := server.NewNatsBackend(cfg.Config.URLs, opts)
	if err := n.Connect(); err != nil {
		return err
	}

	s.SetBackend(n)

	go s.Serve()

	sigTerm := make(chan os.Signal, 1)
	signal.Notify(sigTerm, syscall.SIGINT, syscall.SIGTERM)
	<-sigTerm

	return nil

}
