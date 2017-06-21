package functions_test

import (
	"errors"
	"testing"

	"github.com/docker/libkv/store"
	"github.com/golang/mock/gomock"
	"github.com/serverless/event-gateway/functions"
	"github.com/serverless/event-gateway/functions/mock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestRegisterFunction_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := mock.NewMockStore(ctrl)
	payload := `{"functionId":"testfunc",` +
		`"awsLambda":{"arn":"arn:","region":"us-east-1","version":"latest","accessKeyID":"xxx","secretAccessKey":"xxx"}}`
	db.EXPECT().Put("testfunc", []byte(payload), nil).Return(nil)
	registry := &functions.Functions{DB: db, Logger: zap.NewNop()}

	fn, _ := registry.RegisterFunction(&functions.Function{
		ID: "testfunc",
		AWSLambda: &functions.AWSLambdaProperties{
			ARN:             "arn:",
			Region:          "us-east-1",
			Version:         "latest",
			AccessKeyID:     "xxx",
			SecretAccessKey: "xxx",
		},
	})

	assert.Equal(t, &functions.Function{
		ID: "testfunc",
		AWSLambda: &functions.AWSLambdaProperties{
			ARN:             "arn:",
			Region:          "us-east-1",
			Version:         "latest",
			AccessKeyID:     "xxx",
			SecretAccessKey: "xxx",
		},
	}, fn)
}

func TestRegisterFunction_Failed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := mock.NewMockStore(ctrl)
	registry := &functions.Functions{DB: db, Logger: zap.NewNop()}

	for _, test := range registerFunctionFailedTests {
		_, err := registry.RegisterFunction(test.in)
		assert.EqualError(t, err, test.error)
	}
}

func TestGetFunction_Found(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := mock.NewMockStore(ctrl)
	db.EXPECT().Get("testfunc").Return(&store.KVPair{Value: []byte(`{"functionId": "testfunc"}`)}, nil)
	registry := &functions.Functions{DB: db, Logger: zap.NewNop()}

	fn, _ := registry.GetFunction("testfunc")

	assert.Equal(t, &functions.Function{ID: "testfunc"}, fn)
}

func TestGetFunction_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := mock.NewMockStore(ctrl)
	db.EXPECT().Get("nofunc").Return(nil, errors.New("not found"))
	registry := &functions.Functions{DB: db, Logger: zap.NewNop()}

	_, err := registry.GetFunction("nofunc")

	assert.EqualError(t, err, `Function "nofunc" not found.`)
}

var registerFunctionFailedTests = []struct {
	in    *functions.Function
	error string
}{
	{&functions.Function{ID: "testfunc"}, "Function properties not specified."},
	{&functions.Function{
		ID: "testfunc",
		AWSLambda: &functions.AWSLambdaProperties{
			ARN:             "",
			Region:          "us-east-1",
			Version:         "latest",
			AccessKeyID:     "xxx",
			SecretAccessKey: "xx",
		},
	}, `Function doesn't validate. Validation error: "Key: 'Function.AWSLambda.ARN' Error:Field validation for 'ARN' failed on the 'required' tag"`},
	{&functions.Function{
		ID: "testfunc",
		AWSLambda: &functions.AWSLambdaProperties{
			ARN:             "",
			Region:          "us-east-1",
			Version:         "latest",
			AccessKeyID:     "xxx",
			SecretAccessKey: "xx",
		},
		HTTP: &functions.HTTPProperties{
			URL: "http://example.com",
		},
	}, "More that one function type specified."},
}
