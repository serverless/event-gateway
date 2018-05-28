package libkv

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/serverless/event-gateway/event"
	"github.com/serverless/event-gateway/mock"
	"github.com/serverless/libkv/store"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	_ "github.com/serverless/event-gateway/providers/http"
)

func TestCreateEventType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testEventType := &event.Type{Name: "test.event"}

	t.Run("event type created", func(t *testing.T) {
		db := mock.NewMockStore(ctrl)
		db.EXPECT().
			Get("default/test.event", &store.ReadOptions{Consistent: true}).
			Return(nil, errors.New("KV type not found"))
		payload := []byte(`{"space":"default","name":"test.event"}`)
		db.EXPECT().Put("default/test.event", payload, nil).Return(nil)
		service := &Service{EventTypeStore: db, Log: zap.NewNop()}

		_, err := service.CreateEventType(testEventType)

		assert.Nil(t, err)
	})

	t.Run("event type already exists", func(t *testing.T) {
		db := mock.NewMockStore(ctrl)
		db.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, nil)
		service := &Service{EventTypeStore: db, Log: zap.NewNop()}

		_, err := service.CreateEventType(testEventType)

		assert.Equal(t, &event.ErrEventTypeAlreadyExists{Name: "test.event"}, err)
	})

	t.Run("validation error", func(t *testing.T) {
		service := &Service{Log: zap.NewNop()}

		_, err := service.CreateEventType(&event.Type{})

		assert.Equal(t, &event.ErrEventTypeValidation{
			Message: "Key: 'Type.Name' Error:Field validation for 'Name' failed on the 'required' tag",
		}, err)
	})

	t.Run("KV Put error", func(t *testing.T) {
		db := mock.NewMockStore(ctrl)
		db.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, errors.New("KV type not found"))
		db.EXPECT().Put(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("KV put error"))
		service := &Service{EventTypeStore: db, Log: zap.NewNop()}

		_, err := service.CreateEventType(testEventType)

		assert.EqualError(t, err, "KV put error")
	})
}

func TestGetEventType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testEventType := &event.Type{Space: "default", Name: "test.event"}
	testPayload := []byte(`{"space":"default","name":"test.event"}}`)

	t.Run("event type returned", func(t *testing.T) {
		db := mock.NewMockStore(ctrl)
		db.EXPECT().
			Get("default/test.event", &store.ReadOptions{Consistent: true}).
			Return(&store.KVPair{Value: testPayload}, nil)
		service := &Service{EventTypeStore: db, Log: zap.NewNop()}

		eventType, _ := service.GetEventType("default", event.TypeName("test.event"))

		assert.Equal(t, testEventType, eventType)
	})

	t.Run("event type not found", func(t *testing.T) {
		db := mock.NewMockStore(ctrl)
		db.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, errors.New("Key not found in store"))
		service := &Service{EventTypeStore: db, Log: zap.NewNop()}

		_, err := service.GetEventType("default", event.TypeName("test.event"))

		assert.Equal(t, &event.ErrEventTypeNotFound{Name: "test.event"}, err)
	})

	t.Run("KV Get error", func(t *testing.T) {
		db := mock.NewMockStore(ctrl)
		db.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, errors.New("KV get err"))
		service := &Service{EventTypeStore: db, Log: zap.NewNop()}

		_, err := service.GetEventType("default", event.TypeName("test.event"))

		assert.EqualError(t, err, "KV get err")
	})
}

func TestGetEventTypes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testEventType := &event.Type{Space: "default", Name: "test.event"}
	testPayload := []byte(`{"space":"default","name":"test.event"}}`)

	t.Run("event types returned", func(t *testing.T) {
		kvs := []*store.KVPair{&store.KVPair{Value: testPayload}}
		db := mock.NewMockStore(ctrl)
		db.EXPECT().List("default/", &store.ReadOptions{Consistent: true}).Return(kvs, nil)
		service := &Service{EventTypeStore: db, Log: zap.NewNop()}

		list, err := service.GetEventTypes("default")

		assert.Nil(t, err)
		assert.Equal(t, event.Types{testEventType}, list)
	})

	t.Run("KV List error", func(t *testing.T) {
		db := mock.NewMockStore(ctrl)
		db.EXPECT().List(gomock.Any(), gomock.Any()).Return([]*store.KVPair{}, errors.New("KV list err"))
		service := &Service{EventTypeStore: db, Log: zap.NewNop()}

		_, err := service.GetEventTypes("default")

		assert.EqualError(t, err, "KV list err")
	})

	t.Run("KV List directory not found", func(t *testing.T) {
		db := mock.NewMockStore(ctrl)
		db.EXPECT().List(gomock.Any(), gomock.Any()).Return([]*store.KVPair{}, errors.New("Key not found in store"))
		service := &Service{EventTypeStore: db, Log: zap.NewNop()}

		list, _ := service.GetEventTypes("default")

		assert.Equal(t, event.Types{}, list)
	})
}

func TestDeleteEventType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("event type deleted", func(t *testing.T) {
		subscriptionsDB := mock.NewMockStore(ctrl)
		subscriptionsDB.EXPECT().List("default/", &store.ReadOptions{Consistent: true}).Return([]*store.KVPair{}, nil)
		eventTypesDB := mock.NewMockStore(ctrl)
		eventTypesDB.EXPECT().Delete("default/test.event").Return(nil)
		service := &Service{EventTypeStore: eventTypesDB, SubscriptionStore: subscriptionsDB, Log: zap.NewNop()}

		err := service.DeleteEventType("default", event.TypeName("test.event"))

		assert.Nil(t, err)
	})

	t.Run("event type not found", func(t *testing.T) {
		subscriptionsDB := mock.NewMockStore(ctrl)
		subscriptionsDB.EXPECT().List(gomock.Any(), gomock.Any()).Return([]*store.KVPair{}, nil)
		eventTypeDB := mock.NewMockStore(ctrl)
		eventTypeDB.EXPECT().Delete(gomock.Any()).Return(errors.New("KV func not found"))
		service := &Service{EventTypeStore: eventTypeDB, SubscriptionStore: subscriptionsDB, Log: zap.NewNop()}

		err := service.DeleteEventType("default", event.TypeName("test.event"))

		assert.Equal(t, &event.ErrEventTypeNotFound{Name: "test.event"}, err)
	})

	t.Run("subscriptions exist", func(t *testing.T) {
		kvs := []*store.KVPair{
			{Value: []byte(`{"subscriptionId":"s1","space":"default","eventType":"test.event","functionId":"testid"}`)}}
		subscriptionsDB := mock.NewMockStore(ctrl)
		subscriptionsDB.EXPECT().List(gomock.Any(), gomock.Any()).Return(kvs, nil)
		eventTypesDB := mock.NewMockStore(ctrl)
		service := &Service{EventTypeStore: eventTypesDB, SubscriptionStore: subscriptionsDB, Log: zap.NewNop()}

		err := service.DeleteEventType("default", event.TypeName("test.event"))

		assert.Equal(t, &event.ErrEventTypeHasSubscriptionsError{}, err)
	})
}
