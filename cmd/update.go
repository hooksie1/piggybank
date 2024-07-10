package cmd

import (
	"runtime"
	"time"

	"github.com/CoverWhale/gupdate"
	"github.com/briandowns/spinner"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "updates piggybank",
	RunE:  update,
}

func init() {
	rootCmd.AddCommand(updateCmd)
}

func update(cmd *cobra.Command, args []string) error {

	gh := gupdate.GitHubProject{
		Name:         "piggybank",
		Owner:        "hooksie1",
		Platform:     runtime.GOOS,
		Arch:         runtime.GOARCH,
		ChecksumFunc: gupdate.GoReleaserChecksum,
	}

	s := spinner.New(spinner.CharSets[33], 100*time.Millisecond)
	s.Suffix = " updating piggybankctl..."
	s.Start()
	release, err := gupdate.GetLatestRelease(gh)
	if err != nil {
		return err
	}

	if err := release.Update(); err != nil {
		return err
	}
	s.Stop()

	return nil
}
