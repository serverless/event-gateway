//go:generate mockgen -package mock -destination kinesisiface.go github.com/aws/aws-sdk-go/service/kinesis/kinesisiface KinesisAPI

package mock
