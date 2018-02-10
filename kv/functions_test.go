package kv

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/serverless/event-gateway/function"
	"github.com/serverless/event-gateway/mock"
	"github.com/serverless/libkv/store"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestRegisterFunction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := mock.NewMockStore(ctrl)
	db.EXPECT().Get("testid", &store.ReadOptions{Consistent: true}).Return(nil, errors.New("KV func not found"))
	db.EXPECT().Put("testid", []byte(`{"functionId":"testid","provider":{"type":"http","url":"http://example.com"}}`), nil).Return(nil)
	service := &Functions{DB: db, Log: zap.NewNop()}

	_, err := service.RegisterFunction(&function.Function{ID: "testid", Provider: &function.Provider{Type: function.HTTPEndpoint, URL: "http://example.com"}})

	assert.Nil(t, err)
}

func TestRegisterFunction_ValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := mock.NewMockStore(ctrl)
	service := &Functions{DB: db, Log: zap.NewNop()}

	_, err := service.RegisterFunction(&function.Function{ID: "testid", Provider: &function.Provider{Type: function.HTTPEndpoint}})

	assert.Equal(t, err, &ErrValidation{"Missing required fields for HTTP endpoint."})
}

func TestRegisterFunction_AlreadyExistsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := mock.NewMockStore(ctrl)
	db.EXPECT().Get("testid", gomock.Any()).Return(nil, nil)
	service := &Functions{DB: db, Log: zap.NewNop()}

	_, err := service.RegisterFunction(&function.Function{ID: "testid", Provider: &function.Provider{Type: function.HTTPEndpoint, URL: "http://example.com"}})

	assert.Equal(t, err, &ErrAlreadyRegistered{ID: "testid"})
}

func TestRegisterFunction_PutError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := mock.NewMockStore(ctrl)
	db.EXPECT().Get("testid", gomock.Any()).Return(nil, errors.New("KV func not found"))
	db.EXPECT().Put("testid", []byte(`{"functionId":"testid","provider":{"type":"http","url":"http://example.com"}}`), nil).Return(errors.New("KV put error"))
	service := &Functions{DB: db, Log: zap.NewNop()}

	_, err := service.RegisterFunction(&function.Function{ID: "testid", Provider: &function.Provider{Type: function.HTTPEndpoint, URL: "http://example.com"}})

	assert.EqualError(t, err, "KV put error")
}

func TestUpdateFunction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := mock.NewMockStore(ctrl)
	db.EXPECT().Get("testid", &store.ReadOptions{Consistent: true}).Return(&store.KVPair{Value: []byte(`{"functionId":"testid", "provider":{"type":"http","url":"http://example.com"}}`)}, nil)
	db.EXPECT().Put("testid", []byte(`{"functionId":"testid","provider":{"type":"http","url":"http://example1.com"}}`), nil).Return(nil)
	service := &Functions{DB: db, Log: zap.NewNop()}

	_, err := service.UpdateFunction(&function.Function{ID: "testid", Provider: &function.Provider{Type: function.HTTPEndpoint, URL: "http://example1.com"}})

	assert.Nil(t, err)
}

func TestUpdateFunction_ValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := mock.NewMockStore(ctrl)
	db.EXPECT().Get("testid", gomock.Any()).Return(&store.KVPair{Value: []byte(`{"functionId":"testid", "provider":{"type":"http","url":"http://example.com"}}`)}, nil)
	service := &Functions{DB: db, Log: zap.NewNop()}

	_, err := service.UpdateFunction(&function.Function{ID: "testid", Provider: &function.Provider{Type: function.HTTPEndpoint}})

	assert.Equal(t, err, &ErrValidation{"Missing required fields for HTTP endpoint."})
}

func TestUpdateFunction_NotFoundError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := mock.NewMockStore(ctrl)
	db.EXPECT().Get("testid", gomock.Any()).Return(nil, errors.New("KV not found"))
	service := &Functions{DB: db, Log: zap.NewNop()}

	_, err := service.UpdateFunction(&function.Function{ID: "testid", Provider: &function.Provider{Type: function.HTTPEndpoint, URL: "http://example.com"}})

	assert.Equal(t, err, &ErrNotFound{ID: "testid"})
}

func TestUpdateFunction_PutError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := mock.NewMockStore(ctrl)
	db.EXPECT().Get("testid", gomock.Any()).Return(&store.KVPair{Value: []byte(`{"functionId":"testid", "provider":{"type":"http","url":"http://example.com"}}`)}, nil)
	db.EXPECT().Put("testid", []byte(`{"functionId":"testid","provider":{"type":"http","url":"http://example1.com"}}`), nil).Return(errors.New("KV put error"))
	service := &Functions{DB: db, Log: zap.NewNop()}

	_, err := service.UpdateFunction(&function.Function{ID: "testid", Provider: &function.Provider{Type: function.HTTPEndpoint, URL: "http://example1.com"}})

	assert.EqualError(t, err, "KV put error")
}

