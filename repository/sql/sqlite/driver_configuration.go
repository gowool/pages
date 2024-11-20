package sqlite

import (
	"context"
	"fmt"
	"strings"

	"github.com/gowool/pages/repository/sql"
)

type ConfigurationDriver struct{}

func NewConfigurationDriver() sql.ConfigurationDriver {
	return ConfigurationDriver{}
}

func (ConfigurationDriver) SelectQuery(_ context.Context, metadata sql.Metadata) (string, error) {
	return fmt.Sprintf(selectSQL, metadata.Select(), metadata.TableName), nil
}

func (ConfigurationDriver) SaveQuery(_ context.Context, metadata sql.Metadata, data map[string]any) (string, []any, error) {
	values := make([]string, 0, len(data))
	args := make([]any, 0, len(data)*2)

	for key, value := range data {
		args = append(args, key, value)
		values = append(values, fmt.Sprintf("($%d,$%d)", len(args)-1, len(args)))
	}

	return fmt.Sprintf(cfgInsertSQL, metadata.TableName, strings.Join(values, ",")), args, nil
}
