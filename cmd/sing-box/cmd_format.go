package main

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/sagernet/sing-box/option"

	"github.com/goccy/go-json"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var commandFormatFlagWrite bool

var commandFormat = &cobra.Command{
	Use:   "format",
	Short: "Format configuration",
	Run:   formatConfiguration,
}

func init() {
	commandFormat.Flags().BoolVarP(&commandFormatFlagWrite, "write", "w", false, "write result to (source) file instead of stdout")
}

func formatConfiguration(cmd *cobra.Command, args []string) {
	configContent, err := os.ReadFile(configPath)
	if err != nil {
		logrus.Fatal("read config: ", err)
	}
	var options option.Options
	err = json.Unmarshal(configContent, &options)
	if err != nil {
		logrus.Fatal("decode config: ", err)
	}
	buffer := new(bytes.Buffer)
	encoder := json.NewEncoder(buffer)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(options)
	if err != nil {
		logrus.Fatal("encode config: ", err)
	}
	if !commandFormatFlagWrite {
		os.Stdout.WriteString(buffer.String() + "\n")
		return
	}
	if bytes.Compare(configContent, buffer.Bytes()) == 0 {
		return
	}
	output, err := os.Create(configPath)
	if err != nil {
		logrus.Fatal("open output: ", err)
	}
	_, err = output.Write(buffer.Bytes())
	output.Close()
	if err != nil {
		logrus.Fatal("write output: ", err)
	}
	outputPath, _ := filepath.Abs(configPath)
	os.Stderr.WriteString(outputPath + "\n")
}
