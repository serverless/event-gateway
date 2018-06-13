//go:generate mockgen -package mock -destination ./store.go github.com/serverless/libkv/store Store
//go:generate mockgen -package mock -destination ./eventtype.go -mock_names "Service=MockEventTypeService" github.com/serverless/event-gateway/event Service
//go:generate mockgen -package mock -destination ./function.go -mock_names "Service=MockFunctionService" github.com/serverless/event-gateway/function Service
//go:generate mockgen -package mock -destination ./subscription.go -mock_names "Service=MockSubscriptionService" github.com/serverless/event-gateway/subscription Service
//go:generate mockgen -package mock -destination ./cors.go -mock_names "Service=MockCORSService" github.com/serverless/event-gateway/subscription/cors Service

package mock
