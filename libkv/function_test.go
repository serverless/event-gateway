package libkv

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/serverless/event-gateway/function"
	"github.com/serverless/event-gateway/metadata"
	"github.com/serverless/event-gateway/mock"
	"github.com/serverless/event-gateway/providers/http"
	"github.com/serverless/libkv/store"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	_ "github.com/serverless/event-gateway/providers/http"
)

func TestCreateFunction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	fn := &function.Function{
		ID:           "testid",
		ProviderType: http.Type,
		Provider:     &http.HTTP{URL: "http://example.com"},
	}

	t.Run("function registered", func(t *testing.T) {
		db := mock.NewMockStore(ctrl)
		db.EXPECT().Get("default/testid", &store.ReadOptions{Consistent: true}).Return(nil, errors.New("KV func not found"))
		payload := []byte(`{"space":"default","functionId":"testid","type":"http","provider":{"url":"http://example.com"}}`)
		db.EXPECT().AtomicPut("default/testid", payload, nil, nil).Return(true, nil, nil)
		service := &Service{FunctionStore: db, Log: zap.NewNop()}

		_, err := service.CreateFunction(fn)

		assert.Nil(t, err)
	})

	t.Run("function already exists", func(t *testing.T) {
		db := mock.NewMockStore(ctrl)
		db.EXPECT().Get("default/testid", gomock.Any()).Return(nil, nil)
		service := &Service{FunctionStore: db, Log: zap.NewNop()}

		_, err := service.CreateFunction(fn)

		assert.Equal(t, err, &function.ErrFunctionAlreadyRegistered{ID: "testid"})
	})

	t.Run("KV Put error", func(t *testing.T) {
		db := mock.NewMockStore(ctrl)
		db.EXPECT().Get("default/testid", gomock.Any()).Return(nil, errors.New("KV func not found"))
		db.EXPECT().AtomicPut(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil, errors.New("KV put error"))
		service := &Service{FunctionStore: db, Log: zap.NewNop()}

		_, err := service.CreateFunction(fn)

		assert.EqualError(t, err, "KV put error")
	})
}

func TestUpdateFunction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	returned := []byte(`{"space":"default","functionId":"testid","type":"http","provider":{"url":"http://example.com"}}`)
	payload := []byte(`{"space":"default","functionId":"testid","type":"http","provider":{"url":"http://example1.com"}}`)
	fn := &function.Function{
		ID:           "testid",
		Space:        "default",
		ProviderType: http.Type,
		Provider:     &http.HTTP{URL: "http://example.com"},
	}

	t.Run("function updated", func(t *testing.T) {
		db := mock.NewMockStore(ctrl)
		db.EXPECT().Get("default/testid", &store.ReadOptions{Consistent: true}).Return(&store.KVPair{Value: returned}, nil)
		db.EXPECT().Put("default/testid", payload, nil).Return(nil)
		service := &Service{FunctionStore: db, Log: zap.NewNop()}

		fn = &function.Function{
			ID:           "testid",
			Space:        "default",
			ProviderType: http.Type,
			Provider:     &http.HTTP{URL: "http://example1.com"},
		}
		_, err := service.UpdateFunction(fn)

		assert.Nil(t, err)
	})

	t.Run("function not found", func(t *testing.T) {
		db := mock.NewMockStore(ctrl)
		db.EXPECT().Get("default/testid", gomock.Any()).Return(nil, errors.New("KV not found"))
		service := &Service{FunctionStore: db, Log: zap.NewNop()}

		_, err := service.UpdateFunction(fn)

		assert.Equal(t, err, &function.ErrFunctionNotFound{ID: "testid"})
	})

	t.Run("KV Put error", func(t *testing.T) {
		db := mock.NewMockStore(ctrl)
		db.EXPECT().Get("default/testid", gomock.Any()).Return(&store.KVPair{Value: returned}, nil)
		db.EXPECT().Put("default/testid", payload, nil).Return(errors.New("KV put error"))
		service := &Service{FunctionStore: db, Log: zap.NewNop()}

		_, err := service.UpdateFunction(fn)

		assert.EqualError(t, err, "KV put error")
	})
}

