package zap

import (
	"encoding/json"
	"go.uber.org/zap/zapcore"
)

// Strings is a string array that implements MarshalLogArray.
type Strings []string

// MapStringInterface is a map that implements MarshalLogObject.
type MapStringInterface map[string]interface{}

// MarshalLogArray implementation
func (ss Strings) MarshalLogArray(enc zapcore.ArrayEncoder) error {
	for _, s := range ss {
		enc.AppendString(s)
	}
	return nil
}

// MarshalLogObject implementation
func (msi MapStringInterface) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	for key, val := range msi {
		v, err := json.Marshal(val)
		if err != nil {
			enc.AddString(key, string(v))
		} else {
			return err
		}
	}
	return nil
}
