// Copyright © 2019 Annchain Authors <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package cmd

import (
	"fmt"
	"github.com/annchain/BlockDB/engine"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"syscall"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Start a full node",
	Long:  `Start a full node`,
	Run: func(cmd *cobra.Command, args []string) {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println(err)
				panic(err)
			}
		}()

		// init logs and other facilities before the node starts
		readConfig()
		initLogger()
		log.Info("BlockDB Starting")
		eng := engine.NewEngine()
		eng.Start()

		// prevent sudden stop. Do your clean up here
		var gracefulStop = make(chan os.Signal)

		signal.Notify(gracefulStop, syscall.SIGTERM)
		signal.Notify(gracefulStop, syscall.SIGINT)

		func() {
			sig := <-gracefulStop
			log.Infof("caught sig: %+v", sig)
			log.Info("Exiting... Please do no kill me")
			eng.Stop()
			os.Exit(0)
		}()

	},
}

func readConfig() {
	configPath := viper.GetString("config")
	//datadirPath := viper.GetString("datadir")

	absPath, err := filepath.Abs(configPath)
	fmt.Println(absPath)
	panicIfError(err, fmt.Sprintf("Error on parsing config file path: %s", absPath))

	file, err := os.Open(absPath)
	panicIfError(err, fmt.Sprintf("Error on opening config file: %s", absPath))
	defer file.Close()

	viper.SetConfigType("toml")
	err = viper.MergeConfig(file)
	panicIfError(err, fmt.Sprintf("Error on reading config file: %s", absPath))

	viper.SetEnvPrefix("blockdb")
	viper.AutomaticEnv() // read in environment variables that match

	keys := viper.AllKeys()
	sort.Strings(keys)
	for _, key := range keys {
		fmt.Printf("%s:%v\n", key, viper.Get(key))
	}

}

func init() {
	rootCmd.AddCommand(runCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
