package libkv

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/serverless/event-gateway/function"
	"github.com/serverless/event-gateway/mock"
	"github.com/serverless/event-gateway/providers/http"
	"github.com/serverless/libkv/store"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	_ "github.com/serverless/event-gateway/providers/http"
)

func TestRegisterFunction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := mock.NewMockStore(ctrl)
	db.EXPECT().Get("default/testid", &store.ReadOptions{Consistent: true}).Return(nil, errors.New("KV func not found"))
	payload := []byte(`{"space":"default","functionId":"testid","type":"http","provider":{"url":"http://example.com"}}`)
	db.EXPECT().Put("default/testid", payload, nil).Return(nil)
	service := &Service{FunctionStore: db, Log: zap.NewNop()}

	_, err := service.RegisterFunction(
		&function.Function{
			ID:           "testid",
			ProviderType: http.Type,
			Provider:     &http.HTTP{URL: "http://example.com"},
		},
	)

	assert.Nil(t, err)
}

func TestRegisterFunction_AlreadyExistsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := mock.NewMockStore(ctrl)
	db.EXPECT().Get("default/testid", gomock.Any()).Return(nil, nil)
	service := &Service{FunctionStore: db, Log: zap.NewNop()}

	fn := &function.Function{
		ID:           "testid",
		ProviderType: http.Type,
		Provider:     &http.HTTP{URL: "http://example.com"},
	}
	_, err := service.RegisterFunction(fn)

	assert.Equal(t, err, &function.ErrFunctionAlreadyRegistered{ID: "testid"})
}

func TestRegisterFunction_PutError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := mock.NewMockStore(ctrl)
	db.EXPECT().Get("default/testid", gomock.Any()).Return(nil, errors.New("KV func not found"))
	payload := []byte(`{"space":"default","functionId":"testid","type":"http","provider":{"url":"http://example.com"}}`)
	db.EXPECT().Put("default/testid", payload, nil).Return(errors.New("KV put error"))
	service := &Service{FunctionStore: db, Log: zap.NewNop()}

	fn := &function.Function{
		ID:           "testid",
		ProviderType: http.Type,
		Provider:     &http.HTTP{URL: "http://example.com"},
	}
	_, err := service.RegisterFunction(fn)

	assert.EqualError(t, err, "KV put error")
}

func TestUpdateFunction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := mock.NewMockStore(ctrl)
	returned := []byte(`{"space":"default","functionId":"testid","type":"http","provider":{"url":"http://example.com"}}`)
	db.EXPECT().Get("default/testid", &store.ReadOptions{Consistent: true}).Return(&store.KVPair{Value: returned}, nil)
	payload := []byte(`{"space":"default","functionId":"testid","type":"http","provider":{"url":"http://example1.com"}}`)
	db.EXPECT().Put("default/testid", payload, nil).Return(nil)
	service := &Service{FunctionStore: db, Log: zap.NewNop()}

	fn := &function.Function{
		ID:           "testid",
		Space:        "default",
		ProviderType: http.Type,
		Provider:     &http.HTTP{URL: "http://example1.com"},
	}
	_, err := service.UpdateFunction(fn)

	assert.Nil(t, err)
}

func TestUpdateFunction_NotFoundError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := mock.NewMockStore(ctrl)
	db.EXPECT().Get("default/testid", gomock.Any()).Return(nil, errors.New("KV not found"))
	service := &Service{FunctionStore: db, Log: zap.NewNop()}

	fn := &function.Function{
		ID:           "testid",
		Space:        "default",
		ProviderType: http.Type,
		Provider:     &http.HTTP{URL: "http://example.com"},
	}
	_, err := service.UpdateFunction(fn)

	assert.Equal(t, err, &function.ErrFunctionNotFound{ID: "testid"})
}

func TestUpdateFunction_PutError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := mock.NewMockStore(ctrl)
	returned := []byte(`
		{"functionId":"testid", "space": "default", "type": "http", "provider": {"url":"http://example.com"}}
		`)
	db.EXPECT().Get("default/testid", gomock.Any()).Return(&store.KVPair{Value: returned}, nil)
	payload := []byte(`{"space":"default","functionId":"testid","type":"http","provider":{"url":"http://example.com"}}`)
	db.EXPECT().Put("default/testid", payload, nil).Return(errors.New("KV put error"))
	service := &Service{FunctionStore: db, Log: zap.NewNop()}

	fn := &function.Function{
		ID:           "testid",
		Space:        "default",
		ProviderType: http.Type,
		Provider:     &http.HTTP{URL: "http://example.com"},
	}
	_, err := service.UpdateFunction(fn)

	assert.EqualError(t, err, "KV put error")
}

func TestGetFunction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := mock.NewMockStore(ctrl)
	returned := []byte(`{"functionId":"f1","type":"http","provider":{"url": "http://test.com"}}}`)
	db.EXPECT().Get("default/testid", &store.ReadOptions{Consistent: true}).Return(&store.KVPair{Value: returned}, nil)
	service := &Service{FunctionStore: db, Log: zap.NewNop()}

	f, _ := service.GetFunction("default", function.ID("testid"))

	assert.Equal(t, &function.Function{
		ID:           function.ID("f1"),
		ProviderType: http.Type,
		Provider:     &http.HTTP{URL: "http://test.com"},
	}, f)
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
		&store.KVPair{Value: []byte(`{"functionId":"f1","type":"http","provider":{"url": "http://test.com"}}}`)},
	}
	db := mock.NewMockStore(ctrl)
	db.EXPECT().List("default/", &store.ReadOptions{Consistent: true}).Return(kvs, nil)
	service := &Service{FunctionStore: db, Log: zap.NewNop()}

	list, err := service.GetFunctions("default")

	assert.Nil(t, err)
	assert.Equal(t, function.Functions{{
		ID:           function.ID("f1"),
		ProviderType: http.Type,
		Provider:     &http.HTTP{URL: "http://test.com"},
	}}, list)
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

func TestValidateFunction_MissingID(t *testing.T) {
	service := &Service{Log: zap.NewNop()}

	err := service.validateFunction(&function.Function{})

	assert.Equal(t, err, &function.ErrFunctionValidation{
		Message: "Key: 'Function.ID' Error:Field validation for 'ID' failed on the 'required' tag"})
}

func TestValidateFunction_SpaceInvalid(t *testing.T) {
	service := &Service{Log: zap.NewNop()}

	fn := &function.Function{
		ID:    "id",
		Space: "///"}
	err := service.validateFunction(fn)

	assert.Equal(t, err, &function.ErrFunctionValidation{
		Message: "Key: 'Function.Space' Error:Field validation for 'Space' failed on the 'space' tag"})
}

func TestValidateFunction_SetDefaultSpace(t *testing.T) {
	service := &Service{Log: zap.NewNop()}

	fn := &function.Function{
		ID:           "id",
		ProviderType: http.Type,
		Provider:     http.HTTP{URL: "http://example.com"},
	}
	service.validateFunction(fn)

	assert.Equal(t, "default", fn.Space)
}
