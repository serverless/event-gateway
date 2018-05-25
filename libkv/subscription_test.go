package libkv

import (
	"errors"
	"testing"

	"go.uber.org/zap"

	"github.com/golang/mock/gomock"
	"github.com/serverless/event-gateway/event"
	"github.com/serverless/event-gateway/function"
	"github.com/serverless/event-gateway/mock"
	"github.com/serverless/event-gateway/subscription"
	"github.com/serverless/libkv/store"
	"github.com/stretchr/testify/assert"
)

func TestCreateSubscription(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	asyncKey := "default/YXN5bmMsdXNlci5jcmVhdGVkLGZ1bmMsJTJGLEdFVA"
	asyncValue := []byte(
		`{"space":"default","subscriptionId":"YXN5bmMsdXNlci5jcmVhdGVkLGZ1bmMsJTJGLEdFVA",` +
			`"type":"async","eventType":"user.created","functionId":"func","path":"/","method":"GET"}`)
	asyncSub := &subscription.Subscription{
		Type: subscription.TypeAsync, EventType: "user.created", FunctionID: "func", Path: "/", Method: "GET"}
	syncKey := "default/c3luYyxodHRwLnJlcXVlc3QsZnVuYywlMkYsUE9TVA"
	syncValue := []byte(
		`{"space":"default","subscriptionId":"c3luYyxodHRwLnJlcXVlc3QsZnVuYywlMkYsUE9TVA",` +
			`"type":"sync","eventType":"http.request","functionId":"func","path":"/","method":"POST"}`)
	syncSub := &subscription.Subscription{
		Type: subscription.TypeSync, EventType: "http.request", FunctionID: "func", Path: "/", Method: "POST"}
	funcValue := []byte(`{"functionId":"func","type":"http","provider":{"url": "http://test.com"}}}`)

	t.Run("async subscription created", func(t *testing.T) {
		subscriptionsDB := mock.NewMockStore(ctrl)
		subscriptionsDB.EXPECT().Get(asyncKey, &store.ReadOptions{Consistent: true}).Return(nil, errors.New("KV sub not found"))
		subscriptionsDB.EXPECT().Put(asyncKey, asyncValue, nil).Return(nil)
		functionsDB := mock.NewMockStore(ctrl)
		functionsDB.EXPECT().Get("default/func", &store.ReadOptions{Consistent: true}).Return(&store.KVPair{Value: funcValue}, nil)
		subs := &Service{SubscriptionStore: subscriptionsDB, FunctionStore: functionsDB, Log: zap.NewNop()}

		_, err := subs.CreateSubscription(asyncSub)

		assert.Nil(t, err)
	})

	t.Run("sync subscription created", func(t *testing.T) {
		subscriptionsDB := mock.NewMockStore(ctrl)
		subscriptionsDB.EXPECT().Get(syncKey, &store.ReadOptions{Consistent: true}).Return(nil, errors.New("KV sub not found"))
		subscriptionsDB.EXPECT().Put(syncKey, syncValue, nil).Return(nil)
		subscriptionsDB.EXPECT().List("default/", &store.ReadOptions{Consistent: true}).Return([]*store.KVPair{}, nil)
		functionsDB := mock.NewMockStore(ctrl)
		functionsDB.EXPECT().Get("default/func", &store.ReadOptions{Consistent: true}).Return(&store.KVPair{Value: funcValue}, nil)
		subs := &Service{SubscriptionStore: subscriptionsDB, FunctionStore: functionsDB, Log: zap.NewNop()}

		_, err := subs.CreateSubscription(syncSub)

		assert.Nil(t, err)
	})

	t.Run("validation error", func(t *testing.T) {
		subs := &Service{Log: zap.NewNop()}

		_, err := subs.CreateSubscription(&subscription.Subscription{})

		assert.Equal(t, err, &subscription.ErrSubscriptionValidation{
			Message: "Key: 'Subscription.Type' Error:Field validation for 'Type' failed on the 'required' tag" +
				"\nKey: 'Subscription.EventType' Error:Field validation for 'EventType' failed on the 'required' tag" +
				"\nKey: 'Subscription.FunctionID' Error:Field validation for 'FunctionID' failed on the 'required' tag"})
	})

	t.Run("validation error: CORS settings for async subscription", func(t *testing.T) {
		subs := &Service{Log: zap.NewNop()}

		_, err := subs.CreateSubscription(
			&subscription.Subscription{
				Type:       subscription.TypeAsync,
				EventType:  "user.created",
				FunctionID: "func",
				Path:       "/",
				Method:     "GET",
				CORS: &subscription.CORS{
					Methods: []string{"GET"},
				},
			},
		)

		assert.Equal(t, err, &subscription.ErrSubscriptionValidation{Message: "CORS can be configured only for sync subscriptions."})
	})

	t.Run("subscription already exists", func(t *testing.T) {
		subscriptionsDB := mock.NewMockStore(ctrl)
		subscriptionsDB.EXPECT().Get(gomock.Any(), gomock.Any()).Return(&store.KVPair{Value: []byte(`{"subscriptionId":""}`)}, nil)
		subs := &Service{SubscriptionStore: subscriptionsDB, Log: zap.NewNop()}

		_, err := subs.CreateSubscription(asyncSub)

		assert.Equal(t, err, &subscription.ErrSubscriptionAlreadyExists{ID: "YXN5bmMsdXNlci5jcmVhdGVkLGZ1bmMsJTJGLEdFVA"})
	})

	t.Run("subscription path conflict", func(t *testing.T) {
		subscriptionsDB := mock.NewMockStore(ctrl)
		subscriptionsDB.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, errors.New("KV sub not found"))
		kv := &store.KVPair{
			Value: []byte(`{"subscriptionId":"test","type":"sync","functionId":"func","method":"GET","path":"/:name"}`)}
		subscriptionsDB.EXPECT().List(gomock.Any(), gomock.Any()).Return([]*store.KVPair{kv}, nil)
		subs := &Service{SubscriptionStore: subscriptionsDB, Log: zap.NewNop()}

		_, err := subs.CreateSubscription(
			&subscription.Subscription{
				Space:      "default",
				Type:       subscription.TypeSync,
				EventType:  "http.request",
				FunctionID: "func",
				Method:     "GET",
				Path:       "/:id"})

		assert.Equal(t, err, &subscription.ErrPathConfict{
			Message: `parameter with different name ("name") already defined: for route: /:id`})
	})

	t.Run("function KV Get error", func(t *testing.T) {
		subscriptionsDB := mock.NewMockStore(ctrl)
		subscriptionsDB.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, errors.New("KV sub not found"))
		functionsDB := mock.NewMockStore(ctrl)
		functionsDB.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, errors.New("Key not found in store"))
		subs := &Service{SubscriptionStore: subscriptionsDB, FunctionStore: functionsDB, Log: zap.NewNop()}

		_, err := subs.CreateSubscription(asyncSub)

		assert.Equal(t, err, &function.ErrFunctionNotFound{ID: "func"})
	})

	t.Run("KV Put error", func(t *testing.T) {
		subscriptionsDB := mock.NewMockStore(ctrl)
		subscriptionsDB.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, errors.New("KV sub not found"))
		subscriptionsDB.EXPECT().List(gomock.Any(), gomock.Any()).Return([]*store.KVPair{}, nil)
		subscriptionsDB.EXPECT().Put(gomock.Any(), gomock.Any(), nil).Return(errors.New("KV Put err"))
		functionsDB := mock.NewMockStore(ctrl)
		functionsDB.EXPECT().Get("default/func", gomock.Any()).Return(
			&store.KVPair{Value: []byte(`{"functionId":"func","type":"http","provider":{"url": "http://test.com"}}`)}, nil)
		subs := &Service{SubscriptionStore: subscriptionsDB, FunctionStore: functionsDB, Log: zap.NewNop()}

		_, err := subs.CreateSubscription(syncSub)

		assert.EqualError(t, err, "KV Put err")
	})
}

