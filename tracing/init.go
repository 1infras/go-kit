package tracing

import (
	"os"

	"github.com/spf13/viper"
)

var Enabled = false

func Init(enabled bool) {
	Enabled = enabled
	if !enabled {
		if os.Getenv("TRACING_ENABLED") == "true" {
			Enabled = true
			return
		}
		if viper.GetBool("tracing.enabled") {
			Enabled = true
			return
		}
	}
}