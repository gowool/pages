package pg

import (
	"context"
	"fmt"

	"github.com/gowool/cr"

	"github.com/gowool/pages/repository/sql"
)

type SequenceDriver struct{}

func NewSequenceDriver() sql.SequenceDriver {
	return SequenceDriver{}
}

func (SequenceDriver) InsertQuery(_ context.Context, metadata sql.Metadata) (string, error) {
	return fmt.Sprintf(insertSeqSQL, metadata.TableName), nil
}

func (SequenceDriver) DeleteQuery(_ context.Context, metadata sql.Metadata, filter cr.Filter) (string, []any, error) {
	where, args := toWhere(filter)

	return fmt.Sprintf(deleteSQL, metadata.TableName) + where, args, nil
}
