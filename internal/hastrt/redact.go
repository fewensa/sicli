package hastrt

import (
	"github.com/fewensa/sicli/internal/authflow"
	"os"
	"strings"
)

func RedactSecrets(input string) string {
	out := authflow.RedactSecrets(input)
	if token := os.Getenv("HAST_API_KEY"); token != "" {
		out = strings.ReplaceAll(out, token, "***")
	}
	return out
}
