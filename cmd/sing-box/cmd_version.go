package main

import (
	"os"
	"runtime"

	F "github.com/sagernet/sing/common/format"

	C "github.com/sagernet/sing-box/constant"

	"github.com/spf13/cobra"
)

var commandVersion = &cobra.Command{
	Use:   "version",
	Short: "Print current version of sing-box",
	Run:   printVersion,
}

func printVersion(cmd *cobra.Command, args []string) {
	os.Stderr.WriteString(F.ToString("sing-box version ", C.Version, " (", runtime.Version(), " ", runtime.GOOS, "/", runtime.GOARCH, ")\n"))
}