func TestUpdateSubscription(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	syncID := subscription.ID("c3luYyxodHRwLnJlcXVlc3QsZnVuYywlMkYsUE9TVA")
	syncKey := "default/c3luYyxodHRwLnJlcXVlc3QsZnVuYywlMkYsUE9TVA"
	syncValue := []byte(
		`{"space":"default","subscriptionId":"c3luYyxodHRwLnJlcXVlc3QsZnVuYywlMkYsUE9TVA",` +
			`"type":"sync","eventType":"http.request","functionId":"func","path":"/","method":"POST"}`)
	funcValue := []byte(`{"functionId":"func","type":"http","provider":{"url": "http://test.com"}}}`)

	t.Run("subscription updated", func(t *testing.T) {
		subscriptionsDB := mock.NewMockStore(ctrl)
		subscriptionsDB.EXPECT().Get(syncKey, &store.ReadOptions{Consistent: true}).Return(&store.KVPair{Value: syncValue}, nil)
		subscriptionsDB.EXPECT().Put(
			syncKey,
			[]byte(
				`{"space":"default","subscriptionId":"c3luYyxodHRwLnJlcXVlc3QsZnVuYywlMkYsUE9TVA","type":"sync",`+
					`"eventType":"http.request","functionId":"func","path":"/","method":"POST",`+
					`"cors":{"origins":["*"],"methods":["HEAD","GET","POST"],`+
					`"headers":["Origin","Accept","Content-Type"],"allowCredentials":false}}`),
			nil).Return(nil)
		functionsDB := mock.NewMockStore(ctrl)
		functionsDB.EXPECT().Get("default/func", &store.ReadOptions{Consistent: true}).Return(&store.KVPair{Value: funcValue}, nil)
		subs := &Service{SubscriptionStore: subscriptionsDB, FunctionStore: functionsDB, Log: zap.NewNop()}

		_, err := subs.UpdateSubscription(
			syncID,
			&subscription.Subscription{
				ID:         syncID,
				Type:       subscription.TypeSync,
				EventType:  "http.request",
				FunctionID: "func",
				Path:       "/",
				Method:     "POST",
				CORS:       &subscription.CORS{Origins: []string{"*"}}})

		assert.Nil(t, err)
	})

	t.Run("validation error", func(t *testing.T) {
		subs := &Service{Log: zap.NewNop()}

		_, err := subs.UpdateSubscription(syncID, &subscription.Subscription{Type: subscription.TypeSync})

		assert.Equal(t, err, &subscription.ErrSubscriptionValidation{
			Message: "Key: 'Subscription.EventType' Error:Field validation for 'EventType' failed on the 'required' tag" +
				"\nKey: 'Subscription.FunctionID' Error:Field validation for 'FunctionID' failed on the 'required' tag"})
	})

	t.Run("invalid subscription update", func(t *testing.T) {
		subscriptionsDB := mock.NewMockStore(ctrl)
		subscriptionsDB.EXPECT().Get(gomock.Any(), gomock.Any()).Return(&store.KVPair{Value: syncValue}, nil)
		subs := &Service{SubscriptionStore: subscriptionsDB, Log: zap.NewNop()}
		_, err := subs.UpdateSubscription(
			syncID,
			&subscription.Subscription{
				ID:         syncID,
				Type:       subscription.TypeSync,
				EventType:  "http.request",
				FunctionID: "func2",
				Path:       "/",
				Method:     "POST"})

		assert.Equal(t, err, &subscription.ErrInvalidSubscriptionUpdate{Field: "FunctionID"})
	})

	t.Run("subscription not found", func(t *testing.T) {
		subscriptionsDB := mock.NewMockStore(ctrl)
		subscriptionsDB.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, errors.New("Key not found in store"))
		functionsDB := mock.NewMockStore(ctrl)
		subs := &Service{SubscriptionStore: subscriptionsDB, FunctionStore: functionsDB, Log: zap.NewNop()}
		_, err := subs.UpdateSubscription(
			syncID,
			&subscription.Subscription{
				ID:         syncID,
				Type:       subscription.TypeSync,
				EventType:  "http.request",
				FunctionID: "func",
				Path:       "/",
				Method:     "POST"})

		assert.Equal(t, err, &subscription.ErrSubscriptionNotFound{ID: syncID})
	})

	t.Run("function not found", func(t *testing.T) {
		subscriptionsDB := mock.NewMockStore(ctrl)
		subscriptionsDB.EXPECT().Get(gomock.Any(), gomock.Any()).Return(&store.KVPair{Value: syncValue}, nil)
		functionsDB := mock.NewMockStore(ctrl)
		functionsDB.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, errors.New("Key not found in store"))
		subs := &Service{SubscriptionStore: subscriptionsDB, FunctionStore: functionsDB, Log: zap.NewNop()}

		_, err := subs.UpdateSubscription(
			syncID,
			&subscription.Subscription{
				ID:         syncID,
				Type:       subscription.TypeSync,
				EventType:  "http.request",
				FunctionID: "func",
				Path:       "/",
				Method:     "POST"})

		assert.Equal(t, err, &function.ErrFunctionNotFound{ID: "func"})
	})

	t.Run("KV Put error", func(t *testing.T) {
		subscriptionsDB := mock.NewMockStore(ctrl)
		subscriptionsDB.EXPECT().Get(gomock.Any(), gomock.Any()).Return(&store.KVPair{Value: syncValue}, nil)
		subscriptionsDB.EXPECT().Put(gomock.Any(), gomock.Any(), nil).Return(errors.New("KV Put err"))
		functionsDB := mock.NewMockStore(ctrl)
		functionsDB.EXPECT().Get(gomock.Any(), gomock.Any()).Return(&store.KVPair{Value: funcValue}, nil)
		subs := &Service{SubscriptionStore: subscriptionsDB, FunctionStore: functionsDB, Log: zap.NewNop()}

		_, err := subs.UpdateSubscription(
			syncID,
			&subscription.Subscription{
				ID:         syncID,
				Type:       subscription.TypeSync,
				EventType:  "http.request",
				FunctionID: "func",
				Path:       "/",
				Method:     "POST"})

		assert.EqualError(t, err, "KV Put err")
	})
}

