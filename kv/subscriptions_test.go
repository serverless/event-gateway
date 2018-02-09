package kv

import (
	"errors"
	"testing"

	"go.uber.org/zap"

	"github.com/golang/mock/gomock"
	"github.com/serverless/event-gateway/api"
	"github.com/serverless/event-gateway/mock"
	"github.com/serverless/libkv/store"
	"github.com/stretchr/testify/assert"
)

func TestCreateSubscription_HTTPOK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	subscriptionsDB := mock.NewMockStore(ctrl)
	subscriptionsDB.EXPECT().Get("http,GET,%2F", &store.ReadOptions{Consistent: true}).Return(nil, errors.New("KV sub not found"))
	subscriptionsDB.EXPECT().Put("http,GET,%2F", []byte(`{"subscriptionId":"http,GET,%2F","event":"http","functionId":"func","method":"GET","path":"/"}`), nil).Return(nil)
	endpointsDB := mock.NewMockStore(ctrl)
	endpointsDB.EXPECT().Put("GET,%2F", []byte(`{"endpointId":"GET,%2F"}`), nil).Return(nil)
	endpointsDB.EXPECT().List("", &store.ReadOptions{Consistent: true}).Return([]*store.KVPair{}, nil)
	functionsDB := mock.NewMockStore(ctrl)
	functionsDB.EXPECT().Exists("func", &store.ReadOptions{Consistent: true}).Return(true, nil)
	subs := &Subscriptions{SubscriptionsDB: subscriptionsDB, EndpointsDB: endpointsDB, FunctionsDB: functionsDB, Log: zap.NewNop()}

	_, err := subs.CreateSubscription(&api.Subscription{ID: "testid", Event: "http", FunctionID: "func", Method: "GET", Path: "/"})

	assert.Nil(t, err)
}

func TestCreateSubscription_HTTPValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	subs := &Subscriptions{Log: zap.NewNop()}

	_, err := subs.CreateSubscription(&api.Subscription{ID: "testid", Event: "http", FunctionID: "func"})

	assert.Equal(t, err, &ErrSubscriptionValidation{original: "Missing required fields (method, path) for HTTP event."})
}

func TestCreateSubscription_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	subscriptionsDB := mock.NewMockStore(ctrl)
	subscriptionsDB.EXPECT().Get("test,func,%2F", &store.ReadOptions{Consistent: true}).Return(nil, errors.New("KV sub not found"))
	subscriptionsDB.EXPECT().Put("test,func,%2F", []byte(`{"subscriptionId":"test,func,%2F","event":"test","functionId":"func","path":"/"}`), nil).Return(nil)
	functionsDB := mock.NewMockStore(ctrl)
	functionsDB.EXPECT().Exists("func", &store.ReadOptions{Consistent: true}).Return(true, nil)
	subs := &Subscriptions{SubscriptionsDB: subscriptionsDB, FunctionsDB: functionsDB, Log: zap.NewNop()}

	_, err := subs.CreateSubscription(&api.Subscription{ID: "testid", Event: "test", FunctionID: "func"})

	assert.Nil(t, err)
}

func TestCreateSubscription_ValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	subs := &Subscriptions{Log: zap.NewNop()}

	_, err := subs.CreateSubscription(&api.Subscription{})

	assert.Equal(t, err, &ErrSubscriptionValidation{original: "Key: 'Subscription.Event' Error:Field validation for 'Event' failed on the 'required' tag\nKey: 'Subscription.FunctionID' Error:Field validation for 'FunctionID' failed on the 'required' tag"})
}

func TestCreateSubscription_AlreadyExistsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	subscriptionsDB := mock.NewMockStore(ctrl)
	subscriptionsDB.EXPECT().Get("test,func,%2F", gomock.Any()).Return(&store.KVPair{Value: []byte(`{"subscriptionId":"testid"}`)}, nil)
	subs := &Subscriptions{SubscriptionsDB: subscriptionsDB, Log: zap.NewNop()}

	_, err := subs.CreateSubscription(&api.Subscription{ID: "testid", Event: "test", FunctionID: "func"})

	assert.Equal(t, err, &ErrSubscriptionAlreadyExists{ID: "test,func,%2F"})
}