func TestGetFunction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("function returned", func(t *testing.T) {
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
	})

	t.Run("function not found", func(t *testing.T) {
		db := mock.NewMockStore(ctrl)
		db.EXPECT().Get("default/testid", gomock.Any()).Return(nil, errors.New("Key not found in store"))
		service := &Service{FunctionStore: db, Log: zap.NewNop()}

		_, err := service.GetFunction("default", function.ID("testid"))

		assert.Equal(t, err, &function.ErrFunctionNotFound{ID: "testid"})
	})

	t.Run("KV Get error", func(t *testing.T) {
		db := mock.NewMockStore(ctrl)
		db.EXPECT().Get("default/testid", gomock.Any()).Return(nil, errors.New("KV get err"))
		service := &Service{FunctionStore: db, Log: zap.NewNop()}

		_, err := service.GetFunction("default", function.ID("testid"))

		assert.EqualError(t, err, "KV get err")
	})
}

func TestListFunctions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("list returned", func(t *testing.T) {
		kvs := []*store.KVPair{
			&store.KVPair{Value: []byte(`{"functionId":"f1","type":"http","provider":{"url": "http://test.com"}}}`)},
		}
		db := mock.NewMockStore(ctrl)
		db.EXPECT().List("default/", &store.ReadOptions{Consistent: true}).Return(kvs, nil)
		service := &Service{FunctionStore: db, Log: zap.NewNop()}

		list, err := service.ListFunctions("default")

		assert.Nil(t, err)
		assert.Equal(t, function.Functions{{
			ID:           function.ID("f1"),
			ProviderType: http.Type,
			Provider:     &http.HTTP{URL: "http://test.com"},
		}}, list)
	})

	t.Run("filtered list returned", func(t *testing.T) {
		kvs := []*store.KVPair{
			&store.KVPair{Value: []byte(`{"functionId":"f1","type":"http","provider":{"url": "http://test.com"},"metadata":{"key1":"val1"}}`)},
			&store.KVPair{Value: []byte(`{"functionId":"f2","type":"http","provider":{"url": "http://test.com"}}`)},
		}
		db := mock.NewMockStore(ctrl)
		db.EXPECT().List("default/", &store.ReadOptions{Consistent: true}).Return(kvs, nil)
		service := &Service{FunctionStore: db, Log: zap.NewNop()}

		list, err := service.ListFunctions("default", metadata.Filter{Key: "key1", Value: "val1"})

		assert.Nil(t, err)
		assert.Equal(t, function.Functions{{
			ID:           function.ID("f1"),
			ProviderType: http.Type,
			Provider:     &http.HTTP{URL: "http://test.com"},
			Metadata:     metadata.Metadata{"key1": "val1"},
		}}, list)
	})

	t.Run("KV List error", func(t *testing.T) {
		db := mock.NewMockStore(ctrl)
		db.EXPECT().List("default/", gomock.Any()).Return([]*store.KVPair{}, errors.New("KV list err"))
		service := &Service{FunctionStore: db, Log: zap.NewNop()}

		_, err := service.ListFunctions("default")

		assert.EqualError(t, err, "KV list err")
	})

	t.Run("KV key not found", func(t *testing.T) {
		db := mock.NewMockStore(ctrl)
		db.EXPECT().List("default/", gomock.Any()).Return([]*store.KVPair{}, errors.New("Key not found in store"))
		service := &Service{FunctionStore: db, Log: zap.NewNop()}

		list, _ := service.ListFunctions("default")

		assert.Equal(t, function.Functions{}, list)
	})
}