func TestDeleteSubscription(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("subscription deleted", func(t *testing.T) {
		kv := &store.KVPair{Value: []byte(`{"subscriptionId":"testid","event":"test","functionId":"f1"}`)}
		subscriptionsDB := mock.NewMockStore(ctrl)
		subscriptionsDB.EXPECT().Get("default/testid", gomock.Any()).Return(kv, nil)
		subscriptionsDB.EXPECT().Delete("default/testid").Return(nil)
		subs := &Service{SubscriptionStore: subscriptionsDB, Log: zap.NewNop()}

		err := subs.DeleteSubscription("default", subscription.ID("testid"))

		assert.Nil(t, err)
	})

	t.Run("subscription Get KV error", func(t *testing.T) {
		subscriptionsDB := mock.NewMockStore(ctrl)
		subscriptionsDB.EXPECT().Get("default/testid", gomock.Any()).Return(nil, errors.New("Key not found in store"))
		subs := &Service{SubscriptionStore: subscriptionsDB, Log: zap.NewNop()}

		err := subs.DeleteSubscription("default", subscription.ID("testid"))

		assert.Equal(t, err, &subscription.ErrSubscriptionNotFound{ID: "testid"})
	})

	t.Run("KV Delete error", func(t *testing.T) {
		kv := &store.KVPair{Value: []byte(`{"subscriptionId":"testid","event":"test","functionId":"f1"}`)}
		subscriptionsDB := mock.NewMockStore(ctrl)
		subscriptionsDB.EXPECT().Get("default/testid", gomock.Any()).Return(kv, nil)
		subscriptionsDB.EXPECT().Delete("default/testid").Return(errors.New("KV Delete err"))
		subs := &Service{SubscriptionStore: subscriptionsDB, Log: zap.NewNop()}

		err := subs.DeleteSubscription("default", subscription.ID("testid"))

		assert.Equal(t, err, &subscription.ErrSubscriptionNotFound{ID: "testid"})
	})
}

