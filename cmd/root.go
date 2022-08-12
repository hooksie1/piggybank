package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/errors"
	"cuelang.org/go/cue/load"
	"github.com/spf13/cobra"
)

var cfgFile string
var cfg Config

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "piggy",
	Short: "piggy is the cli utility for Piggy Bank",
	Long: `piggy can either start a Piggy Bank server or be
a cli agent to interact with an existing Piggy Bank server.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	//cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "./piggybank.cue", "config file")
	cobra.OnInitialize(initConfig)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	ctx := cuecontext.New()
	val := cue.Value{}
	ext := filepath.Ext(cfgFile)

	_, err := os.Stat(cfgFile)
	if err != nil {
		cobra.CheckErr(err)
	}

	switch ext {
	case ".cue":
		buildInstances := load.Instances([]string{cfgFile}, nil)
		insts := cue.Build(buildInstances)
		val = insts[0].Value()
	case ".json", ".yaml":
		r, err := os.Open(cfgFile)
		if err != nil {
			cobra.CheckErr(err)
		}
		data, err := ioutil.ReadAll(r)
		if err != nil {
			cobra.CheckErr(err)
		}
		val = ctx.CompileBytes(data)
	default:
		cobra.CheckErr(fmt.Errorf("config file must be json or yaml format"))
	}

	if err := cfg.BuildConfig(val); err != nil {
		var errs []string
		for _, v := range errors.Errors(err) {
			_, args := v.Msg()
			path := strings.Join(v.Path(), ".")
			msg := fmt.Sprintf("%s invalid value %v", path, args[0])
			errs = append(errs, msg)
		}

		log.Fatalf("errors loading config: %v", errs[1:])
	}
}
