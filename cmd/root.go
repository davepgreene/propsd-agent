package cmd

import (
	"fmt"
	"os"
	"time"
	log "github.com/sirupsen/logrus"
	"github.com/davepgreene/propsd-agent/http"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/davepgreene/propsd-agent/config"
	"github.com/davepgreene/propsd-agent/sources"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/davepgreene/propsd-agent/utils"
)

var cfgFile string
var verbose bool

var PropsdCmd = &cobra.Command{
	Use:   "propsd",
	Short: "Dynamic property management at scale",
	Long: `The Propsd Agent is the local interface for Propsd, a
	dynamic property management system that runs at scale, across
	thousands of servers and changes from hundreds of developers, leveraging
	Amazon S3 to deliver properties and Consul to handle service discovery.
	Composable layering lets you set properties for an organization, a single
	server, and everything in between. Plus, flat file storage makes backups
	and audits a breeze.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := initializeConfig()
		initializeLog()
		if err != nil {
			panic(err)
		}

		s, err := session.NewSession()
		if err == nil {
			m := sources.NewMetadataSource(*s)
			m.Get()
			m.Tags()
			utils.Schedule(m.Get, time.Millisecond * viper.GetDuration("metadata.interval"))
			utils.Schedule(m.Tags, time.Millisecond * viper.GetDuration("tags.interval"))
			http.Handler(m)
		}
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the PropsdCmd.
func Execute() {
	if err := PropsdCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	PropsdCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file")
	PropsdCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose level logging")
	validConfigFilenames := []string{"json"}
	PropsdCmd.PersistentFlags().SetAnnotation("config", cobra.BashCompFilenameExt, validConfigFilenames)
}

func initializeLog() {
	log.RegisterExitHandler(func() {
		log.Info("Shutting down")
	})

	// Set logging options based on config
	if lvl, err := log.ParseLevel(viper.GetString("log.level")); err == nil {
		log.SetLevel(lvl)
	} else {
		log.Info("Unable to parse log level in settings. Defaulting to INFO")
	}

	// If using verbose mode, log at debug level
	if verbose {
		log.SetLevel(log.DebugLevel)
	}

	if viper.GetBool("log.json") {
		log.SetFormatter(&log.JSONFormatter{})
	}

	if cfgFile != "" {
		log.WithFields(log.Fields{
			"file": viper.ConfigFileUsed(),
		}).Info("Loaded config file")
	}

}

func initializeConfig(subCmdVs ...*cobra.Command) error {
	config.Defaults()

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)

		// If a config file is found, read it in.
		if err := viper.ReadInConfig(); err != nil {
			return err
		}
	}
	viper.AutomaticEnv() // read in environment variables that match

	return nil
}
