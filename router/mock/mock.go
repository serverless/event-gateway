//go:generate mockgen -package mock -destination ./targetcache.go github.com/serverless/event-gateway/router Targeter
//go:generate mockgen -package mock -destination ./provider.go github.com/serverless/event-gateway/function Provider

package mock
