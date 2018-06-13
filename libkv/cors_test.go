package libkv

import (
	"errors"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/serverless/event-gateway/mock"
	_ "github.com/serverless/event-gateway/providers/http"
	"github.com/serverless/event-gateway/subscription/cors"
	"github.com/serverless/libkv/store"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestCreateCORS(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testID := cors.ID(`GET%2Fhello`)
	testCORSConfig := &cors.CORS{
		ID:             testID,
		Method:         http.MethodGet,
		Path:           "/hello",
		AllowedHeaders: []string{"content-type"},
		AllowedMethods: []string{"GET"},
		AllowedOrigins: []string{"*"},
	}

	t.Run("CORS config created", func(t *testing.T) {
		db := mock.NewMockStore(ctrl)
		db.EXPECT().
			Get(`default/GET%2Fhello`, &store.ReadOptions{Consistent: true}).
			Return(nil, errors.New("KV type not found"))
		payload := []byte(
			`{"space":"default","corsId":"GET%2Fhello","method":"GET","path":"/hello","allowedOrigins":["*"],` +
				`"allowedMethods":["GET"],"allowedHeaders":["content-type"],"allowCredentials":false}`)
		db.EXPECT().Put(`default/GET%2Fhello`, payload, nil).Return(nil)
		service := &Service{CORSStore: db, Log: zap.NewNop()}

		_, err := service.CreateCORS(testCORSConfig)

		assert.Nil(t, err)
	})

	t.Run("CORS config created with default values", func(t *testing.T) {
		db := mock.NewMockStore(ctrl)
		db.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, errors.New("KV type not found"))
		db.EXPECT().Put(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
		service := &Service{CORSStore: db, Log: zap.NewNop()}

		created, _ := service.CreateCORS(&cors.CORS{
			ID:     testID,
			Method: http.MethodGet,
			Path:   "/hello",
		})

		assert.Equal(t, []string{"*"}, created.AllowedOrigins)
		assert.Equal(t, []string{"Origin", "Accept", "Content-Type"}, created.AllowedHeaders)
		assert.Equal(t, []string{"HEAD", "GET", "POST"}, created.AllowedMethods)
		assert.Equal(t, false, created.AllowCredentials)
	})

	t.Run("CORS configuration for specified method and path already exists", func(t *testing.T) {
		db := mock.NewMockStore(ctrl)
		db.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, nil)
		service := &Service{CORSStore: db, Log: zap.NewNop()}

		_, err := service.CreateCORS(testCORSConfig)

		assert.Equal(t, &cors.ErrCORSAlreadyExists{ID: testID}, err)
	})

	t.Run("validation error", func(t *testing.T) {
		service := &Service{Log: zap.NewNop()}

		_, err := service.CreateCORS(&cors.CORS{
			Method:         "NOT",
			Path:           "/hello",
			AllowedHeaders: []string{"content-type"},
			AllowedMethods: []string{"GET"},
		})

		assert.Equal(t, &cors.ErrCORSValidation{
			Message: "Key: 'CORS.Method' Error:Field validation for 'Method' failed on the 'eq=GET|eq=POST|eq=DELETE|eq=PUT|eq=PATCH|eq=HEAD|eq=OPTIONS' tag",
		}, err)
	})

	t.Run("KV Put error", func(t *testing.T) {
		db := mock.NewMockStore(ctrl)
		db.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, errors.New("KV type not found"))
		db.EXPECT().Put(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("KV put error"))
		service := &Service{CORSStore: db, Log: zap.NewNop()}

		_, err := service.CreateCORS(testCORSConfig)

		assert.EqualError(t, err, "KV put error")
	})
}

func TestGetCORS(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testID := cors.ID(`GET%2Fhello`)
	testCORSConfig := &cors.CORS{
		Space:          "default",
		ID:             testID,
		Method:         http.MethodGet,
		Path:           "/hello",
		AllowedHeaders: []string{"content-type"},
		AllowedMethods: []string{"GET"},
		AllowedOrigins: []string{"*"},
	}
	testPayload := []byte(
		`{"space":"default","corsId":"GET%2Fhello","method":"GET","path":"/hello","allowedOrigins":["*"],` +
			`"allowedMethods":["GET"],"allowedHeaders":["content-type"],"allowCredentials":false}`)

	t.Run("CORS config returned", func(t *testing.T) {
		db := mock.NewMockStore(ctrl)
		db.EXPECT().
			Get(`default/GET%2Fhello`, &store.ReadOptions{Consistent: true}).
			Return(&store.KVPair{Value: testPayload}, nil)
		service := &Service{CORSStore: db, Log: zap.NewNop()}

		config, _ := service.GetCORS("default", cors.ID("GET%2Fhello"))

		assert.Equal(t, testCORSConfig, config)
	})

	t.Run("CORS not found", func(t *testing.T) {
		db := mock.NewMockStore(ctrl)
		db.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, errors.New("Key not found in store"))
		service := &Service{CORSStore: db, Log: zap.NewNop()}

		_, err := service.GetCORS("default", cors.ID("GET%2Fhello"))

		assert.Equal(t, &cors.ErrCORSNotFound{ID: "GET%2Fhello"}, err)
	})

	t.Run("KV Get error", func(t *testing.T) {
		db := mock.NewMockStore(ctrl)
		db.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, errors.New("KV get err"))
		service := &Service{CORSStore: db, Log: zap.NewNop()}

		_, err := service.GetCORS("default", cors.ID("GET%2Fhello"))

		assert.EqualError(t, err, "KV get err")
	})
}

