package cors

import (
	"github.com/serverless/event-gateway/internal/zap"
	"go.uber.org/zap/zapcore"
)

// ID uniquely identifies a CORS configuration.
type ID string

// CORS is used to configure CORS on HTTP subscriptions.
type CORS struct {
	Space            string   `json:"space" validate:"required,min=3,space"`
	ID               ID       `json:"corsId"`
	Method           string   `json:"method" validate:"eq=GET|eq=POST|eq=DELETE|eq=PUT|eq=PATCH|eq=HEAD|eq=OPTIONS"`
	Path             string   `json:"path" validate:"path"`
	AllowedOrigins   []string `json:"allowedOrigins" validate:"min=1"`
	AllowedMethods   []string `json:"allowedMethods" validate:"min=1"`
	AllowedHeaders   []string `json:"allowedHeaders" validate:"min=1"`
	AllowCredentials bool     `json:"allowCredentials"`
}

// CORSes is an array of CORS configurations.
type CORSes []*CORS

// MarshalLogObject is a part of zapcore.ObjectMarshaler interface
func (c CORS) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("space", string(c.Space))
	enc.AddString("corsId", string(c.ID))
	enc.AddString("method", string(c.Method))
	enc.AddString("path", string(c.Path))
	enc.AddArray("allowedOrigins", zap.Strings(c.AllowedOrigins))
	enc.AddArray("allowedMethods", zap.Strings(c.AllowedMethods))
	enc.AddArray("allowedHeaders", zap.Strings(c.AllowedHeaders))
	enc.AddBool("allowCredentials", c.AllowCredentials)

	return nil
}
