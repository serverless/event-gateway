package plugin

import (
	"fmt"
	"io/ioutil"
	"log"

	hclog "github.com/hashicorp/go-hclog"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Hclog2ZapLogger implements Hashicorp's hclog.Logger interface using Uber's zap.Logger. It's a workaround for plugin
// system. go-plugin doesn't support other logger than hclog. This logger implements only methods used by the go-plugin.
type Hclog2ZapLogger struct {
	Zap *zap.Logger
}

// Trace implementation.
func (l Hclog2ZapLogger) Trace(msg string, args ...interface{}) {}

// Debug implementation.
func (l Hclog2ZapLogger) Debug(msg string, args ...interface{}) {
	l.Zap.Debug(msg, argsToFields(args...)...)
}

// Info implementation.
func (l Hclog2ZapLogger) Info(msg string, args ...interface{}) {
	l.Zap.Info(msg, argsToFields(args...)...)
}

// Warn implementation.
func (l Hclog2ZapLogger) Warn(msg string, args ...interface{}) {
	l.Zap.Warn(msg, argsToFields(args...)...)
}

// Error implementation.
func (l Hclog2ZapLogger) Error(msg string, args ...interface{}) {
	l.Zap.Error(msg, argsToFields(args...)...)
}

// IsTrace implementation.
func (l Hclog2ZapLogger) IsTrace() bool { return false }

// IsDebug implementation.
func (l Hclog2ZapLogger) IsDebug() bool { return false }

// IsInfo implementation.
func (l Hclog2ZapLogger) IsInfo() bool { return false }

// IsWarn implementation.
func (l Hclog2ZapLogger) IsWarn() bool { return false }

// IsError implementation.
func (l Hclog2ZapLogger) IsError() bool { return false }

// With implementation.
func (l Hclog2ZapLogger) With(args ...interface{}) hclog.Logger {
	return Hclog2ZapLogger{l.Zap.With(argsToFields(args...)...)}
}

// Named implementation.
func (l Hclog2ZapLogger) Named(name string) hclog.Logger {
	return Hclog2ZapLogger{l.Zap.Named(name)}
}

// ResetNamed implementation.
func (l Hclog2ZapLogger) ResetNamed(name string) hclog.Logger {
	// no need to implement that as go-plugin doesn't use this method.
	return Hclog2ZapLogger{}
}

// StandardLogger implementation.
func (l Hclog2ZapLogger) StandardLogger(opts *hclog.StandardLoggerOptions) *log.Logger {
	// no need to implement that as go-plugin doesn't use this method.
	return log.New(ioutil.Discard, "", 0)
}

func argsToFields(args ...interface{}) []zapcore.Field {
	fields := []zapcore.Field{}
	for i := 0; i < len(args); i += 2 {
		fields = append(fields, zap.String(args[i].(string), fmt.Sprintf("%v", args[i+1])))
	}

	return fields
}
