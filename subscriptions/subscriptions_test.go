package subscriptions

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/serverless/event-gateway/functions"
	"github.com/serverless/event-gateway/subscriptions/mock"
	"github.com/serverless/libkv/store"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestCreateSubscription_HTTPOK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	subscriptionsDB := mock.NewMockStore(ctrl)
	subscriptionsDB.EXPECT().Get("http-GET-%2F").Return(nil, errors.New("KV sub not found"))
	subscriptionsDB.EXPECT().Put("http-GET-%2F", []byte(`{"subscriptionId":"http-GET-%2F","event":"http","functionId":"func","method":"GET","path":"/"}`), nil).Return(nil)
	endpointsDB := mock.NewMockStore(ctrl)
	endpointsDB.EXPECT().Put("GET-%2F", []byte(`{"endpointId":"GET-%2F","functionId":"func","method":"GET","path":"/"}`), nil).Return(nil)
	functionsDB := mock.NewMockStore(ctrl)
	functionsDB.EXPECT().Exists("func").Return(true, nil)
	subs := &Subscriptions{SubscriptionsDB: subscriptionsDB, EndpointsDB: endpointsDB, FunctionsDB: functionsDB, Log: zap.NewNop()}

	_, err := subs.CreateSubscription(&Subscription{ID: "testid", Event: "http", FunctionID: "func", Method: "GET", Path: "/"})

	assert.Nil(t, err)
}

func TestCreateSubscription_HTTPNormalizeMethodPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	subscriptionsDB := mock.NewMockStore(ctrl)
	subscriptionsDB.EXPECT().Get("http-GET-%2F").Return(nil, errors.New("KV sub not found"))
	subscriptionsDB.EXPECT().Put("http-GET-%2F", []byte(`{"subscriptionId":"http-GET-%2F","event":"http","functionId":"func","method":"GET","path":"/"}`), nil).Return(nil)
	endpointsDB := mock.NewMockStore(ctrl)
	endpointsDB.EXPECT().Put("GET-%2F", []byte(`{"endpointId":"GET-%2F","functionId":"func","method":"GET","path":"/"}`), nil).Return(nil)
	functionsDB := mock.NewMockStore(ctrl)
	functionsDB.EXPECT().Exists("func").Return(true, nil)
	subs := &Subscriptions{SubscriptionsDB: subscriptionsDB, EndpointsDB: endpointsDB, FunctionsDB: functionsDB, Log: zap.NewNop()}

	_, err := subs.CreateSubscription(&Subscription{ID: "testid", Event: "http", FunctionID: "func", Method: "get", Path: ""})

	assert.Nil(t, err)
}

func TestCreateSubscription_HTTPValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	subs := &Subscriptions{Log: zap.NewNop()}

	_, err := subs.CreateSubscription(&Subscription{ID: "testid", Event: "http", FunctionID: "func"})

	assert.Equal(t, err, &ErrSubscriptionValidation{original: "Missing required fields (method, path) for HTTP event."})
}

func TestCreateSubscription_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	subscriptionsDB := mock.NewMockStore(ctrl)
	subscriptionsDB.EXPECT().Get("test-func").Return(nil, errors.New("KV sub not found"))
	subscriptionsDB.EXPECT().Put("test-func", []byte(`{"subscriptionId":"test-func","event":"test","functionId":"func"}`), nil).Return(nil)
	topicsDB := mock.NewMockStore(ctrl)
	topicsDB.EXPECT().Get("test").Return(nil, nil)
	functionsDB := mock.NewMockStore(ctrl)
	functionsDB.EXPECT().Exists("func").Return(true, nil)
	subs := &Subscriptions{SubscriptionsDB: subscriptionsDB, TopicsDB: topicsDB, FunctionsDB: functionsDB, Log: zap.NewNop()}

	_, err := subs.CreateSubscription(&Subscription{ID: "testid", Event: "test", FunctionID: "func"})

	assert.Nil(t, err)
}

func TestCreateSubscription_CreateTopicIfNotExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	subscriptionsDB := mock.NewMockStore(ctrl)
	subscriptionsDB.EXPECT().Get("test-func").Return(nil, errors.New("KV sub not found"))
	subscriptionsDB.EXPECT().Put("test-func", []byte(`{"subscriptionId":"test-func","event":"test","functionId":"func"}`), nil).Return(nil)
	topicsDB := mock.NewMockStore(ctrl)
	topicsDB.EXPECT().Get("test").Return(nil, errors.New("KV topic not found"))
	topicsDB.EXPECT().Put("test", []byte(`{"topicId":"test"}`), nil).Return(nil)
	functionsDB := mock.NewMockStore(ctrl)
	functionsDB.EXPECT().Exists("func").Return(true, nil)
	subs := &Subscriptions{SubscriptionsDB: subscriptionsDB, TopicsDB: topicsDB, FunctionsDB: functionsDB, Log: zap.NewNop()}

	_, err := subs.CreateSubscription(&Subscription{ID: "testid", Event: "test", FunctionID: "func"})

	assert.Nil(t, err)
}

