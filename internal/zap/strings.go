package zap

import "go.uber.org/zap/zapcore"

// Strings is a string array that implements MarshalLogArray.
type Strings []string

// MarshalLogArray implementation
func (ss Strings) MarshalLogArray(enc zapcore.ArrayEncoder) error {
	for _, s := range ss {
		enc.AppendString(s)
	}
	return nil
}