func TestDeleteFunction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("function deleted", func(t *testing.T) {
		kvs := []*store.KVPair{}
		subscriptionsDB := mock.NewMockStore(ctrl)
		subscriptionsDB.EXPECT().List("default/", &store.ReadOptions{Consistent: true}).Return(kvs, nil)
		eventTypesDB := mock.NewMockStore(ctrl)
		eventTypesDB.EXPECT().List("default/", &store.ReadOptions{Consistent: true}).Return(kvs, nil)
		functionsDB := mock.NewMockStore(ctrl)
		functionsDB.EXPECT().Delete("default/testid").Return(nil)
		service := &Service{FunctionStore: functionsDB, SubscriptionStore: subscriptionsDB, EventTypeStore: eventTypesDB, Log: zap.NewNop()}

		err := service.DeleteFunction("default", function.ID("testid"))

		assert.Nil(t, err)
	})

	t.Run("function not found", func(t *testing.T) {
		kvs := []*store.KVPair{}
		subscriptionsDB := mock.NewMockStore(ctrl)
		subscriptionsDB.EXPECT().List("default/", &store.ReadOptions{Consistent: true}).Return(kvs, nil)
		eventTypesDB := mock.NewMockStore(ctrl)
		eventTypesDB.EXPECT().List("default/", &store.ReadOptions{Consistent: true}).Return(kvs, nil)
		functionsDB := mock.NewMockStore(ctrl)
		functionsDB.EXPECT().Delete("default/testid").Return(errors.New("KV func not found"))
		service := &Service{FunctionStore: functionsDB, SubscriptionStore: subscriptionsDB, EventTypeStore: eventTypesDB, Log: zap.NewNop()}

		err := service.DeleteFunction("default", function.ID("testid"))

		assert.EqualError(t, err, `Function "testid" not found.`)
	})

	t.Run("function has subscriptions", func(t *testing.T) {
		kvs := []*store.KVPair{
			{Value: []byte(`{"subscriptionId":"s1","default":"default","event":"test","functionId":"testid"}`)}}
		subscriptionsDB := mock.NewMockStore(ctrl)
		subscriptionsDB.EXPECT().List("default/", &store.ReadOptions{Consistent: true}).Return(kvs, nil)
		functionsDB := mock.NewMockStore(ctrl)
		service := &Service{FunctionStore: functionsDB, SubscriptionStore: subscriptionsDB, Log: zap.NewNop()}

		err := service.DeleteFunction("default", function.ID("testid"))

		assert.Equal(t, err, &function.ErrFunctionHasSubscriptions{})
	})

	t.Run("function is authorizer", func(t *testing.T) {
		kvs := []*store.KVPair{
			{Value: []byte(`{"name":"test.event.noauth"}`)},
			{Value: []byte(`{"name":"test.event","authorizerId":"testid"}`)}}
		eventTypesDB := mock.NewMockStore(ctrl)
		eventTypesDB.EXPECT().List("default/", &store.ReadOptions{Consistent: true}).Return(kvs, nil)
		subscriptionsDB := mock.NewMockStore(ctrl)
		subscriptionsDB.EXPECT().List("default/", &store.ReadOptions{Consistent: true}).Return([]*store.KVPair{}, nil)
		functionsDB := mock.NewMockStore(ctrl)
		service := &Service{FunctionStore: functionsDB, SubscriptionStore: subscriptionsDB, EventTypeStore: eventTypesDB, Log: zap.NewNop()}

		err := service.DeleteFunction("default", function.ID("testid"))

		assert.Equal(t, err, &function.ErrFunctionIsAuthorizer{ID: function.ID("testid"), EventType: "test.event"})
	})
}

func TestValidateFunction(t *testing.T) {
	t.Run("missing ID", func(t *testing.T) {
		err := validateFunction(&function.Function{})

		assert.Equal(t, err, &function.ErrFunctionValidation{
			Message: "Key: 'Function.ID' Error:Field validation for 'ID' failed on the 'required' tag"})
	})

	t.Run("invalid space", func(t *testing.T) {
		fn := &function.Function{
			ID:    "id",
			Space: "///"}
		err := validateFunction(fn)

		assert.Equal(t, err, &function.ErrFunctionValidation{
			Message: "Key: 'Function.Space' Error:Field validation for 'Space' failed on the 'space' tag"})
	})

	t.Run("set default space", func(t *testing.T) {
		fn := &function.Function{
			ID:           "id",
			ProviderType: http.Type,
			Provider:     http.HTTP{URL: "http://example.com"},
		}
		validateFunction(fn)

		assert.Equal(t, "default", fn.Space)
	})
}