func TestCreateSubscription_EndpointPathConflictError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	kv := &store.KVPair{Value: []byte(`{"endpointId":"GET,%2F:name","functionId":"func","method":"GET","path":"/:name"}`)}
	subscriptionsDB := mock.NewMockStore(ctrl)
	subscriptionsDB.EXPECT().Get("http,GET,%2F:id", gomock.Any()).Return(nil, errors.New("KV sub not found"))
	endpointsDB := mock.NewMockStore(ctrl)
	endpointsDB.EXPECT().List("", gomock.Any()).Return([]*store.KVPair{kv}, nil)
	subs := &Subscriptions{SubscriptionsDB: subscriptionsDB, EndpointsDB: endpointsDB, Log: zap.NewNop()}

	_, err := subs.CreateSubscription(&api.Subscription{ID: "testid", Event: "http", FunctionID: "func", Method: "GET", Path: "/:id"})

	assert.Equal(t, err, &ErrPathConfict{`parameter with different name ("name") already defined: for route: /:id`})
}

func TestCreateSubscription_EndpointPutError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	subscriptionsDB := mock.NewMockStore(ctrl)
	subscriptionsDB.EXPECT().Get("http,GET,%2F", gomock.Any()).Return(nil, errors.New("KV sub not found"))
	endpointsDB := mock.NewMockStore(ctrl)
	endpointsDB.EXPECT().List("", gomock.Any()).Return([]*store.KVPair{}, nil)
	endpointsDB.EXPECT().Put("GET,%2F", []byte(`{"endpointId":"GET,%2F"}`), nil).Return(errors.New("KV Put err"))
	subs := &Subscriptions{SubscriptionsDB: subscriptionsDB, EndpointsDB: endpointsDB, Log: zap.NewNop()}

	_, err := subs.CreateSubscription(&api.Subscription{ID: "testid", Event: "http", FunctionID: "func", Method: "GET", Path: "/"})

	assert.EqualError(t, err, "KV Put err")
}

func TestCreateSubscription_EndpointListError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	subscriptionsDB := mock.NewMockStore(ctrl)
	subscriptionsDB.EXPECT().Get("http,GET,%2F", gomock.Any()).Return(nil, errors.New("KV sub not found"))
	endpointsDB := mock.NewMockStore(ctrl)
	endpointsDB.EXPECT().List("", gomock.Any()).Return(nil, errors.New("KV List err"))
	subs := &Subscriptions{SubscriptionsDB: subscriptionsDB, EndpointsDB: endpointsDB, Log: zap.NewNop()}

	_, err := subs.CreateSubscription(&api.Subscription{ID: "testid", Event: "http", FunctionID: "func", Method: "GET", Path: "/"})

	assert.EqualError(t, err, "KV List err")
}

func TestCreateSubscription_FunctionExistsKVError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	subscriptionsDB := mock.NewMockStore(ctrl)
	subscriptionsDB.EXPECT().Get("http,GET,%2F", gomock.Any()).Return(nil, errors.New("KV sub not found"))
	endpointsDB := mock.NewMockStore(ctrl)
	endpointsDB.EXPECT().List("", gomock.Any()).Return([]*store.KVPair{}, nil)
	endpointsDB.EXPECT().Put("GET,%2F", []byte(`{"endpointId":"GET,%2F"}`), nil).Return(nil)
	functionsDB := mock.NewMockStore(ctrl)
	functionsDB.EXPECT().Exists("func", gomock.Any()).Return(false, errors.New("KV Exists err"))
	subs := &Subscriptions{SubscriptionsDB: subscriptionsDB, EndpointsDB: endpointsDB, FunctionsDB: functionsDB, Log: zap.NewNop()}

	_, err := subs.CreateSubscription(&api.Subscription{ID: "testid", Event: "http", FunctionID: "func", Method: "GET", Path: "/"})

	assert.EqualError(t, err, "KV Exists err")
}

