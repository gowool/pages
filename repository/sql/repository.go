package sql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/gowool/cr"

	"github.com/gowool/pages/repository"
)

func WithTx(ctx context.Context, tx *sql.Tx) context.Context {
	return context.WithValue(ctx, ctxTxKey{}, tx)
}

func Tx(ctx context.Context) *sql.Tx {
	tx, _ := ctx.Value(ctxTxKey{}).(*sql.Tx)
	return tx
}

type (
	ctxTxKey struct{}
	DB       interface {
		QueryContext(context.Context, string, ...any) (*sql.Rows, error)
		QueryRowContext(context.Context, string, ...any) *sql.Row
		ExecContext(context.Context, string, ...any) (sql.Result, error)
	}
)

type Metadata struct {
	TableName   string
	PrimaryKeys []string
	Columns     []string
	MinSize     int
}

func (m Metadata) PKs() []string {
	if len(m.PrimaryKeys) == 0 {
		return []string{"id"}
	}
	return m.PrimaryKeys
}

func (m Metadata) Select() string {
	if len(m.Columns) == 0 {
		return "*"
	}
	return strings.Join(m.Columns, ",")
}

func (m Metadata) minSize() int {
	if m.MinSize == 0 {
		return 20
	}
	return m.MinSize
}

type Driver interface {
	SelectQuery(ctx context.Context, metadata Metadata, criteria *cr.Criteria) (query string, args []any, err error)
	CountQuery(ctx context.Context, metadata Metadata, filter cr.Filter) (query string, args []any, err error)
	DeleteQuery(ctx context.Context, metadata Metadata, filter cr.Filter) (query string, args []any, err error)
	UpdateQuery(ctx context.Context, metadata Metadata, values map[string]any, filter cr.Filter) (query string, args []any, err error)
	InsertQuery(ctx context.Context, metadata Metadata, values map[string]any) (query string, args []any, err error)
	IsUniqueViolation(err error) bool
}

type Repository[T interface{ GetID() ID }, ID any] struct {
	DB           *sql.DB
	Metadata     Metadata
	Driver       Driver
	RowScan      func(interface{ Scan(dest ...any) error }, *T) error
	InsertValues func(*T) map[string]any
	UpdateValues func(*T) map[string]any
	OnError      func(error) error
}

func (r Repository[T, ID]) FindAndCount(ctx context.Context, criteria *cr.Criteria) ([]T, int, error) {
	if criteria == nil {
		criteria = cr.New()
	}

	count, err := r.Count(ctx, criteria.Filter)
	if err != nil {
		return nil, 0, err
	}
	if count == 0 {
		return nil, 0, nil
	}

	data, err := r.Find(ctx, criteria)
	return data, count, err
}

func (r Repository[T, ID]) Find(ctx context.Context, criteria *cr.Criteria) ([]T, error) {
	if criteria == nil {
		criteria = cr.New()
	}

	query, args, err := r.Driver.SelectQuery(ctx, r.Metadata, criteria)
	if err != nil {
		return nil, r.Error(err)
	}

	rows, err := r.DBTx(ctx).QueryContext(ctx, query, args...)
	if err != nil {
		return nil, r.Error(err)
	}
	defer func() {
		_ = rows.Close()
	}()

	size := r.Metadata.minSize()
	if criteria.Size != nil {
		size = *criteria.Size
	}
	data := make([]T, 0, size)
	for rows.Next() {
		var item T
		if err = r.RowScan(rows, &item); err != nil {
			return nil, r.Error(err)
		}
		data = append(data, item)
	}
	return slices.Clip(data), nil
}

func (r Repository[T, ID]) Count(ctx context.Context, filter cr.Filter) (int, error) {
	query, args, err := r.Driver.CountQuery(ctx, r.Metadata, filter)
	if err != nil {
		return 0, r.Error(err)
	}

	var total int
	err = r.DBTx(ctx).QueryRowContext(ctx, query, args...).Scan(&total)
	return total, r.Error(err)
}

func (r Repository[T, ID]) FindByID(ctx context.Context, id ID) (T, error) {
	return r.FindBy(ctx, r.Metadata.PKs()[0], id)
}

func (r Repository[T, ID]) FindBy(ctx context.Context, column string, value any) (m T, err error) {
	criteria := cr.New().
		SetSize(1).
		SetFilter(
			cr.Filter{
				Conditions: []any{
					cr.Condition{Column: column, Operator: cr.OpEqual, Value: value},
				},
			},
		)

	data, err := r.Find(ctx, criteria)
	if err != nil {
		return m, err
	}
	if len(data) == 0 {
		return m, r.Error(sql.ErrNoRows)
	}
	return data[0], nil
}

func (r Repository[T, ID]) Create(ctx context.Context, m *T) error {
	if m == nil {
		return fmt.Errorf("sql: `%s`: create called with %w", r.typeName(), repository.ErrNil)
	}

	query, args, err := r.Driver.InsertQuery(ctx, r.Metadata, r.InsertValues(m))
	if err != nil {
		return r.Error(err)
	}

	row := r.DBTx(ctx).QueryRowContext(ctx, query, args...)
	return r.Error(r.RowScan(row, m))
}

func (r Repository[T, ID]) Update(ctx context.Context, m *T) error {
	if m == nil {
		return fmt.Errorf("sql: `%s`: update called with %w", r.typeName(), repository.ErrNil)
	}

	filter := cr.Filter{
		Conditions: []any{
			cr.Condition{Column: r.Metadata.PKs()[0], Operator: cr.OpEqual, Value: (*m).GetID()},
		},
	}

	query, args, err := r.Driver.UpdateQuery(ctx, r.Metadata, r.InsertValues(m), filter)
	if err != nil {
		return r.Error(err)
	}

	row := r.DBTx(ctx).QueryRowContext(ctx, query, args...)
	return r.Error(r.RowScan(row, m))
}

func (r Repository[T, ID]) Delete(ctx context.Context, ids ...ID) error {
	filter := cr.Filter{
		Conditions: []any{
			cr.Condition{Column: r.Metadata.PKs()[0], Operator: cr.OpIN, Value: ids},
		},
	}

	query, args, err := r.Driver.DeleteQuery(ctx, r.Metadata, filter)
	if err != nil {
		return r.Error(err)
	}

	_, err = r.DBTx(ctx).ExecContext(ctx, query, args...)
	return r.Error(err)
}

func (r Repository[T, ID]) Error(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		err = errors.Join(err, repository.ErrNotFound)
	} else if r.Driver.IsUniqueViolation(err) {
		err = errors.Join(err, repository.ErrUniqueViolation)
	}
	if r.OnError != nil {
		return r.OnError(err)
	}
	return err
}

func (r Repository[T, ID]) typeName() string {
	var t *T
	return reflect.TypeOf(t).Elem().Name()
}

func (r Repository[T, ID]) DBTx(ctx context.Context) DB {
	if db := Tx(ctx); db != nil {
		return db
	}
	return r.DB
}