func TestCreateSubscription_CreateTopicError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	subscriptionsDB := mock.NewMockStore(ctrl)
	subscriptionsDB.EXPECT().Get("test-func").Return(nil, errors.New("KV sub not found"))
	topicsDB := mock.NewMockStore(ctrl)
	topicsDB.EXPECT().Get("test").Return(nil, errors.New("KV topic not found"))
	topicsDB.EXPECT().Put("test", []byte(`{"topicId":"test"}`), nil).Return(errors.New("KV Put err"))
	functionsDB := mock.NewMockStore(ctrl)
	subs := &Subscriptions{SubscriptionsDB: subscriptionsDB, TopicsDB: topicsDB, FunctionsDB: functionsDB, Log: zap.NewNop()}

	_, err := subs.CreateSubscription(&Subscription{ID: "testid", Event: "test", FunctionID: "func"})

	assert.EqualError(t, err, "KV Put err")
}

func TestCreateSubscription_ValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	subs := &Subscriptions{Log: zap.NewNop()}

	_, err := subs.CreateSubscription(&Subscription{})

	assert.Equal(t, err, &ErrSubscriptionValidation{original: "Key: 'Subscription.Event' Error:Field validation for 'Event' failed on the 'required' tag\nKey: 'Subscription.FunctionID' Error:Field validation for 'FunctionID' failed on the 'required' tag"})
}

func TestCreateSubscription_AlreadyExistsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	subscriptionsDB := mock.NewMockStore(ctrl)
	subscriptionsDB.EXPECT().Get("test-func").Return(&store.KVPair{Value: []byte(`{"subscriptionId":"testid"}`)}, nil)
	subs := &Subscriptions{SubscriptionsDB: subscriptionsDB, Log: zap.NewNop()}

	_, err := subs.CreateSubscription(&Subscription{ID: "testid", Event: "test", FunctionID: "func"})

	assert.Equal(t, err, &ErrSubscriptionAlreadyExists{ID: "test-func"})
}

func TestCreateSubscription_EndpointPutError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	subscriptionsDB := mock.NewMockStore(ctrl)
	subscriptionsDB.EXPECT().Get("http-GET-%2F").Return(nil, errors.New("KV sub not found"))
	endpointsDB := mock.NewMockStore(ctrl)
	endpointsDB.EXPECT().Put("GET-%2F", []byte(`{"endpointId":"GET-%2F","functionId":"func","method":"GET","path":"/"}`), nil).Return(errors.New("KV Put err"))
	subs := &Subscriptions{SubscriptionsDB: subscriptionsDB, EndpointsDB: endpointsDB, Log: zap.NewNop()}

	_, err := subs.CreateSubscription(&Subscription{ID: "testid", Event: "http", FunctionID: "func", Method: "GET", Path: "/"})

	assert.EqualError(t, err, "KV Put err")
}

func TestCreateSubscription_FunctionExistsKVError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	subscriptionsDB := mock.NewMockStore(ctrl)
	subscriptionsDB.EXPECT().Get("http-GET-%2F").Return(nil, errors.New("KV sub not found"))
	endpointsDB := mock.NewMockStore(ctrl)
	endpointsDB.EXPECT().Put("GET-%2F", []byte(`{"endpointId":"GET-%2F","functionId":"func","method":"GET","path":"/"}`), nil).Return(nil)
	functionsDB := mock.NewMockStore(ctrl)
	functionsDB.EXPECT().Exists("func").Return(false, errors.New("KV Exists err"))
	subs := &Subscriptions{SubscriptionsDB: subscriptionsDB, EndpointsDB: endpointsDB, FunctionsDB: functionsDB, Log: zap.NewNop()}

	_, err := subs.CreateSubscription(&Subscription{ID: "testid", Event: "http", FunctionID: "func", Method: "GET", Path: "/"})

	assert.EqualError(t, err, "KV Exists err")
}

func TestCreateSubscription_FunctionExistsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	subscriptionsDB := mock.NewMockStore(ctrl)
	subscriptionsDB.EXPECT().Get("http-GET-%2F").Return(nil, errors.New("KV sub not found"))
	endpointsDB := mock.NewMockStore(ctrl)
	endpointsDB.EXPECT().Put("GET-%2F", []byte(`{"endpointId":"GET-%2F","functionId":"func","method":"GET","path":"/"}`), nil).Return(nil)
	functionsDB := mock.NewMockStore(ctrl)
	functionsDB.EXPECT().Exists("func").Return(false, nil)
	subs := &Subscriptions{SubscriptionsDB: subscriptionsDB, EndpointsDB: endpointsDB, FunctionsDB: functionsDB, Log: zap.NewNop()}

	_, err := subs.CreateSubscription(&Subscription{ID: "testid", Event: "http", FunctionID: "func", Method: "GET", Path: "/"})

	assert.Equal(t, err, &ErrFunctionNotFound{functionID: "func"})
}

