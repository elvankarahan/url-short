package logger

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net"
	"os"
	"time"
)

// New creates a new logger with the specified log level and outputs.
// If isDebug is true, it sets the log level to Trace; otherwise, it sets it to Info.
// It attempts to create a TCP connection to Logstash with a timeout of 5 seconds.
// It returns a logger configured with the appropriate outputs.
func New(isDebug bool) *zerolog.Logger {
	logLevel := zerolog.InfoLevel
	if isDebug {
		logLevel = zerolog.TraceLevel
	}

	zerolog.SetGlobalLevel(logLevel)

	// Create a TCP connection to Logstash
	conn, err := net.DialTimeout("udp", os.Getenv("LOGSTASH_ADDR"), 5*time.Second)
	if err != nil {
		log.Error().Err(err)
	}

	// TODO need more dynamic approach for connection lost
	var multiWriter zerolog.LevelWriter
	if conn != nil {
		log.Info().Msg(`Connected to Logstash`)
		multiWriter = zerolog.MultiLevelWriter(os.Stdout, conn)
	} else {
		multiWriter = zerolog.MultiLevelWriter(os.Stdout)
	}

	logger := zerolog.New(multiWriter).With().Timestamp().Logger()

	return &logger
}
