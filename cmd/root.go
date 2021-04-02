/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"

	"github.com/sebbegg/avahisync/avahisync"

	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile string
var syncConfig *avahisync.SyncConfig

func portMapperFromFlags(cmd *cobra.Command) avahisync.PortMapper {

	portMap := make(avahisync.StaticPortMap, 0)
	portMappingsRaw, err := cmd.Flags().GetStringArray("port")
	if err != nil {
		panic(err)
	}

	parseGroup := func(s string) uint16 {
		number, err := strconv.ParseUint(s, 10, 16)
		if err != nil {
			panic(fmt.Sprintf("Not a 32bit uint: %s", s))
		}
		return uint16(number)
	}

	pattern := regexp.MustCompile("(\\d+):(\\d+)")
	for _, s := range portMappingsRaw {
		if groups := pattern.FindStringSubmatch(s); groups != nil {
			portMap[parseGroup(groups[1])] = parseGroup(groups[2])
		}
	}
	fmt.Printf("Parsed portmap: %v\n", portMap)

	return &avahisync.StaticPortMapper{
		PortMap: portMap,
	}
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "avahisync",
	Short: "Sync zeroconf/bonjour/avahi services to avahi xmls",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {

		initSyncConfig(cmd)
		avahisync.Sync(syncConfig)
	},
}

func initSyncConfig(cmd *cobra.Command) {

	//syncConfig.PortMapper = portMapperFromFlags(cmd)
	if useDocker, err := cmd.Flags().GetBool("docker"); useDocker && err == nil {
		syncConfig.PortMapper, err = avahisync.NewDockerPortMapper()
	} else if err != nil {
		log.Fatalf("Could not init docker port mapper: %s", err.Error())
	} else {
		syncConfig.PortMapper = portMapperFromFlags(cmd)
	}

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
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.avahisync.yaml)")

	syncConfig = &avahisync.SyncConfig{
		Service:      "",
		Domain:       "local.",
		FilePrefix:   "sync_",
		OutputFolder: ".",
	}
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().StringVarP(&syncConfig.Domain, "domain", "d", syncConfig.Domain, "Domain to query")
	rootCmd.Flags().StringVarP(&syncConfig.Service, "service", "s", syncConfig.Service, "Service to query")
	rootCmd.MarkFlagRequired("service")

	rootCmd.Flags().StringVarP(&syncConfig.FilePrefix, "prefix", "x", syncConfig.FilePrefix, "Prefix for .service xml files")
	rootCmd.Flags().StringVarP(&syncConfig.OutputFolder, "output", "o", syncConfig.OutputFolder, "Folder where .service files will be written to")
	rootCmd.Flags().StringArrayP("port", "p", make([]string, 0), "Port mapping of type in:out")

	hostname, _ := os.Hostname()
	rootCmd.Flags().StringVar(&syncConfig.HostName, "hostname", hostname+".local.", "Hostname of the sync'd services")

	rootCmd.Flags().Bool("docker", false, "Use docker api for port mapping")
	rootCmd.Flags().Lookup("docker").NoOptDefVal = "true"
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".avahisync" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".avahisync")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

}
