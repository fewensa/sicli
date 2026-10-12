package main

import (
	"fmt"
	"os"

	"github.com/fewensa/sicli/internal/cli"
	"github.com/fewensa/sicli/internal/registry"
)

const help = `sicli

Umbrella CLI for service command-line tools.

Usage:
  sicli --help
  sicli services

Commands:
  services    List registered service CLI definitions.
`

func main() {
	args := os.Args[1:]
	if len(args) == 0 || hasHelp(args) {
		fmt.Print(help)
		return
	}

	if len(args) == 1 && args[0] == "services" {
		_ = cli.WriteJSON(os.Stdout, cli.Response{
			OK: true,
			Data: map[string]any{
				"services": registry.Services(),
			},
			Meta: map[string]any{
				"registry": "services",
			},
		})
		return
	}

	_ = cli.UnknownCommand(os.Stdout, args)
	os.Exit(64)
}

func hasHelp(args []string) bool {
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			return true
		}
	}
	return false
}
