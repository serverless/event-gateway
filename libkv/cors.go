package libkv

import (
	"bytes"
	"encoding/json"
	"net/url"
	"path"

	validator "gopkg.in/go-playground/validator.v9"

	"go.uber.org/zap"

	"github.com/serverless/event-gateway/subscription/cors"
	"github.com/serverless/libkv/store"
)

// CORSKey is a key under which CORS data is stored KV store.
type CORSKey struct {
	Space string
	ID    cors.ID
}

func (key CORSKey) String() string {
	return key.Space + "/" + string(key.ID)
}

// CreateCORS creates CORS configuration.
func (service Service) CreateCORS(config *cors.CORS) (*cors.CORS, error) {
	if err := validateCORS(config); err != nil {
		return nil, err
	}

	config.ID = newCORSID(config)
	_, err := service.CORSStore.Get(CORSKey{config.Space, config.ID}.String(), &store.ReadOptions{Consistent: true})
	if err == nil {
		return nil, &cors.ErrCORSAlreadyExists{ID: config.ID}
	}

	byt, err := json.Marshal(config)
	if err != nil {
		return nil, &cors.ErrCORSValidation{Message: err.Error()}
	}

	err = service.CORSStore.Put(CORSKey{config.Space, config.ID}.String(), byt, nil)
	if err != nil {
		return nil, err
	}

	service.Log.Debug("CORS configuration created.", zap.Object("cors", config))

	return config, nil
}

// GetCORS returns function from configuration.
func (service Service) GetCORS(space string, id cors.ID) (*cors.CORS, error) {
	kv, err := service.CORSStore.Get(CORSKey{space, id}.String(), &store.ReadOptions{Consistent: true})
	if err != nil {
		if err.Error() == errKeyNotFound {
			return nil, &cors.ErrCORSNotFound{ID: id}
		}
		return nil, err
	}

	config := cors.CORS{}
	dec := json.NewDecoder(bytes.NewReader(kv.Value))
	err = dec.Decode(&config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// GetCORSes returns an array of all CORS configuration in the space.
func (service Service) GetCORSes(space string) (cors.CORSes, error) {
	configs := []*cors.CORS{}

	kvs, err := service.CORSStore.List(spacePath(space), &store.ReadOptions{Consistent: true})
	if err != nil && err.Error() != errKeyNotFound {
		return nil, err
	}

	for _, kv := range kvs {
		config := &cors.CORS{}
		dec := json.NewDecoder(bytes.NewReader(kv.Value))
		err = dec.Decode(config)
		if err != nil {
			return nil, err
		}

		configs = append(configs, config)
	}

	return cors.CORSes(configs), nil
}

// UpdateCORS updates CORS configuration.
func (service Service) UpdateCORS(config *cors.CORS) (*cors.CORS, error) {
	if err := validateCORS(config); err != nil {
		return nil, err
	}

	oldConfig, err := service.GetCORS(config.Space, config.ID)
	if err != nil {
		return nil, err
	}

	err = validateCORSUpdate(config, oldConfig)
	if err != nil {
		return nil, err
	}

	buf, err := json.Marshal(config)
	if err != nil {
		return nil, &cors.ErrCORSValidation{Message: err.Error()}
	}

	err = service.CORSStore.Put(CORSKey{config.Space, config.ID}.String(), buf, nil)
	if err != nil {
		return nil, err
	}

	service.Log.Debug("CORS updated.", zap.Object("cors", config))

	return config, nil
}

// DeleteCORS deletes CORS config from the configuration.
func (service Service) DeleteCORS(space string, id cors.ID) error {
	if err := service.CORSStore.Delete(CORSKey{space, id}.String()); err != nil {
		return &cors.ErrCORSNotFound{ID: id}
	}

	service.Log.Debug("CORS deleted.", zap.String("space", space), zap.String("id", string(id)))

	return nil
}

func validateCORS(config *cors.CORS) error {
	if config.Space == "" {
		config.Space = defaultSpace
	}

	if len(config.AllowedOrigins) == 0 {
		config.AllowedOrigins = []string{"*"}
	}

	if len(config.AllowedMethods) == 0 {
		config.AllowedMethods = []string{"HEAD", "GET", "POST"}
	}

	if len(config.AllowedHeaders) == 0 {
		config.AllowedHeaders = []string{"Origin", "Accept", "Content-Type"}
	}

	validate := validator.New()
	validate.RegisterValidation("space", spaceValidator)
	validate.RegisterValidation("path", pathValidator)
	err := validate.Struct(config)
	if err != nil {
		return &cors.ErrCORSValidation{Message: err.Error()}
	}

	return nil
}

func validateCORSUpdate(new *cors.CORS, old *cors.CORS) error {
	if new.Method != old.Method {
		return &cors.ErrInvalidCORSUpdate{Field: "Method"}
	}
	if new.Path != old.Path {
		return &cors.ErrInvalidCORSUpdate{Field: "Path"}
	}

	return nil
}

func newCORSID(config *cors.CORS) cors.ID {
	return cors.ID(config.Method + url.PathEscape(config.Path))
}

// pathValidator validates if field contains URL path
func pathValidator(fl validator.FieldLevel) bool {
	return path.IsAbs(fl.Field().String())
}
