package store

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/serverless/event-gateway/mock"
	"github.com/serverless/libkv/store"
	"github.com/stretchr/testify/assert"
)

func TestPrefixedList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	kvs := []*store.KVPair{
		&store.KVPair{Key: "testroot/", Value: []byte(nil)},
		&store.KVPair{Key: "testroot/testdir/key1", Value: []byte("value1")},
		&store.KVPair{Key: "testroot/testdir/key2", Value: []byte("value2")},
	}
	kv := mock.NewMockStore(ctrl)
	kv.EXPECT().List("testroot/testdir", &store.ReadOptions{Consistent: true}).Return(kvs, nil)
	ps := NewPrefixed("testroot", kv)

	values, err := ps.List("testdir", &store.ReadOptions{Consistent: true})
	assert.Nil(t, err)
	assert.Equal(t, []*store.KVPair{
		&store.KVPair{Key: "testdir/key1", Value: []byte("value1")},
		&store.KVPair{Key: "testdir/key2", Value: []byte("value2")},
	}, values)
}

func TestPrefixedList_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	kv := mock.NewMockStore(ctrl)
	kv.EXPECT().List("testroot/key", nil).Return(nil, errors.New("KV error"))
	ps := NewPrefixed("testroot", kv)

	values, err := ps.List("key", nil)
	assert.Nil(t, values)
	assert.EqualError(t, err, "KV error")
}
