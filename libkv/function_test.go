package libkv

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
	db.EXPECT().Get("default/testid", &store.ReadOptions{Consistent: true}).Return(nil, errors.New("KV func not found"))
	payload := []byte(`{"space":"default","functionId":"testid","provider":{"type":"http","url":"http://example.com"}}`)
	db.EXPECT().Put("default/testid", payload, nil).Return(nil)
	service := &Service{FunctionStore: db, Log: zap.NewNop()}

	_, err := service.RegisterFunction(
		&function.Function{
			ID:       "testid",
			Provider: &function.Provider{Type: function.HTTPEndpoint, URL: "http://example.com"}},
	)

	assert.Nil(t, err)
}

func TestRegisterFunction_ValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := mock.NewMockStore(ctrl)
	service := &Service{FunctionStore: db, Log: zap.NewNop()}

	fn := &function.Function{ID: "testid", Provider: &function.Provider{Type: function.HTTPEndpoint}}
	_, err := service.RegisterFunction(fn)

	assert.Equal(t, err, &function.ErrFunctionValidation{Message: "Missing required fields for HTTP endpoint."})
}

func TestRegisterFunction_AlreadyExistsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := mock.NewMockStore(ctrl)
	db.EXPECT().Get("default/testid", gomock.Any()).Return(nil, nil)
	service := &Service{FunctionStore: db, Log: zap.NewNop()}

	fn := &function.Function{
		ID:       "testid",
		Provider: &function.Provider{Type: function.HTTPEndpoint, URL: "http://example.com"}}
	_, err := service.RegisterFunction(fn)

	assert.Equal(t, err, &function.ErrFunctionAlreadyRegistered{ID: "testid"})
}

func TestRegisterFunction_PutError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := mock.NewMockStore(ctrl)
	db.EXPECT().Get("default/testid", gomock.Any()).Return(nil, errors.New("KV func not found"))
	payload := []byte(`{"space":"default","functionId":"testid","provider":{"type":"http","url":"http://example.com"}}`)
	db.EXPECT().Put("default/testid", payload, nil).Return(errors.New("KV put error"))
	service := &Service{FunctionStore: db, Log: zap.NewNop()}

	fn := &function.Function{
		ID:       "testid",
		Provider: &function.Provider{Type: function.HTTPEndpoint, URL: "http://example.com"}}
	_, err := service.RegisterFunction(fn)

	assert.EqualError(t, err, "KV put error")
}

func TestUpdateFunction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := mock.NewMockStore(ctrl)
	returned := []byte(`{"space":"default","functionId":"testid","provider":{"type":"http","url":"http://example.com"}}`)
	db.EXPECT().Get("default/testid", &store.ReadOptions{Consistent: true}).Return(&store.KVPair{Value: returned}, nil)
	payload := []byte(`{"space":"default","functionId":"testid","provider":{"type":"http","url":"http://example1.com"}}`)
	db.EXPECT().Put("default/testid", payload, nil).Return(nil)
	service := &Service{FunctionStore: db, Log: zap.NewNop()}

	fn := &function.Function{
		ID:       "testid",
		Space:    "default",
		Provider: &function.Provider{Type: function.HTTPEndpoint, URL: "http://example1.com"}}
	_, err := service.UpdateFunction(fn)

	assert.Nil(t, err)
}

func TestUpdateFunction_ValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := mock.NewMockStore(ctrl)
	service := &Service{FunctionStore: db, Log: zap.NewNop()}

	fn := &function.Function{ID: "testid", Space: "default", Provider: &function.Provider{Type: function.HTTPEndpoint}}
	_, err := service.UpdateFunction(fn)

	assert.Equal(t, err, &function.ErrFunctionValidation{Message: "Missing required fields for HTTP endpoint."})
}

func TestUpdateFunction_NotFoundError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := mock.NewMockStore(ctrl)
	db.EXPECT().Get("default/testid", gomock.Any()).Return(nil, errors.New("KV not found"))
	service := &Service{FunctionStore: db, Log: zap.NewNop()}

	fn := &function.Function{
		ID:       "testid",
		Space:    "default",
		Provider: &function.Provider{Type: function.HTTPEndpoint, URL: "http://example.com"}}
	_, err := service.UpdateFunction(fn)

	assert.Equal(t, err, &function.ErrFunctionNotFound{ID: "testid"})
}

