package kv

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/serverless/event-gateway/functions/mock"
	"github.com/serverless/libkv/store"
	"github.com/stretchr/testify/assert"
)

func TestPrefixedStoreList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	kvs := []*store.KVPair{
		&store.KVPair{Key: "testroot/", Value: []byte(nil)},
		&store.KVPair{Key: "testroot/testdir/key1", Value: []byte("value1")},
		&store.KVPair{Key: "testroot/testdir/key2", Value: []byte("value2")},
	}
	kv := mock.NewMockStore(ctrl)
	kv.EXPECT().List("testroot/testdir").Return(kvs, nil)
	ps := NewPrefixedStore("testroot", kv)

	values, err := ps.List("testdir")
	assert.Nil(t, err)
	assert.Equal(t, []*store.KVPair{
		&store.KVPair{Key: "testdir/key1", Value: []byte("value1")},
		&store.KVPair{Key: "testdir/key2", Value: []byte("value2")},
	}, values)
}

func TestPrefixedStoreList_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	kv := mock.NewMockStore(ctrl)
	kv.EXPECT().List("testroot/key").Return(nil, errors.New("KV error"))
	ps := NewPrefixedStore("testroot", kv)

	values, err := ps.List("key")
	assert.Nil(t, values)
	assert.EqualError(t, err, "KV error")
}
