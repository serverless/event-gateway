package libkv

import (
	"errors"
	"testing"

	"go.uber.org/zap"

	"github.com/golang/mock/gomock"
	"github.com/serverless/event-gateway/function"
	"github.com/serverless/event-gateway/mock"
	"github.com/serverless/event-gateway/subscription"
	"github.com/serverless/libkv/store"
	"github.com/stretchr/testify/assert"
)

func TestCreateSubscription_HTTPOK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	subscriptionsDB := mock.NewMockStore(ctrl)
	subscriptionsDB.EXPECT().Get("default/http,GET,%2F", &store.ReadOptions{Consistent: true}).Return(nil, errors.New("KV sub not found"))
	subscriptionsDB.EXPECT().Put("default/http,GET,%2F", []byte(`{"space":"default","subscriptionId":"http,GET,%2F","event":"http","functionId":"func","method":"GET","path":"/"}`), nil).Return(nil)
	endpointsDB := mock.NewMockStore(ctrl)
	endpointsDB.EXPECT().Put("default/GET,%2F", []byte(`{"endpointId":"GET,%2F"}`), nil).Return(nil)
	endpointsDB.EXPECT().List("default/", &store.ReadOptions{Consistent: true}).Return([]*store.KVPair{}, nil)
	functionsDB := mock.NewMockStore(ctrl)
	functionsDB.EXPECT().Get("default/func", &store.ReadOptions{Consistent: true}).Return(&store.KVPair{Value: []byte(`{"functionId":"func"}`)}, nil)
	subs := &Service{SubscriptionStore: subscriptionsDB, EndpointStore: endpointsDB, FunctionStore: functionsDB, Log: zap.NewNop()}

	_, err := subs.CreateSubscription(&subscription.Subscription{ID: "testid", Event: "http", FunctionID: "func", Method: "GET", Path: "/"})

	assert.Nil(t, err)
}

func TestCreateSubscription_HTTPValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	subs := &Service{Log: zap.NewNop()}

	_, err := subs.CreateSubscription(&subscription.Subscription{ID: "testid", Event: "http", FunctionID: "func"})

	assert.Equal(t, err, &subscription.ErrSubscriptionValidation{Message: "Missing required fields (method, path) for HTTP event."})
}

func TestCreateSubscription_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	subscriptionsDB := mock.NewMockStore(ctrl)
	subscriptionsDB.EXPECT().Get("default/test,func,%2F", &store.ReadOptions{Consistent: true}).Return(nil, errors.New("KV sub not found"))
	subscriptionsDB.EXPECT().Put("default/test,func,%2F", []byte(`{"space":"default","subscriptionId":"test,func,%2F","event":"test","functionId":"func","path":"/"}`), nil).Return(nil)
	functionsDB := mock.NewMockStore(ctrl)
	functionsDB.EXPECT().Get("default/func", &store.ReadOptions{Consistent: true}).Return(&store.KVPair{Value: []byte(`{"functionId":"func"}`)}, nil)
	subs := &Service{SubscriptionStore: subscriptionsDB, FunctionStore: functionsDB, Log: zap.NewNop()}

	_, err := subs.CreateSubscription(&subscription.Subscription{ID: "testid", Event: "test", FunctionID: "func"})

	assert.Nil(t, err)
}

func TestCreateSubscription_ValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	subs := &Service{Log: zap.NewNop()}

	_, err := subs.CreateSubscription(&subscription.Subscription{})

	assert.Equal(t, err, &subscription.ErrSubscriptionValidation{Message: "Key: 'Subscription.Event' Error:Field validation for 'Event' failed on the 'required' tag\nKey: 'Subscription.FunctionID' Error:Field validation for 'FunctionID' failed on the 'required' tag"})
}

func TestCreateSubscription_AlreadyExistsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	subscriptionsDB := mock.NewMockStore(ctrl)
	subscriptionsDB.EXPECT().Get("default/test,func,%2F", gomock.Any()).Return(&store.KVPair{Value: []byte(`{"subscriptionId":"testid"}`)}, nil)
	subs := &Service{SubscriptionStore: subscriptionsDB, Log: zap.NewNop()}

	_, err := subs.CreateSubscription(&subscription.Subscription{ID: "testid", Event: "test", FunctionID: "func"})

	assert.Equal(t, err, &subscription.ErrSubscriptionAlreadyExists{ID: "test,func,%2F"})
}