func TestUpdateFunction_PutError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := mock.NewMockStore(ctrl)
	returned := []byte(`
		{"functionId":"testid", "space": "default", "provider":{"type":"http","url":"http://example.com"}}
		`)
	db.EXPECT().Get("default/testid", gomock.Any()).Return(&store.KVPair{Value: returned}, nil)
	payload := []byte(`{"space":"default","functionId":"testid","provider":{"type":"http","url":"http://example1.com"}}`)
	db.EXPECT().Put("default/testid", payload, nil).Return(errors.New("KV put error"))
	service := &Service{FunctionStore: db, Log: zap.NewNop()}

	fn := &function.Function{
		ID:       "testid",
		Space:    "default",
		Provider: &function.Provider{Type: function.HTTPEndpoint, URL: "http://example1.com"}}
	_, err := service.UpdateFunction(fn)

	assert.EqualError(t, err, "KV put error")
}

func TestGetFunction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := mock.NewMockStore(ctrl)
	returned := []byte(`{"functionId":"testid"}`)
	db.EXPECT().Get("default/testid", &store.ReadOptions{Consistent: true}).Return(&store.KVPair{Value: returned}, nil)
	service := &Service{FunctionStore: db, Log: zap.NewNop()}

	f, _ := service.GetFunction("default", function.ID("testid"))

	assert.Equal(t, &function.Function{ID: "testid"}, f)
}

func TestGetFunction_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := mock.NewMockStore(ctrl)
	db.EXPECT().Get("default/testid", gomock.Any()).Return(nil, errors.New("Key not found in store"))
	service := &Service{FunctionStore: db, Log: zap.NewNop()}

	_, err := service.GetFunction("default", function.ID("testid"))

	assert.Equal(t, err, &function.ErrFunctionNotFound{ID: "testid"})
}

func TestGetFunction_GetError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := mock.NewMockStore(ctrl)
	db.EXPECT().Get("default/testid", gomock.Any()).Return(nil, errors.New("KV get err"))
	service := &Service{FunctionStore: db, Log: zap.NewNop()}

	_, err := service.GetFunction("default", function.ID("testid"))

	assert.EqualError(t, err, "KV get err")
}

func TestGetFunctions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	kvs := []*store.KVPair{
		&store.KVPair{Value: []byte(`{"functionId":"f1"}`)},
		&store.KVPair{Value: []byte(`{"functionId":"f2"}`)},
	}
	db := mock.NewMockStore(ctrl)
	db.EXPECT().List("default/", &store.ReadOptions{Consistent: true}).Return(kvs, nil)
	service := &Service{FunctionStore: db, Log: zap.NewNop()}

	list, _ := service.GetFunctions("default")

	assert.Equal(t, function.Functions{{ID: function.ID("f1")}, {ID: function.ID("f2")}}, list)
}

func TestGetFunctions_ListError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := mock.NewMockStore(ctrl)
	db.EXPECT().List("default/", gomock.Any()).Return([]*store.KVPair{}, errors.New("KV list err"))
	service := &Service{FunctionStore: db, Log: zap.NewNop()}

	_, err := service.GetFunctions("default")

	assert.EqualError(t, err, "KV list err")
}

func TestGetFunctions_ListKeyNotFoundError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := mock.NewMockStore(ctrl)
	db.EXPECT().List("default/", gomock.Any()).Return([]*store.KVPair{}, errors.New("Key not found in store"))
	service := &Service{FunctionStore: db, Log: zap.NewNop()}

	list, _ := service.GetFunctions("default")

	assert.Equal(t, function.Functions{}, list)
}

func TestDeleteFunction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	kvs := []*store.KVPair{}
	subscriptionsDB := mock.NewMockStore(ctrl)
	subscriptionsDB.EXPECT().List("default/", &store.ReadOptions{Consistent: true}).Return(kvs, nil)
	functionsDB := mock.NewMockStore(ctrl)
	functionsDB.EXPECT().Delete("default/testid").Return(nil)
	service := &Service{FunctionStore: functionsDB, SubscriptionStore: subscriptionsDB, Log: zap.NewNop()}

	err := service.DeleteFunction("default", function.ID("testid"))

	assert.Nil(t, err)
}

