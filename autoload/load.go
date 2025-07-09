package autoload

import (
	"log/slog"

	"github.com/bsagat/envzilla/v2"
)

func init() {
	err := envzilla.Loader()
	if err != nil {
		slog.Error("Environment variable autolaod error: ", "message", err.Error())
		return
	}
	slog.Info("Config file has been parsed :D")
}