func TestUpdateCORS(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	existingCORSKV := &store.KVPair{Value: []byte(
		`{"space":"default","corsId":"GET%2Fhello","method":"GET","path":"/hello","allowedOrigins":["*"],` +
			`"allowedMethods":["GET"],"allowedHeaders":["content-type"],"allowCredentials":false}`)}
	newConfigPayload := []byte(
		`{"space":"default","corsId":"GET%2Fhello","method":"GET","path":"/hello","allowedOrigins":` +
			`["http://example.com"],"allowedMethods":["GET"],"allowedHeaders":["content-type"],"allowCredentials":false}`)
	newConfig := &cors.CORS{
		Space:          "default",
		ID:             cors.ID("GET%2Fhello"),
		Method:         "GET",
		Path:           "/hello",
		AllowedHeaders: []string{"content-type"},
		AllowedMethods: []string{"GET"},
		AllowedOrigins: []string{"http://example.com"},
	}

	t.Run("CORS config updated", func(t *testing.T) {
		db := mock.NewMockStore(ctrl)
		db.EXPECT().
			Get("default/GET%2Fhello", &store.ReadOptions{Consistent: true}).
			Return(existingCORSKV, nil)
		db.EXPECT().Put("default/GET%2Fhello", newConfigPayload, nil).Return(nil)
		service := &Service{CORSStore: db, Log: zap.NewNop()}

		_, err := service.UpdateCORS(newConfig)

		assert.Nil(t, err)
	})

	t.Run("disallow updating method field", func(t *testing.T) {
		db := mock.NewMockStore(ctrl)
		db.EXPECT().Get(gomock.Any(), gomock.Any()).Return(existingCORSKV, nil)
		service := &Service{CORSStore: db, Log: zap.NewNop()}

		_, err := service.UpdateCORS(&cors.CORS{
			Space:          "default",
			ID:             cors.ID("GET%2Fhello"),
			Method:         "PUT",
			Path:           "/hello",
			AllowedHeaders: []string{"content-type"},
			AllowedMethods: []string{"GET"},
			AllowedOrigins: []string{"http://example.com"},
		})

		assert.Equal(t, &cors.ErrInvalidCORSUpdate{Field: "Method"}, err)
	})

	t.Run("disallow updating path field", func(t *testing.T) {
		db := mock.NewMockStore(ctrl)
		db.EXPECT().Get(gomock.Any(), gomock.Any()).Return(existingCORSKV, nil)
		service := &Service{CORSStore: db, Log: zap.NewNop()}

		_, err := service.UpdateCORS(&cors.CORS{
			Space:          "default",
			ID:             cors.ID("GET%2Fhello"),
			Method:         "GET",
			Path:           "/hello1",
			AllowedHeaders: []string{"content-type"},
			AllowedMethods: []string{"GET"},
			AllowedOrigins: []string{"http://example.com"},
		})

		assert.Equal(t, &cors.ErrInvalidCORSUpdate{Field: "Path"}, err)
	})

	t.Run("CORS config not found", func(t *testing.T) {
		db := mock.NewMockStore(ctrl)
		db.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, errors.New("Key not found in store"))
		service := &Service{CORSStore: db, Log: zap.NewNop()}

		_, err := service.UpdateCORS(newConfig)

		assert.Equal(t, &cors.ErrCORSNotFound{ID: "GET%2Fhello"}, err)
	})

	t.Run("KV Put error", func(t *testing.T) {
		db := mock.NewMockStore(ctrl)
		db.EXPECT().Get(gomock.Any(), gomock.Any()).Return(existingCORSKV, nil)
		db.EXPECT().Put(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("KV put error"))
		service := &Service{CORSStore: db, Log: zap.NewNop()}

		_, err := service.UpdateCORS(newConfig)

		assert.EqualError(t, err, "KV put error")
	})
}

func TestDeleteCORS(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("CORS deleted", func(t *testing.T) {
		db := mock.NewMockStore(ctrl)
		db.EXPECT().Delete("default/GET%2Fhello").Return(nil)
		service := &Service{CORSStore: db, Log: zap.NewNop()}

		err := service.DeleteCORS("default", cors.ID("GET%2Fhello"))

		assert.Nil(t, err)
	})

	t.Run("CORS not found", func(t *testing.T) {
		db := mock.NewMockStore(ctrl)
		db.EXPECT().Delete(gomock.Any()).Return(errors.New("KV func not found"))
		service := &Service{CORSStore: db, Log: zap.NewNop()}

		err := service.DeleteCORS("default", cors.ID("GET%2Fhello"))

		assert.Equal(t, &cors.ErrCORSNotFound{ID: "GET%2Fhello"}, err)
	})
}
