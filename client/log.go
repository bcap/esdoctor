package client

import (
	"bytes"

	"github.com/elastic/go-elasticsearch/v8/estransport"
	log "github.com/sirupsen/logrus"
)

func NewRequestResponseLogger(level log.Level, logRequestBody bool, logResponseBody bool) estransport.Logger {
	return &estransport.ColorLogger{
		Output: loggerWriter{
			logger: log.StandardLogger(),
			level:  level,
		},
		EnableRequestBody:  logRequestBody,
		EnableResponseBody: logResponseBody,
	}
}

type loggerWriter struct {
	logger *log.Logger
	level  log.Level
}

func (w loggerWriter) Write(data []byte) (int, error) {
	if !w.logger.IsLevelEnabled(w.level) {
		return 0, nil
	}
	lines := bytes.Split(data, []byte{'\n'})
	written := 0
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		w.logger.Log(w.level, string(line))
		written += len(line)
	}
	return written, nil
}