func TestCreateSubscription_EndpointPathConflictError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	subscriptionsDB := mock.NewMockStore(ctrl)
	subscriptionsDB.EXPECT().Get("default/http,GET,%2F:id", gomock.Any()).Return(nil, errors.New("KV sub not found"))
	endpointsDB := mock.NewMockStore(ctrl)
	kv := &store.KVPair{Value: []byte(`{"endpointId":"GET,%2F:name","functionId":"func","method":"GET","path":"/:name"}`)}
	endpointsDB.EXPECT().List("default/", gomock.Any()).Return([]*store.KVPair{kv}, nil)
	subs := &Service{SubscriptionStore: subscriptionsDB, EndpointStore: endpointsDB, Log: zap.NewNop()}

	_, err := subs.CreateSubscription(&subscription.Subscription{ID: "testid", Space: "default", Event: "http", FunctionID: "func", Method: "GET", Path: "/:id"})

	assert.Equal(t, err, &subscription.ErrPathConfict{Message: `parameter with different name ("name") already defined: for route: /:id`})
}

func TestCreateSubscription_EndpointPutError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	subscriptionsDB := mock.NewMockStore(ctrl)
	subscriptionsDB.EXPECT().Get("default/http,GET,%2F", gomock.Any()).Return(nil, errors.New("KV sub not found"))
	endpointsDB := mock.NewMockStore(ctrl)
	endpointsDB.EXPECT().List("default/", gomock.Any()).Return([]*store.KVPair{}, nil)
	endpointsDB.EXPECT().Put("default/GET,%2F", []byte(`{"endpointId":"GET,%2F"}`), nil).Return(errors.New("KV Put err"))
	subs := &Service{SubscriptionStore: subscriptionsDB, EndpointStore: endpointsDB, Log: zap.NewNop()}

	_, err := subs.CreateSubscription(&subscription.Subscription{ID: "testid", Event: "http", FunctionID: "func", Method: "GET", Path: "/"})

	assert.EqualError(t, err, "KV Put err")
}

func TestCreateSubscription_EndpointListError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	subscriptionsDB := mock.NewMockStore(ctrl)
	subscriptionsDB.EXPECT().Get("default/http,GET,%2F", gomock.Any()).Return(nil, errors.New("KV sub not found"))
	endpointsDB := mock.NewMockStore(ctrl)
	endpointsDB.EXPECT().List("default/", gomock.Any()).Return(nil, errors.New("KV List err"))
	subs := &Service{SubscriptionStore: subscriptionsDB, EndpointStore: endpointsDB, Log: zap.NewNop()}

	_, err := subs.CreateSubscription(&subscription.Subscription{ID: "testid", Event: "http", FunctionID: "func", Method: "GET", Path: "/"})

	assert.EqualError(t, err, "KV List err")
}

func TestCreateSubscription_GetFunctionKVError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	subscriptionsDB := mock.NewMockStore(ctrl)
	subscriptionsDB.EXPECT().Get("default/http,GET,%2F", gomock.Any()).Return(nil, errors.New("KV sub not found"))
	endpointsDB := mock.NewMockStore(ctrl)
	endpointsDB.EXPECT().List("default/", gomock.Any()).Return([]*store.KVPair{}, nil)
	endpointsDB.EXPECT().Put("default/GET,%2F", []byte(`{"endpointId":"GET,%2F"}`), nil).Return(nil)
	functionsDB := mock.NewMockStore(ctrl)
	functionsDB.EXPECT().Get("default/func", &store.ReadOptions{Consistent: true}).Return(nil, errors.New("Key not found in store"))
	subs := &Service{SubscriptionStore: subscriptionsDB, EndpointStore: endpointsDB, FunctionStore: functionsDB, Log: zap.NewNop()}

	_, err := subs.CreateSubscription(&subscription.Subscription{ID: "testid", Event: "http", FunctionID: "func", Method: "GET", Path: "/"})

	assert.EqualError(t, err, "Function \"func\" not found.")
}

func TestCreateSubscription_PutError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	subscriptionsDB := mock.NewMockStore(ctrl)
	subscriptionsDB.EXPECT().Get("default/http,GET,%2F", gomock.Any()).Return(nil, errors.New("KV sub not found"))
	subscriptionsDB.EXPECT().Put("default/http,GET,%2F", []byte(`{"space":"default","subscriptionId":"http,GET,%2F","event":"http","functionId":"func","method":"GET","path":"/"}`), nil).Return(errors.New("KV Put err"))
	endpointsDB := mock.NewMockStore(ctrl)
	endpointsDB.EXPECT().List("default/", gomock.Any()).Return([]*store.KVPair{}, nil)
	endpointsDB.EXPECT().Put("default/GET,%2F", []byte(`{"endpointId":"GET,%2F"}`), nil).Return(nil)
	functionsDB := mock.NewMockStore(ctrl)
	functionsDB.EXPECT().Get("default/func", gomock.Any()).Return(&store.KVPair{Value: []byte(`{"functionId":"func"}`)}, nil)
	subs := &Service{SubscriptionStore: subscriptionsDB, EndpointStore: endpointsDB, FunctionStore: functionsDB, Log: zap.NewNop()}

	_, err := subs.CreateSubscription(&subscription.Subscription{ID: "testid", Event: "http", FunctionID: "func", Method: "GET", Path: "/"})

	assert.EqualError(t, err, "KV Put err")
}

