package common 
import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
)

var Logger *zap.SugaredLogger 

func InitLogger(mode string) {
	var zapLogger *zap.Logger
	var err error
	if mode == "release" {
		// Production: JSON, info level, no stacktrace unless error
		config := zap.NewProductionConfig()
		config.EncoderConfig.TimeKey = "ts"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		zapLogger, err = config.Build()
	} else {
		// Development: colored, debug level, stacktrace on warnings
		config := zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		zapLogger, err = config.Build()
	}
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	Logger = zapLogger.Sugar()
}