//go:generate mockgen -package mock -destination sqsiface.go github.com/aws/aws-sdk-go/service/sqs/sqsiface SQSAPI

package mock
