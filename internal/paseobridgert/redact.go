package paseobridgert

import (
	"github.com/fewensa/sicli/internal/authflow"
	"os"
	"strings"
)

func RedactSecrets(input string) string {
	out := authflow.RedactSecrets(input)
	if token := os.Getenv("PASEO_BRIDGE_ACCESS_TOKEN"); token != "" {
		out = strings.ReplaceAll(out, token, "***")
	}
	return out
}
