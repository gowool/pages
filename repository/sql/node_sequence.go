package sql

import (
	"context"
	"database/sql"

	"github.com/gowool/cr"

	"github.com/gowool/pages/repository"
)

var _ repository.SequenceNode = (*SequenceNodeRepository)(nil)

type SequenceDriver interface {
	InsertQuery(ctx context.Context, metadata Metadata) (query string, err error)
	DeleteQuery(ctx context.Context, metadata Metadata, filter cr.Filter) (query string, args []any, err error)
}

type SequenceNodeRepository struct {
	DB       *sql.DB
	Driver   SequenceDriver
	Metadata Metadata
}

func NewSequenceNodeRepository(db *sql.DB, driver SequenceDriver) *SequenceNodeRepository {
	return &SequenceNodeRepository{
		DB:     db,
		Driver: driver,
		Metadata: Metadata{
			TableName: "sequence_nodes",
			Columns:   []string{"id"},
		},
	}
}

func (r *SequenceNodeRepository) Create(ctx context.Context) (int64, error) {
	query, err := r.Driver.InsertQuery(ctx, r.Metadata)
	if err != nil {
		return 0, err
	}

	row := r.DBTx(ctx).QueryRowContext(ctx, query)
	if row.Err() != nil {
		return 0, row.Err()
	}

	var id int64
	err = row.Scan(&id)
	return id, err
}

func (r *SequenceNodeRepository) Delete(ctx context.Context, ids ...int64) error {
	filter := cr.Filter{
		Conditions: []any{
			cr.Condition{Column: r.Metadata.PKs()[0], Operator: cr.OpIN, Value: ids},
		},
	}

	query, args, err := r.Driver.DeleteQuery(ctx, r.Metadata, filter)
	if err != nil {
		return err
	}

	_, err = r.DBTx(ctx).ExecContext(ctx, query, args...)
	return err
}

func (r *SequenceNodeRepository) DBTx(ctx context.Context) DB {
	if db := Tx(ctx); db != nil {
		return db
	}
	return r.DB
}