func TestCreateSubscription_FunctionExistsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	subscriptionsDB := mock.NewMockStore(ctrl)
	subscriptionsDB.EXPECT().Get("http,GET,%2F", gomock.Any()).Return(nil, errors.New("KV sub not found"))
	endpointsDB := mock.NewMockStore(ctrl)
	endpointsDB.EXPECT().List("", gomock.Any()).Return([]*store.KVPair{}, nil)
	endpointsDB.EXPECT().Put("GET,%2F", []byte(`{"endpointId":"GET,%2F"}`), nil).Return(nil)
	functionsDB := mock.NewMockStore(ctrl)
	functionsDB.EXPECT().Exists("func", gomock.Any()).Return(false, nil)
	subs := &Subscriptions{SubscriptionsDB: subscriptionsDB, EndpointsDB: endpointsDB, FunctionsDB: functionsDB, Log: zap.NewNop()}

	_, err := subs.CreateSubscription(&api.Subscription{ID: "testid", Event: "http", FunctionID: "func", Method: "GET", Path: "/"})

	assert.Equal(t, err, &ErrFunctionNotFound{functionID: "func"})
}

func TestCreateSubscription_PutError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	subscriptionsDB := mock.NewMockStore(ctrl)
	subscriptionsDB.EXPECT().Get("http,GET,%2F", gomock.Any()).Return(nil, errors.New("KV sub not found"))
	subscriptionsDB.EXPECT().Put("http,GET,%2F", []byte(`{"subscriptionId":"http,GET,%2F","event":"http","functionId":"func","method":"GET","path":"/"}`), nil).Return(errors.New("KV Put err"))
	endpointsDB := mock.NewMockStore(ctrl)
	endpointsDB.EXPECT().List("", gomock.Any()).Return([]*store.KVPair{}, nil)
	endpointsDB.EXPECT().Put("GET,%2F", []byte(`{"endpointId":"GET,%2F"}`), nil).Return(nil)
	functionsDB := mock.NewMockStore(ctrl)
	functionsDB.EXPECT().Exists("func", gomock.Any()).Return(true, nil)
	subs := &Subscriptions{SubscriptionsDB: subscriptionsDB, EndpointsDB: endpointsDB, FunctionsDB: functionsDB, Log: zap.NewNop()}

	_, err := subs.CreateSubscription(&api.Subscription{ID: "testid", Event: "http", FunctionID: "func", Method: "GET", Path: "/"})

	assert.EqualError(t, err, "KV Put err")
}

func TestDeleteSubscription_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	kv := &store.KVPair{Value: []byte(`{"subscriptionId":"testid","event":"test","functionId":"f1"}`)}
	subscriptionsDB := mock.NewMockStore(ctrl)
	subscriptionsDB.EXPECT().Get("testid", gomock.Any()).Return(kv, nil)
	subscriptionsDB.EXPECT().Delete("testid").Return(nil)
	subs := &Subscriptions{SubscriptionsDB: subscriptionsDB, Log: zap.NewNop()}

	err := subs.DeleteSubscription(api.SubscriptionID("testid"))

	assert.Nil(t, err)
}

func TestDeleteSubscription_GetError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	subscriptionsDB := mock.NewMockStore(ctrl)
	subscriptionsDB.EXPECT().Get("testid", gomock.Any()).Return(nil, errors.New("KV Get err"))
	subs := &Subscriptions{SubscriptionsDB: subscriptionsDB, Log: zap.NewNop()}

	err := subs.DeleteSubscription(api.SubscriptionID("testid"))

	assert.Equal(t, err, &ErrSubscriptionNotFound{"testid"})
}

func TestDeleteSubscription_DeleteError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	kv := &store.KVPair{Value: []byte(`{"subscriptionId":"testid","event":"test","functionId":"f1"}`)}
	subscriptionsDB := mock.NewMockStore(ctrl)
	subscriptionsDB.EXPECT().Get("testid", gomock.Any()).Return(kv, nil)
	subscriptionsDB.EXPECT().Delete("testid").Return(errors.New("KV Delete err"))
	subs := &Subscriptions{SubscriptionsDB: subscriptionsDB, Log: zap.NewNop()}

	err := subs.DeleteSubscription(api.SubscriptionID("testid"))

	assert.Equal(t, err, &ErrSubscriptionNotFound{"testid"})
}