func TestDeleteSubscription_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	kv := &store.KVPair{Value: []byte(`{"subscriptionId":"testid","event":"test","functionId":"f1"}`)}
	subscriptionsDB := mock.NewMockStore(ctrl)
	subscriptionsDB.EXPECT().Get("default/testid", gomock.Any()).Return(kv, nil)
	subscriptionsDB.EXPECT().Delete("default/testid").Return(nil)
	subs := &Service{SubscriptionStore: subscriptionsDB, Log: zap.NewNop()}

	err := subs.DeleteSubscription("default", subscription.ID("testid"))

	assert.Nil(t, err)
}

func TestDeleteSubscription_GetError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	subscriptionsDB := mock.NewMockStore(ctrl)
	subscriptionsDB.EXPECT().Get("default/testid", gomock.Any()).Return(nil, errors.New("KV Get err"))
	subs := &Service{SubscriptionStore: subscriptionsDB, Log: zap.NewNop()}

	err := subs.DeleteSubscription("default", subscription.ID("testid"))

	assert.Equal(t, err, &subscription.ErrSubscriptionNotFound{ID: "testid"})
}

func TestDeleteSubscription_DeleteError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	kv := &store.KVPair{Value: []byte(`{"subscriptionId":"testid","event":"test","functionId":"f1"}`)}
	subscriptionsDB := mock.NewMockStore(ctrl)
	subscriptionsDB.EXPECT().Get("default/testid", gomock.Any()).Return(kv, nil)
	subscriptionsDB.EXPECT().Delete("default/testid").Return(errors.New("KV Delete err"))
	subs := &Service{SubscriptionStore: subscriptionsDB, Log: zap.NewNop()}

	err := subs.DeleteSubscription("default", subscription.ID("testid"))

	assert.Equal(t, err, &subscription.ErrSubscriptionNotFound{ID: "testid"})
}

func TestDeleteSubscription_DeleteEndpointOK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	kv := &store.KVPair{Value: []byte(`{"subscriptionId":"testid","event":"http","functionId":"f1","method":"GET","path":"/"}`)}
	subscriptionsDB := mock.NewMockStore(ctrl)
	subscriptionsDB.EXPECT().Get("default/testid", gomock.Any()).Return(kv, nil)
	subscriptionsDB.EXPECT().Delete("default/testid").Return(nil)
	endpointsDB := mock.NewMockStore(ctrl)
	endpointsDB.EXPECT().Delete("default/GET,%2F").Return(nil)
	subs := &Service{SubscriptionStore: subscriptionsDB, EndpointStore: endpointsDB, Log: zap.NewNop()}

	err := subs.DeleteSubscription("default", subscription.ID("testid"))

	assert.Nil(t, err)
}

func TestDeleteSubscription_DeleteEndpointError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	kv := &store.KVPair{Value: []byte(`{"subscriptionId":"testid","event":"http","functionId":"f1","method":"GET","path":"/"}`)}
	subscriptionsDB := mock.NewMockStore(ctrl)
	subscriptionsDB.EXPECT().Get("default/testid", gomock.Any()).Return(kv, nil)
	subscriptionsDB.EXPECT().Delete("default/testid").Return(nil)
	endpointsDB := mock.NewMockStore(ctrl)
	endpointsDB.EXPECT().Delete("default/GET,%2F").Return(errors.New("KV Delete err"))
	subs := &Service{SubscriptionStore: subscriptionsDB, EndpointStore: endpointsDB, Log: zap.NewNop()}

	err := subs.DeleteSubscription("default", subscription.ID("testid"))

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
	subs := &Service{SubscriptionStore: subscriptionsDB, Log: zap.NewNop()}

	list, _ := subs.GetSubscriptions("")

	assert.Equal(t, subscription.Subscriptions{
		{ID: subscription.ID("s1"), Event: "test", FunctionID: function.ID("f1")},
		{ID: subscription.ID("s2"), Event: "test", FunctionID: function.ID("f2")},
	}, list)
}

func TestGetAllSubscriptions_ListError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	subscriptionsDB := mock.NewMockStore(ctrl)
	subscriptionsDB.EXPECT().List("", gomock.Any()).Return(nil, errors.New("KV error"))
	subs := &Service{SubscriptionStore: subscriptionsDB, Log: zap.NewNop()}

	_, err := subs.GetSubscriptions("")
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