func TestCreateSubscription_PutError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	subscriptionsDB := mock.NewMockStore(ctrl)
	subscriptionsDB.EXPECT().Get("http-GET-%2F").Return(nil, errors.New("KV sub not found"))
	subscriptionsDB.EXPECT().Put("http-GET-%2F", []byte(`{"subscriptionId":"http-GET-%2F","event":"http","functionId":"func","method":"GET","path":"/"}`), nil).Return(errors.New("KV Put err"))
	endpointsDB := mock.NewMockStore(ctrl)
	endpointsDB.EXPECT().Put("GET-%2F", []byte(`{"endpointId":"GET-%2F","functionId":"func","method":"GET","path":"/"}`), nil).Return(nil)
	functionsDB := mock.NewMockStore(ctrl)
	functionsDB.EXPECT().Exists("func").Return(true, nil)
	subs := &Subscriptions{SubscriptionsDB: subscriptionsDB, EndpointsDB: endpointsDB, FunctionsDB: functionsDB, Log: zap.NewNop()}

	_, err := subs.CreateSubscription(&Subscription{ID: "testid", Event: "http", FunctionID: "func", Method: "GET", Path: "/"})

	assert.EqualError(t, err, "KV Put err")
}

func TestDeleteSubscription_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	kv := &store.KVPair{Value: []byte(`{"subscriptionId":"testid","event":"test","functionId":"f1"}`)}
	subscriptionsDB := mock.NewMockStore(ctrl)
	subscriptionsDB.EXPECT().Get("testid").Return(kv, nil)
	subscriptionsDB.EXPECT().List("").Return([]*store.KVPair{}, nil)
	subscriptionsDB.EXPECT().Delete("testid").Return(nil)
	topicsDB := mock.NewMockStore(ctrl)
	topicsDB.EXPECT().Delete("test").Return(nil)
	subs := &Subscriptions{SubscriptionsDB: subscriptionsDB, TopicsDB: topicsDB, Log: zap.NewNop()}

	err := subs.DeleteSubscription(SubscriptionID("testid"))

	assert.Nil(t, err)
}

func TestDeleteSubscription_GetError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	subscriptionsDB := mock.NewMockStore(ctrl)
	subscriptionsDB.EXPECT().Get("testid").Return(nil, errors.New("KV Get err"))
	subs := &Subscriptions{SubscriptionsDB: subscriptionsDB, Log: zap.NewNop()}

	err := subs.DeleteSubscription(SubscriptionID("testid"))

	assert.Equal(t, err, &ErrSubscriptionNotFound{"testid"})
}

func TestDeleteSubscription_DeleteError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	kv := &store.KVPair{Value: []byte(`{"subscriptionId":"testid","event":"test","functionId":"f1"}`)}
	subscriptionsDB := mock.NewMockStore(ctrl)
	subscriptionsDB.EXPECT().Get("testid").Return(kv, nil)
	subscriptionsDB.EXPECT().Delete("testid").Return(errors.New("KV Delete err"))
	subs := &Subscriptions{SubscriptionsDB: subscriptionsDB, Log: zap.NewNop()}

	err := subs.DeleteSubscription(SubscriptionID("testid"))

	assert.Equal(t, err, &ErrSubscriptionNotFound{"testid"})
}

func TestDeleteSubscription_DeleteEndpointOK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	kv := &store.KVPair{Value: []byte(`{"subscriptionId":"testid","event":"http","functionId":"f1","method":"GET","path":"/"}`)}
	subscriptionsDB := mock.NewMockStore(ctrl)
	subscriptionsDB.EXPECT().Get("testid").Return(kv, nil)
	subscriptionsDB.EXPECT().Delete("testid").Return(nil)
	topicsDB := mock.NewMockStore(ctrl)
	endpointsDB := mock.NewMockStore(ctrl)
	endpointsDB.EXPECT().Delete("GET-%2F").Return(nil)
	subs := &Subscriptions{SubscriptionsDB: subscriptionsDB, TopicsDB: topicsDB, EndpointsDB: endpointsDB, Log: zap.NewNop()}

	err := subs.DeleteSubscription(SubscriptionID("testid"))

	assert.Nil(t, err)
}

