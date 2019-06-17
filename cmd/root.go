// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
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
	"bytes"
	"fmt"
	"github.com/annchain/BlockDB/mylog"
	"github.com/rifflock/lfshook"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime/debug"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "BlockDB",
	Short: "Undeniable DB",
	Long:  `BlockDB to da moon`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	defer DumpStack()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func DumpStack() {
	if err := recover(); err != nil {
		logrus.WithField("obj", err).Error("Fatal error occurred. Program will exit")
		var buf bytes.Buffer
		stack := debug.Stack()
		buf.WriteString(fmt.Sprintf("Panic: %v\n", err))
		buf.Write(stack)
		dumpName := "dump_" + time.Now().Format("20060102-150405")
		nerr := ioutil.WriteFile(dumpName, buf.Bytes(), 0644)
		if nerr != nil {
			fmt.Println("write dump file error", nerr)
			fmt.Println(buf.String())
		}
		logrus.WithField("stack ", buf.String()).Error("panic")
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringP("datadir", "d", "datadir", fmt.Sprintf("Runtime directory for storage and configurations"))
	rootCmd.PersistentFlags().StringP("config", "c", "config.toml", "Path for configuration file or url of config server")
	rootCmd.PersistentFlags().StringP("log_dir", "l", "", "Path for configuration file. Not enabled by default")
	rootCmd.PersistentFlags().BoolP("log_stdout", "s", false, "Whether the log will be printed to stdout")
	rootCmd.PersistentFlags().StringP("log_level", "v", "debug", "Logging verbosity, possible values:[panic, fatal, error, warn, info, debug]")
	rootCmd.PersistentFlags().BoolP("log_line_number", "n", false, "log_line_number")
	rootCmd.PersistentFlags().BoolP("multifile_by_level", "m", false, "multifile_by_level")
	rootCmd.PersistentFlags().BoolP("multifile_by_module", "M", false, "multifile_by_module")

	_ = viper.BindPFlag("datadir", rootCmd.PersistentFlags().Lookup("datadir"))
	_ = viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
	_ = viper.BindPFlag("log.log_dir", rootCmd.PersistentFlags().Lookup("log_dir"))
	_ = viper.BindPFlag("log_line_number", rootCmd.PersistentFlags().Lookup("log_line_number"))
	_ = viper.BindPFlag("multifile_by_level", rootCmd.PersistentFlags().Lookup("multifile_by_level"))
	_ = viper.BindPFlag("multifile_by_module", rootCmd.PersistentFlags().Lookup("multifile_by_module"))
	//viper.BindPFlag("log_stdout", rootCmd.PersistentFlags().Lookup("log_stdout"))
	_ = viper.BindPFlag("log.level", rootCmd.PersistentFlags().Lookup("log_level"))

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
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

		// Search config in home directory with name ".BlockDB1" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".BlockDB1")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func panicIfError(err error, message string) {
	if err != nil {
		fmt.Println(message)
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func initLogger() {
	logdir := viper.GetString("log.log_dir")
	stdout := viper.GetBool("log_stdout")

	var writer io.Writer

	if logdir != "" {
		folderPath, err := filepath.Abs(logdir)
		panicIfError(err, fmt.Sprintf("Error on parsing log path: %s", logdir))

		abspath, err := filepath.Abs(path.Join(logdir, "run"))
		panicIfError(err, fmt.Sprintf("Error on parsing log file path: %s", logdir))

		err = os.MkdirAll(folderPath, os.ModePerm)
		panicIfError(err, fmt.Sprintf("Error on creating log dir: %s", folderPath))

		if stdout {
			logFile, err := os.OpenFile(abspath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			panicIfError(err, fmt.Sprintf("Error on creating log file: %s", abspath))
			abspath += ".log"
			fmt.Println("Will be logged to stdout and ", abspath)
			writer = io.MultiWriter(os.Stdout, logFile)
		} else {
			fmt.Println("Will be logged to ", abspath+".log")
			writer = mylog.RotateLog(abspath)
		}
	} else {
		// stdout only
		fmt.Println("Will be logged to stdout")
		writer = os.Stdout
	}

	logrus.SetOutput(writer)

	// Only log the warning severity or above.
	switch viper.GetString("log.level") {
	case "panic":
		logrus.SetLevel(logrus.PanicLevel)
	case "fatal":
		logrus.SetLevel(logrus.FatalLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "trace":
		logrus.SetLevel(logrus.TraceLevel)
	default:
		fmt.Println("Unknown level", viper.GetString("log.level"), "Set to INFO")
		logrus.SetLevel(logrus.InfoLevel)
	}

	Formatter := new(logrus.TextFormatter)
	Formatter.ForceColors = logdir == ""
	//Formatter.DisableColors = true
	Formatter.TimestampFormat = "2006-01-02 15:04:05.000000"
	Formatter.FullTimestamp = true

	logrus.SetFormatter(Formatter)

	// redirect standard log to logrus
	//log.SetOutput(logrus.StandardLogger().Writer())
	//log.Println("Standard logger. Am I here?")
	lineNum := viper.GetBool("log_line_number")
	if lineNum {
		filenameHook := mylog.NewHook()
		filenameHook.Field = "line"
		logrus.AddHook(filenameHook)
	}
	byLevel := viper.GetBool("multifile_by_level")
	if byLevel && logdir != "" {
		panicLog, _ := filepath.Abs(path.Join(logdir, "panic"))
		fatalLog, _ := filepath.Abs(path.Join(logdir, "fatal"))
		warnLog, _ := filepath.Abs(path.Join(logdir, "warn"))
		errorLog, _ := filepath.Abs(path.Join(logdir, "error"))
		infoLog, _ := filepath.Abs(path.Join(logdir, "info"))
		debugLog, _ := filepath.Abs(path.Join(logdir, "debug"))
		traceLog, _ := filepath.Abs(path.Join(logdir, "trace"))
		writerMap := lfshook.WriterMap{
			logrus.PanicLevel: mylog.RotateLog(panicLog),
			logrus.FatalLevel: mylog.RotateLog(fatalLog),
			logrus.WarnLevel:  mylog.RotateLog(warnLog),
			logrus.ErrorLevel: mylog.RotateLog(errorLog),
			logrus.InfoLevel:  mylog.RotateLog(infoLog),
			logrus.DebugLevel: mylog.RotateLog(debugLog),
			logrus.TraceLevel: mylog.RotateLog(traceLog),
		}
		logrus.AddHook(lfshook.NewHook(
			writerMap,
			Formatter,
		))
	}
	logger := logrus.StandardLogger()
	logrus.Debug("Logger initialized.")
	byModule := viper.GetBool("multifile_by_module")
	if !byModule {
		logdir = ""
	}
	mylog.InitLoggers(logger, logdir)
}
