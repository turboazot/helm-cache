package main

import (
	"github.com/turboazot/helm-cache/cmd"
	"go.uber.org/zap"
)

func main() {
	logger := zap.NewExample()
	defer logger.Sync()

	undo := zap.ReplaceGlobals(logger)
	defer undo()

	cmd.Execute()
}
