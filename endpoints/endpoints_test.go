package endpoints_test

import (
	"errors"
	"testing"

	"github.com/docker/libkv/store"
	"github.com/golang/mock/gomock"
	"github.com/serverless/event-gateway/endpoints"
	"github.com/serverless/event-gateway/endpoints/mock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestCreate_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := mock.NewMockStore(ctrl)
	db.EXPECT().Get("GET-test").Return(nil, errors.New("not found"))
	db.EXPECT().Put("GET-test", []byte(`{"endpointId":"GET-test","functionId":"test","method":"GET","path":"test"}`), nil).Return(nil)
	fundb := mock.NewMockStore(ctrl)
	fundb.EXPECT().Exists("test").Return(true, nil)
	registry := &endpoints.Endpoints{DB: db, Logger: zap.NewNop(), FunctionsDB: fundb}

	en, _ := registry.Create(&endpoints.Endpoint{
		FunctionID: "test",
		Method:     "GET",
		Path:       "/test",
	})

	assert.Equal(t, &endpoints.Endpoint{
		ID:         "GET-test",
		FunctionID: "test",
		Method:     "GET",
		Path:       "test",
	}, en)
}

func TestCreate_EndpointAlreadyExistsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := mock.NewMockStore(ctrl)
	db.EXPECT().Get("GET-test").Return(nil, nil)
	fundb := mock.NewMockStore(ctrl)
	fundb.EXPECT().Exists("test").Return(true, nil)
	registry := &endpoints.Endpoints{DB: db, Logger: zap.NewNop(), FunctionsDB: fundb}

	_, err := registry.Create(&endpoints.Endpoint{
		FunctionID: "test",
		Method:     "GET",
		Path:       "test",
	})

	assert.EqualError(t, err, `Endpoint with method "GET" and path "test" already exits.`)
}

func TestCreate_DBPutError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := mock.NewMockStore(ctrl)
	db.EXPECT().Get("GET-test").Return(nil, errors.New("not found"))
	db.EXPECT().Put(gomock.Any(), gomock.Any(), nil).Return(errors.New("db put failed"))
	fundb := mock.NewMockStore(ctrl)
	fundb.EXPECT().Exists("test").Return(true, nil)
	registry := &endpoints.Endpoints{DB: db, Logger: zap.NewNop(), FunctionsDB: fundb}

	_, err := registry.Create(&endpoints.Endpoint{
		FunctionID: "test",
		Method:     "GET",
		Path:       "test",
	})

	assert.EqualError(t, err, "db put failed")
}

func TestCreate_FunctionNotFoundError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := mock.NewMockStore(ctrl)
	fundb := mock.NewMockStore(ctrl)
	fundb.EXPECT().Exists("test").Return(false, nil)
	registry := &endpoints.Endpoints{DB: db, Logger: zap.NewNop(), FunctionsDB: fundb}

	_, err := registry.Create(&endpoints.Endpoint{
		FunctionID: "test",
		Method:     "GET",
		Path:       "test",
	})

	assert.EqualError(t, err, `Function "test" not found.`)
}

func TestDelete_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := mock.NewMockStore(ctrl)
	db.EXPECT().Delete("testid").Return(nil)
	fundb := mock.NewMockStore(ctrl)
	registry := &endpoints.Endpoints{DB: db, Logger: zap.NewNop(), FunctionsDB: fundb}

	assert.Nil(t, registry.Delete("testid"))
}

func TestDelete_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := mock.NewMockStore(ctrl)
	db.EXPECT().Delete("testid").Return(errors.New("delete failed"))
	fundb := mock.NewMockStore(ctrl)
	registry := &endpoints.Endpoints{DB: db, Logger: zap.NewNop(), FunctionsDB: fundb}

	assert.EqualError(t, registry.Delete("testid"), `Endpoint "testid" not found.`)
}

func TestGetAll_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := mock.NewMockStore(ctrl)
	db.EXPECT().List("").Return([]*store.KVPair{{
		Key:   "",
		Value: []byte(`{"endpointId":"GET-test","functionId":"test","method":"GET","path":"test"}`),
	}}, nil)
	fundb := mock.NewMockStore(ctrl)
	registry := &endpoints.Endpoints{DB: db, Logger: zap.NewNop(), FunctionsDB: fundb}

	ens, err := registry.GetAll()

	assert.Equal(t, []*endpoints.Endpoint{
		&endpoints.Endpoint{
			ID:         "GET-test",
			FunctionID: "test",
			Method:     "GET",
			Path:       "test",
		}}, ens)
	assert.Nil(t, err)
}

func TestGetAll_EmptyListOnDBListError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := mock.NewMockStore(ctrl)
	db.EXPECT().List("").Return(nil, errors.New("db failed"))
	fundb := mock.NewMockStore(ctrl)
	registry := &endpoints.Endpoints{DB: db, Logger: zap.NewNop(), FunctionsDB: fundb}

	ens, err := registry.GetAll()

	assert.Equal(t, []*endpoints.Endpoint{}, ens)
	assert.Nil(t, err)
}