func TestDeleteSubscription_DeleteEndpointError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	kv := &store.KVPair{Value: []byte(`{"subscriptionId":"testid","event":"http","functionId":"f1","method":"GET","path":"/"}`)}
	subscriptionsDB := mock.NewMockStore(ctrl)
	subscriptionsDB.EXPECT().Get("testid").Return(kv, nil)
	subscriptionsDB.EXPECT().Delete("testid").Return(nil)
	endpointsDB := mock.NewMockStore(ctrl)
	endpointsDB.EXPECT().Delete("GET-%2F").Return(errors.New("KV Delete err"))
	subs := &Subscriptions{SubscriptionsDB: subscriptionsDB, EndpointsDB: endpointsDB, Log: zap.NewNop()}

	err := subs.DeleteSubscription(SubscriptionID("testid"))

	assert.EqualError(t, err, "KV Delete err")
}

func TestDeleteSubscription_DeleteTopicError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	kv := &store.KVPair{Value: []byte(`{"subscriptionId":"testid","event":"test","functionId":"f1"}`)}
	subscriptionsDB := mock.NewMockStore(ctrl)
	subscriptionsDB.EXPECT().Get("testid").Return(kv, nil)
	subscriptionsDB.EXPECT().List("").Return([]*store.KVPair{}, nil)
	subscriptionsDB.EXPECT().Delete("testid").Return(nil)
	topicsDB := mock.NewMockStore(ctrl)
	topicsDB.EXPECT().Delete("test").Return(errors.New("KV Delete err"))
	subs := &Subscriptions{SubscriptionsDB: subscriptionsDB, TopicsDB: topicsDB, Log: zap.NewNop()}

	err := subs.DeleteSubscription(SubscriptionID("testid"))

	assert.EqualError(t, err, "KV Delete err")
}

func TestDeleteSubscription_DeleteTopicListError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	kv := &store.KVPair{Value: []byte(`{"subscriptionId":"testid","event":"test","functionId":"f1"}`)}
	subscriptionsDB := mock.NewMockStore(ctrl)
	subscriptionsDB.EXPECT().Get("testid").Return(kv, nil)
	subscriptionsDB.EXPECT().List("").Return(nil, errors.New("KV List error"))
	subscriptionsDB.EXPECT().Delete("testid").Return(nil)
	subs := &Subscriptions{SubscriptionsDB: subscriptionsDB, Log: zap.NewNop()}

	err := subs.DeleteSubscription(SubscriptionID("testid"))

	assert.EqualError(t, err, "KV List error")
}

func TestDeleteSubscription_NotDeleteTopicWithSubOK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	kv := &store.KVPair{Value: []byte(`{"subscriptionId":"s1","event":"test","functionId":"f1"}`)}
	kvs := []*store.KVPair{{Value: []byte(`{"subscriptionId":"s2","event":"test","functionId":"f2"}`)}}
	subscriptionsDB := mock.NewMockStore(ctrl)
	subscriptionsDB.EXPECT().Get("s1").Return(kv, nil)
	subscriptionsDB.EXPECT().List("").Return(kvs, nil)
	subscriptionsDB.EXPECT().Delete("s1").Return(nil)
	topicsDB := mock.NewMockStore(ctrl)
	subs := &Subscriptions{SubscriptionsDB: subscriptionsDB, TopicsDB: topicsDB, Log: zap.NewNop()}

	err := subs.DeleteSubscription(SubscriptionID("s1"))

	assert.Nil(t, err)
}

func TestGetAllSubscriptions_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	kvs := []*store.KVPair{
		{Value: []byte(`{"subscriptionId":"s1","event":"test","functionId":"f1"}`)},
		{Value: []byte(`{"subscriptionId":"s2","event":"test","functionId":"f2"}`)},
	}
	subscriptionsDB := mock.NewMockStore(ctrl)
	subscriptionsDB.EXPECT().List("").Return(kvs, nil)
	subs := &Subscriptions{SubscriptionsDB: subscriptionsDB, Log: zap.NewNop()}

	list, _ := subs.GetAllSubscriptions()

	assert.Equal(t, []*Subscription{
		{ID: SubscriptionID("s1"), Event: "test", FunctionID: functions.FunctionID("f1")},
		{ID: SubscriptionID("s2"), Event: "test", FunctionID: functions.FunctionID("f2")},
	}, list)
}

func TestGetAllSubscriptions_ListError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	subscriptionsDB := mock.NewMockStore(ctrl)
	subscriptionsDB.EXPECT().List("").Return(nil, errors.New("KV error"))
	subs := &Subscriptions{SubscriptionsDB: subscriptionsDB, Log: zap.NewNop()}

	_, err := subs.GetAllSubscriptions()
	assert.EqualError(t, err, "KV error")
}
