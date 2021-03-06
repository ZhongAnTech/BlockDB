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
	"github.com/ZhongAnTech/BlockDB/brefactor/core"
	"github.com/annchain/commongo/mylog"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"syscall"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Start a BlockDB instance",
	Long:  `Start a BlockDB instance`,
	Run: func(cmd *cobra.Command, args []string) {
		logrus.Info("BlockDB Starting")
		folderConfigs := ensureFolders()
		readConfig(folderConfigs.Config)
		mylog.InitLogger(logrus.StandardLogger(), mylog.LogConfig{
			MaxSize:    10,
			MaxBackups: 100,
			MaxAgeDays: 90,
			Compress:   true,
			LogDir:     folderConfigs.Log,
			OutputFile: "blockdb",
		})

		// init logs and other facilities before the node starts

		blockdb := core.BlockDB{}
		blockdb.InitDefault()
		blockdb.Setup()
		blockdb.Start()

		// prevent sudden stop. Do your clean up here
		var gracefulStop = make(chan os.Signal)

		signal.Notify(gracefulStop, syscall.SIGTERM)
		signal.Notify(gracefulStop, syscall.SIGINT)

		func() {
			sig := <-gracefulStop
			logrus.Infof("caught sig: %+v", sig)
			logrus.Info("Exiting... Please do no kill me")
			blockdb.Stop()
			os.Exit(0)
		}()

	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	// ./mongodb run -c config.toml

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
