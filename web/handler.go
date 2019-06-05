package web

import "bitbucket.com/turntwo/quicksight-embeds/config"

// Handler holds the common fields to handle a web request
type Handler struct {
	Config *config.Config
}

// NewHandler builds and returns a handler with the Config object
func NewHandler(cfg config.Config) Handler {
	if len(cfg.Quicksight.Namespace) == 0 {
		cfg.Quicksight.Namespace = "default"
	}

	return Handler{Config: &cfg}
}