func TestGetSubscriptions_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("list returned", func(t *testing.T) {
		kvs := []*store.KVPair{
			{Value: []byte(`{"subscriptionId":"s1","space":"default","type":"async","eventType":"test","functionId":"f1"}`)},
			{Value: []byte(`{"subscriptionId":"s2","space":"default","type":"async","eventType":"test","functionId":"f2"}`)},
		}
		subscriptionsDB := mock.NewMockStore(ctrl)
		subscriptionsDB.EXPECT().List("default/", &store.ReadOptions{Consistent: true}).Return(kvs, nil)
		subs := &Service{SubscriptionStore: subscriptionsDB, Log: zap.NewNop()}

		list, _ := subs.GetSubscriptions("default")

		assert.Equal(t, subscription.Subscriptions{
			{
				ID:         subscription.ID("s1"),
				Space:      "default",
				Type:       subscription.TypeAsync,
				EventType:  "test",
				FunctionID: function.ID("f1")},
			{
				ID:         subscription.ID("s2"),
				Space:      "default",
				Type:       subscription.TypeAsync,
				EventType:  "test",
				FunctionID: function.ID("f2")},
		}, list)
	})

	t.Run("KV List error", func(t *testing.T) {
		subscriptionsDB := mock.NewMockStore(ctrl)
		subscriptionsDB.EXPECT().List("default/", gomock.Any()).Return(nil, errors.New("KV error"))
		subs := &Service{SubscriptionStore: subscriptionsDB, Log: zap.NewNop()}

		_, err := subs.GetSubscriptions("default")

		assert.EqualError(t, err, "KV error")
	})
}

