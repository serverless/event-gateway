//go:generate mockgen -package mock -destination ./store.go github.com/docker/libkv/store Store
//go:generate mockgen -package mock -destination ./shortid.go github.com/serverless/event-gateway/endpoints ShortIDGenerator
//go:generate mockgen -package mock -destination ./functions.go github.com/serverless/event-gateway/endpoints FunctionExister

package mock