func TestGetFunction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := mock.NewMockStore(ctrl)
	db.EXPECT().Get("testid", &store.ReadOptions{Consistent: true}).Return(&store.KVPair{Value: []byte(`{"functionId":"testid"}`)}, nil)
	service := &Functions{DB: db, Log: zap.NewNop()}

	f, _ := service.GetFunction(function.ID("testid"))

	assert.Equal(t, &function.Function{ID: "testid"}, f)
}

func TestGetFunction_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := mock.NewMockStore(ctrl)
	db.EXPECT().Get("testid", gomock.Any()).Return(nil, errors.New("KV func not found"))
	service := &Functions{DB: db, Log: zap.NewNop()}

	_, err := service.GetFunction(function.ID("testid"))

	assert.Equal(t, err, &ErrNotFound{"testid"})
}

func TestGetAllFunctions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	kvs := []*store.KVPair{
		&store.KVPair{Value: []byte(`{"functionId":"f1"}`)},
		&store.KVPair{Value: []byte(`{"functionId":"f2"}`)},
	}
	db := mock.NewMockStore(ctrl)
	db.EXPECT().List("", &store.ReadOptions{Consistent: true}).Return(kvs, nil)
	service := &Functions{DB: db, Log: zap.NewNop()}

	list, _ := service.GetAllFunctions()

	assert.Equal(t, []*function.Function{{ID: function.ID("f1")}, {ID: function.ID("f2")}}, list)
}

func TestGetAllFunctions_ListError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := mock.NewMockStore(ctrl)
	db.EXPECT().List("", gomock.Any()).Return([]*store.KVPair{}, errors.New("KV list err"))
	service := &Functions{DB: db, Log: zap.NewNop()}

	_, err := service.GetAllFunctions()

	assert.EqualError(t, err, "KV list err")
}

func TestDeleteFunction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := mock.NewMockStore(ctrl)
	db.EXPECT().Delete("testid").Return(nil)
	service := &Functions{DB: db, Log: zap.NewNop()}

	err := service.DeleteFunction(function.ID("testid"))

	assert.Nil(t, err)
}

func TestDeleteFunction_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := mock.NewMockStore(ctrl)
	db.EXPECT().Delete("testid").Return(errors.New("KV func not found"))
	service := &Functions{DB: db, Log: zap.NewNop()}

	err := service.DeleteFunction(function.ID("testid"))

	assert.EqualError(t, err, `Function "testid" not found.`)
}

func TestValidateFunction_AWSLambdaMissingRegion(t *testing.T) {
	service := &Functions{Log: zap.NewNop()}

	err := service.validateFunction(&function.Function{ID: "id", Provider: &function.Provider{Type: function.AWSLambda, ARN: "arn::"}})

	assert.Equal(t, err, &ErrValidation{"Missing required fields for AWS Lambda function."})
}

func TestValidateFunction_AWSLambdaMissingARN(t *testing.T) {
	service := &Functions{Log: zap.NewNop()}

	err := service.validateFunction(&function.Function{ID: "id", Provider: &function.Provider{Type: function.AWSLambda, Region: "us-east-1"}})

	assert.Equal(t, err, &ErrValidation{"Missing required fields for AWS Lambda function."})
}

func TestValidateFunction_HTTPMissingURL(t *testing.T) {
	service := &Functions{Log: zap.NewNop()}

	err := service.validateFunction(&function.Function{ID: "id", Provider: &function.Provider{Type: function.HTTPEndpoint}})

	assert.Equal(t, err, &ErrValidation{"Missing required fields for HTTP endpoint."})
}

func TestValidateFunction_MissingID(t *testing.T) {
	service := &Functions{Log: zap.NewNop()}

	err := service.validateFunction(&function.Function{Provider: &function.Provider{Type: function.HTTPEndpoint}})

	assert.Equal(t, err, &ErrValidation{"Key: 'Function.ID' Error:Field validation for 'ID' failed on the 'required' tag"})
}

func TestValidateFunction_EmulatorMissingURL(t *testing.T) {
	service := &Functions{Log: zap.NewNop()}

	err := service.validateFunction(&function.Function{ID: "id", Provider: &function.Provider{Type: function.Emulator}})

	assert.Equal(t, err, &ErrValidation{"Missing required field emulatorURL for Emulator function."})
}

func TestValidateFunction_EmulatorMissingAPIVersion(t *testing.T) {
	service := &Functions{Log: zap.NewNop()}

	err := service.validateFunction(&function.Function{ID: "id", Provider: &function.Provider{Type: function.Emulator, EmulatorURL: "http://example.com"}})

	assert.Equal(t, err, &ErrValidation{"Missing required field apiVersion for Emulator function."})
}
