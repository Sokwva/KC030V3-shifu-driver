package utils

import (
	"log/slog"
	"os"
)

var Log *slog.Logger = slog.Default()

func InitLogger(levelStr string) {
	var level slog.Level
	var logMap map[string]slog.Level = map[string]slog.Level{
		"debug": slog.LevelDebug,
		"info":  slog.LevelInfo,
		"warn":  slog.LevelWarn,
		"error": slog.LevelError,
	}
	if v, ok := logMap[levelStr]; ok {
		level = v
	} else {
		level = slog.LevelInfo
	}
	Log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))
	Log.Info("Init logger, with level: " + level.String())
}
