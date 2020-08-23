/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

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
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/asticode/go-astisub"
	"github.com/spf13/cobra"
)

type JSONSrt struct {
	Index              int     `json:"index"`
	EndAt              string  `json:"end_at"`
	EndAtGroupMinute   float64 `json:"end_at_group_minute"`
	Text               string  `json:"text"`
	StartAt            string  `json:"start_at"`
	StartAtGroupMinute float64 `json:"start_at_group_minute"`
}

// exploreSrtCmd represents the exploreSrt command
var exploreSrtCmd = &cobra.Command{
	Use:   "exploreSrt",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		pathToFile, _ := cmd.Flags().GetString("file")
		s1, err := astisub.OpenFile(pathToFile)
		if err != nil {
			fmt.Println(err)
			return
		}

		for _, item := range s1.Items {
			obj := JSONSrt{
				Index:              item.Index,
				Text:               item.String(),
				StartAt:            formatDuration(item.StartAt, ",", 3),
				StartAtGroupMinute: item.StartAt.Truncate(time.Minute).Minutes(),
				EndAt:              formatDuration(item.EndAt, ",", 3),
				EndAtGroupMinute:   item.EndAt.Round(time.Minute).Minutes(),
			}

			b, _ := json.Marshal(obj)
			fmt.Println(string(b))
		}

	},
}

func init() {
	rootCmd.AddCommand(exploreSrtCmd)
	exploreSrtCmd.Flags().String("file", "", "Path to file")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// exploreSrtCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// exploreSrtCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func formatDuration(i time.Duration, millisecondSep string, numberOfMillisecondDigits int) (s string) {
	// Parse hours
	var hours = int(i / time.Hour)
	var n = i % time.Hour
	if hours < 10 {
		s += "0"
	}
	s += strconv.Itoa(hours) + ":"

	// Parse minutes
	var minutes = int(n / time.Minute)
	n = i % time.Minute
	if minutes < 10 {
		s += "0"
	}
	s += strconv.Itoa(minutes) + ":"

	// Parse seconds
	var seconds = int(n / time.Second)
	n = i % time.Second
	if seconds < 10 {
		s += "0"
	}
	s += strconv.Itoa(seconds) + millisecondSep

	// Parse milliseconds
	var milliseconds = float64(n/time.Millisecond) / float64(1000)
	s += fmt.Sprintf("%."+strconv.Itoa(numberOfMillisecondDigits)+"f", milliseconds)[2:]
	return
}
