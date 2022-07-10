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

	err := cmd.Execute()
	if err != nil {
		zap.L().Sugar().Fatalf("Fail to initialize root command: %v", err)
	}
}