func TestDeleteSubscription_DeleteEndpointOK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	kv := &store.KVPair{Value: []byte(`{"subscriptionId":"testid","event":"http","functionId":"f1","method":"GET","path":"/"}`)}
	subscriptionsDB := mock.NewMockStore(ctrl)
	subscriptionsDB.EXPECT().Get("testid", gomock.Any()).Return(kv, nil)
	subscriptionsDB.EXPECT().Delete("testid").Return(nil)
	endpointsDB := mock.NewMockStore(ctrl)
	endpointsDB.EXPECT().Delete("GET,%2F").Return(nil)
	subs := &Subscriptions{SubscriptionsDB: subscriptionsDB, EndpointsDB: endpointsDB, Log: zap.NewNop()}

	err := subs.DeleteSubscription(api.SubscriptionID("testid"))

	assert.Nil(t, err)
}

func TestDeleteSubscription_DeleteEndpointError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	kv := &store.KVPair{Value: []byte(`{"subscriptionId":"testid","event":"http","functionId":"f1","method":"GET","path":"/"}`)}
	subscriptionsDB := mock.NewMockStore(ctrl)
	subscriptionsDB.EXPECT().Get("testid", gomock.Any()).Return(kv, nil)
	subscriptionsDB.EXPECT().Delete("testid").Return(nil)
	endpointsDB := mock.NewMockStore(ctrl)
	endpointsDB.EXPECT().Delete("GET,%2F").Return(errors.New("KV Delete err"))
	subs := &Subscriptions{SubscriptionsDB: subscriptionsDB, EndpointsDB: endpointsDB, Log: zap.NewNop()}

	err := subs.DeleteSubscription(api.SubscriptionID("testid"))

	assert.EqualError(t, err, "KV Delete err")
}

func TestGetAllSubscriptions_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	kvs := []*store.KVPair{
		{Value: []byte(`{"subscriptionId":"s1","event":"test","functionId":"f1"}`)},
		{Value: []byte(`{"subscriptionId":"s2","event":"test","functionId":"f2"}`)},
	}
	subscriptionsDB := mock.NewMockStore(ctrl)
	subscriptionsDB.EXPECT().List("", &store.ReadOptions{Consistent: true}).Return(kvs, nil)
	subs := &Subscriptions{SubscriptionsDB: subscriptionsDB, Log: zap.NewNop()}

	list, _ := subs.GetAllSubscriptions()

	assert.Equal(t, []*api.Subscription{
		{ID: api.SubscriptionID("s1"), Event: "test", FunctionID: api.FunctionID("f1")},
		{ID: api.SubscriptionID("s2"), Event: "test", FunctionID: api.FunctionID("f2")},
	}, list)
}

func TestGetAllSubscriptions_ListError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	subscriptionsDB := mock.NewMockStore(ctrl)
	subscriptionsDB.EXPECT().List("", gomock.Any()).Return(nil, errors.New("KV error"))
	subs := &Subscriptions{SubscriptionsDB: subscriptionsDB, Log: zap.NewNop()}

	_, err := subs.GetAllSubscriptions()
	assert.EqualError(t, err, "KV error")
}

func TestIsPathInConflict(t *testing.T) {
	assert.False(t, isPathInConflict("/foo", "/foo"))
	assert.False(t, isPathInConflict("/foo", "/bar/baz"))

	assert.True(t, isPathInConflict("/foo", "/:bar"))
	assert.True(t, isPathInConflict("/:foo", "/bar"))
	assert.True(t, isPathInConflict("/:foo", "/:bar"))
	assert.True(t, isPathInConflict("/:foo/:bar", "/baz"))
	assert.True(t, isPathInConflict("/a/b/c/d", "/:b"))
	assert.False(t, isPathInConflict("/:a", "/:a/b"))
	assert.True(t, isPathInConflict("/foo/:bar", "/foo/bar/baz"))
	assert.True(t, isPathInConflict("/:foo/bar/baz", "/foo/:bar"))

	assert.True(t, isPathInConflict("/*foo", "/*bar"))
	assert.True(t, isPathInConflict("/*foo", "/bar"))
	assert.True(t, isPathInConflict("/*foo", "/:bar"))
	assert.True(t, isPathInConflict("/:foo", "/*bar"))
}