func TestGetSubscription(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("subscription returned", func(t *testing.T) {
		kv := &store.KVPair{Value: []byte(`{"subscriptionId":"testid","type":"async","eventType":"test","functionId":"f1"}`)}
		subscriptionsDB := mock.NewMockStore(ctrl)
		subscriptionsDB.EXPECT().Get("default/testid", gomock.Any()).Return(kv, nil)
		subs := &Service{SubscriptionStore: subscriptionsDB, Log: zap.NewNop()}

		sub, _ := subs.GetSubscription("default", subscription.ID("testid"))

		assert.Equal(t, subscription.ID("testid"), sub.ID)
		assert.Equal(t, subscription.TypeAsync, sub.Type)
		assert.Equal(t, event.TypeName("test"), sub.EventType)
		assert.Equal(t, function.ID("f1"), sub.FunctionID)
	})

	t.Run("not found", func(t *testing.T) {
		subscriptionsDB := mock.NewMockStore(ctrl)
		subscriptionsDB.EXPECT().Get("default/testid", gomock.Any()).Return(nil, errors.New("Key not found in store"))
		subs := &Service{SubscriptionStore: subscriptionsDB, Log: zap.NewNop()}

		_, err := subs.GetSubscription("default", subscription.ID("testid"))

		assert.Equal(t, err, &subscription.ErrSubscriptionNotFound{ID: "testid"})
	})

	t.Run("KV Get error", func(t *testing.T) {
		subscriptionsDB := mock.NewMockStore(ctrl)
		subscriptionsDB.EXPECT().Get("default/testid", gomock.Any()).Return(nil, errors.New("KV error"))
		subs := &Service{SubscriptionStore: subscriptionsDB, Log: zap.NewNop()}

		_, err := subs.GetSubscription("default", subscription.ID("testid"))

		assert.EqualError(t, err, "KV error")
	})
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
