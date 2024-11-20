package sql

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/gowool/pages/internal"
	"github.com/gowool/pages/model"
	"github.com/gowool/pages/repository"
)

var _ repository.Configuration = (*ConfigurationRepository)(nil)

type ConfigurationDriver interface {
	SelectQuery(ctx context.Context, metadata Metadata) (query string, err error)
	SaveQuery(ctx context.Context, metadata Metadata, values map[string]any) (query string, args []any, err error)
}

type ConfigurationRepository struct {
	DB       *sql.DB
	Driver   ConfigurationDriver
	Metadata Metadata
	OnError  func(error) error
}

func NewConfigurationRepository(db *sql.DB, driver ConfigurationDriver) *ConfigurationRepository {
	return &ConfigurationRepository{
		DB:     db,
		Driver: driver,
		Metadata: Metadata{
			TableName:   "pages_configuration",
			PrimaryKeys: []string{"key"},
			Columns:     []string{"key", "value"},
		},
	}
}

func (r *ConfigurationRepository) Load(ctx context.Context) (model.Configuration, error) {
	query, err := r.Driver.SelectQuery(ctx, r.Metadata)
	if err != nil {
		return model.Configuration{}, r.Error(err)
	}

	rows, err := r.DBTx(ctx).QueryContext(ctx, query)
	if err != nil {
		return model.Configuration{}, r.Error(err)
	}
	defer func() {
		_ = rows.Close()
	}()

	rawData := make(map[string]json.RawMessage)
	for rows.Next() {
		var key, value string
		if err = rows.Scan(&key, &value); err != nil {
			return model.Configuration{}, r.Error(err)
		}
		rawData[key] = internal.Bytes(value)
	}

	raw, err := json.Marshal(rawData)
	if err != nil {
		return model.Configuration{}, r.Error(err)
	}

	m := model.NewConfiguration()
	err = json.Unmarshal(raw, &m)

	return m, r.Error(err)
}

func (r *ConfigurationRepository) Save(ctx context.Context, m *model.Configuration) error {
	raw, err := json.Marshal(m)
	if err != nil {
		return r.Error(err)
	}

	var rawData map[string]json.RawMessage
	if err = json.Unmarshal(raw, &rawData); err != nil {
		return r.Error(err)
	}

	data := make(map[string]any)
	for key, value := range rawData {
		data[key] = internal.String(value)
	}

	query, args, err := r.Driver.SaveQuery(ctx, r.Metadata, data)
	if err != nil {
		return r.Error(err)
	}

	_, err = r.DB.ExecContext(ctx, query, args...)
	return r.Error(err)
}

func (r *ConfigurationRepository) Error(err error) error {
	if err != nil && r.OnError != nil {
		return r.OnError(err)
	}
	return err
}

func (r *ConfigurationRepository) DBTx(ctx context.Context) DB {
	if db := Tx(ctx); db != nil {
		return db
	}
	return r.DB
}
