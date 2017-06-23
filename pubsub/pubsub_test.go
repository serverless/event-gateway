package pubsub_test

import (
	"errors"
	"testing"

	"github.com/docker/libkv/store"
	"github.com/golang/mock/gomock"
	"github.com/serverless/event-gateway/endpoints/mock"
	"github.com/serverless/event-gateway/pubsub"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestCreate_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	topicsDB := mock.NewMockStore(ctrl)
	topicsDB.EXPECT().Get("test").Return(nil, errors.New("not found"))
	topicsDB.EXPECT().Put("test", []byte(`{"topicId":"test"}`), nil).Return(nil)
	ps := &pubsub.PubSub{TopicsDB: topicsDB, Logger: zap.NewNop()}

	tc, _ := ps.CreateTopic(&pubsub.Topic{ID: "test"})

	assert.Equal(t, &pubsub.Topic{ID: "test"}, tc)
}

func TestCreate_TopicAlreadyExistsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	topicsDB := mock.NewMockStore(ctrl)
	topicsDB.EXPECT().Get("test").Return(nil, nil)
	ps := &pubsub.PubSub{TopicsDB: topicsDB, Logger: zap.NewNop()}

	_, err := ps.CreateTopic(&pubsub.Topic{ID: "test"})

	assert.EqualError(t, err, `Topic "test" already exits.`)
}

func TestCreate_ValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	topicsDB := mock.NewMockStore(ctrl)
	ps := &pubsub.PubSub{TopicsDB: topicsDB, Logger: zap.NewNop()}

	_, err := ps.CreateTopic(&pubsub.Topic{ID: ""})

	assert.EqualError(t, err, `Topic doesn't validate. Validation error: "Key: 'Topic.ID' Error:Field validation for 'ID' failed on the 'required' tag"`)
}

func TestCreate_DBPutError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	topicsDB := mock.NewMockStore(ctrl)
	topicsDB.EXPECT().Get("test").Return(nil, errors.New("not found"))
	topicsDB.EXPECT().Put(gomock.Any(), gomock.Any(), nil).Return(errors.New("db put failed"))
	ps := &pubsub.PubSub{TopicsDB: topicsDB, Logger: zap.NewNop()}

	_, err := ps.CreateTopic(&pubsub.Topic{ID: "test"})

	assert.EqualError(t, err, "db put failed")
}

func TestDelete_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	topicsDB := mock.NewMockStore(ctrl)
	topicsDB.EXPECT().Delete("testid").Return(nil)
	ps := &pubsub.PubSub{TopicsDB: topicsDB, Logger: zap.NewNop()}

	assert.Nil(t, ps.DeleteTopic("testid"))
}

func TestDelete_DBDeleteError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	topicsDB := mock.NewMockStore(ctrl)
	topicsDB.EXPECT().Delete("testid").Return(errors.New("delete failed"))
	ps := &pubsub.PubSub{TopicsDB: topicsDB, Logger: zap.NewNop()}

	assert.EqualError(t, ps.DeleteTopic("testid"), `Topic "testid" not found.`)
}

func TestGetAll_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	topicsDB := mock.NewMockStore(ctrl)
	topicsDB.EXPECT().List("").Return([]*store.KVPair{{
		Key:   "",
		Value: []byte(`{"topicId":"test"}`),
	}}, nil)
	ps := &pubsub.PubSub{TopicsDB: topicsDB, Logger: zap.NewNop()}

	topics, err := ps.GetAllTopics()

	assert.Equal(t, []*pubsub.Topic{&pubsub.Topic{ID: "test"}}, topics)
	assert.Nil(t, err)
}

func TestGetAll_EmptyListOnDBListError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	topicsDB := mock.NewMockStore(ctrl)
	topicsDB.EXPECT().List("").Return(nil, errors.New("db failed"))
	ps := &pubsub.PubSub{TopicsDB: topicsDB, Logger: zap.NewNop()}

	topics, err := ps.GetAllTopics()

	assert.Equal(t, []*pubsub.Topic{}, topics)
	assert.Nil(t, err)
}