func TestDeleteFunction_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	kvs := []*store.KVPair{}
	subscriptionsDB := mock.NewMockStore(ctrl)
	subscriptionsDB.EXPECT().List("default/", &store.ReadOptions{Consistent: true}).Return(kvs, nil)
	functionsDB := mock.NewMockStore(ctrl)
	functionsDB.EXPECT().Delete("default/testid").Return(errors.New("KV func not found"))
	service := &Service{FunctionStore: functionsDB, SubscriptionStore: subscriptionsDB, Log: zap.NewNop()}

	err := service.DeleteFunction("default", function.ID("testid"))

	assert.EqualError(t, err, `Function "testid" not found.`)
}

func TestDeleteFunction_SubscriptionExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	kvs := []*store.KVPair{
		{Value: []byte(`{"subscriptionId":"s1","default":"default","event":"test","functionId":"testid"}`)}}
	subscriptionsDB := mock.NewMockStore(ctrl)
	subscriptionsDB.EXPECT().List("default/", &store.ReadOptions{Consistent: true}).Return(kvs, nil)
	functionsDB := mock.NewMockStore(ctrl)
	service := &Service{FunctionStore: functionsDB, SubscriptionStore: subscriptionsDB, Log: zap.NewNop()}

	err := service.DeleteFunction("default", function.ID("testid"))

	assert.Equal(t, err, &function.ErrFunctionHasSubscriptionsError{})
}

func TestValidateFunction_AWSLambdaMissingRegion(t *testing.T) {
	service := &Service{Log: zap.NewNop()}

	fn := &function.Function{ID: "id", Provider: &function.Provider{Type: function.AWSLambda, ARN: "arn::"}}
	err := service.validateFunction(fn)

	assert.Equal(t, err, &function.ErrFunctionValidation{Message: "Missing required fields for AWS Lambda function."})
}

func TestValidateFunction_AWSLambdaMissingARN(t *testing.T) {
	service := &Service{Log: zap.NewNop()}

	fn := &function.Function{ID: "id", Provider: &function.Provider{Type: function.AWSLambda, Region: "us-east-1"}}
	err := service.validateFunction(fn)

	assert.Equal(t, err, &function.ErrFunctionValidation{Message: "Missing required fields for AWS Lambda function."})
}

func TestValidateFunction_HTTPMissingURL(t *testing.T) {
	service := &Service{Log: zap.NewNop()}

	fn := &function.Function{ID: "id", Provider: &function.Provider{Type: function.HTTPEndpoint}}
	err := service.validateFunction(fn)

	assert.Equal(t, err, &function.ErrFunctionValidation{Message: "Missing required fields for HTTP endpoint."})
}

func TestValidateFunction_MissingID(t *testing.T) {
	service := &Service{Log: zap.NewNop()}

	err := service.validateFunction(&function.Function{Provider: &function.Provider{Type: function.HTTPEndpoint}})

	assert.Equal(t, err, &function.ErrFunctionValidation{
		Message: "Key: 'Function.ID' Error:Field validation for 'ID' failed on the 'required' tag"})
}

func TestValidateFunction_EmulatorMissingURL(t *testing.T) {
	service := &Service{Log: zap.NewNop()}

	err := service.validateFunction(&function.Function{ID: "id", Provider: &function.Provider{Type: function.Emulator}})

	assert.Equal(t, err, &function.ErrFunctionValidation{
		Message: "Missing required field emulatorURL for Emulator function."})
}

func TestValidateFunction_EmulatorMissingAPIVersion(t *testing.T) {
	service := &Service{Log: zap.NewNop()}

	fn := &function.Function{
		ID:       "id",
		Provider: &function.Provider{Type: function.Emulator, EmulatorURL: "http://example.com"}}
	err := service.validateFunction(fn)

	assert.Equal(t, err, &function.ErrFunctionValidation{
		Message: "Missing required field apiVersion for Emulator function."})
}

func TestValidateFunction_SpaceInvalid(t *testing.T) {
	service := &Service{Log: zap.NewNop()}

	fn := &function.Function{
		ID:       "id",
		Space:    "///",
		Provider: &function.Provider{Type: function.Emulator, EmulatorURL: "http://example.com"}}
	err := service.validateFunction(fn)

	assert.Equal(t, err, &function.ErrFunctionValidation{
		Message: "Key: 'Function.Space' Error:Field validation for 'Space' failed on the 'space' tag"})
}

func TestValidateFunction_SetDefaultSpace(t *testing.T) {
	service := &Service{Log: zap.NewNop()}

	fn := &function.Function{
		ID:       "id",
		Provider: &function.Provider{Type: function.Emulator, EmulatorURL: "http://example.com"}}
	service.validateFunction(fn)

	assert.Equal(t, "default", fn.Space)
}
