package logger

import (
	"github.com/rs/zerolog/log"
	"net"
	"os"
	"time"

	"github.com/rs/zerolog"
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
	conn, err := net.DialTimeout("tcp", "localhost:5000", 5*time.Second)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to connect to Logstash")
	}

	// Close the Logstash connection if it's not nil
	defer func() {
		if conn != nil {
			conn.Close()
		}
	}()

	// TODO need more dynamic approach for connection lost
	var multiWriter zerolog.LevelWriter
	if conn != nil {
		multiWriter = zerolog.MultiLevelWriter(os.Stdout, conn)
	} else {
		multiWriter = zerolog.MultiLevelWriter(os.Stdout)
	}

	logger := zerolog.New(multiWriter).With().Timestamp().Logger()

	return &logger
}
