package pg

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/gowool/cr"

	"github.com/gowool/pages/repository/sql"
)

const (
	selectSQL       = "SELECT %s FROM %s"
	countSQL        = "SELECT COUNT(*) FROM %s"
	insertSQL       = "INSERT INTO %s (%s) VALUES (%s) RETURNING %s"
	insertSeqSQL    = "INSERT INTO %s values (DEFAULT) RETURNING id"
	cfgInsertSQL    = "INSERT INTO %s (key,value) VALUES %s ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value"
	updateSQL       = "UPDATE %s SET %s%s RETURNING %s"
	deleteSQL       = "DELETE FROM %s"
	uniqueViolation = "23505"
)

type Driver struct{}

func NewDriver() sql.Driver {
	return Driver{}
}

func (Driver) SelectQuery(_ context.Context, metadata sql.Metadata, criteria *cr.Criteria) (string, []any, error) {
	if criteria == nil {
		criteria = cr.New()
	}

	where, args := toWhere(criteria.Filter)
	index := len(args)

	var b strings.Builder
	b.WriteString(where)

	if len(criteria.SortBy) > 0 {
		b.WriteString(" ORDER BY ")
		b.WriteString(criteria.SortBy.String())
	}

	if criteria.Size != nil && *criteria.Size > 0 {
		b.WriteString(" LIMIT $")
		b.WriteString(strconv.Itoa(index + 1))
		b.WriteString(" OFFSET $")
		b.WriteString(strconv.Itoa(index + 2))

		size := *criteria.Size
		args = append(args, size, criteria.GetOffset())
	}

	return fmt.Sprintf(selectSQL, metadata.Select(), metadata.TableName) + b.String(), args, nil
}

func (Driver) CountQuery(_ context.Context, metadata sql.Metadata, filter cr.Filter) (string, []any, error) {
	where, args := toWhere(filter)

	return fmt.Sprintf(countSQL, metadata.TableName) + where, args, nil
}

func (Driver) DeleteQuery(_ context.Context, metadata sql.Metadata, filter cr.Filter) (string, []any, error) {
	where, args := toWhere(filter)

	return fmt.Sprintf(deleteSQL, metadata.TableName) + where, args, nil
}

func (Driver) UpdateQuery(_ context.Context, metadata sql.Metadata, data map[string]any, filter cr.Filter) (string, []any, error) {
	where, args := toWhere(filter)
	columns := make([]string, 0, len(data))

	for column, value := range data {
		args = append(args, value)
		columns = append(columns, fmt.Sprintf("%s = $%d", column, len(args)))
	}

	return fmt.Sprintf(
		updateSQL,
		metadata.TableName,
		strings.Join(columns, ","),
		where,
		metadata.Select(),
	), args, nil
}

func (Driver) InsertQuery(_ context.Context, metadata sql.Metadata, data map[string]any) (string, []any, error) {
	columns := make([]string, 0, len(data))
	values := make([]string, 0, len(data))
	args := make([]any, 0, len(data))

	for column, value := range data {
		columns = append(columns, column)
		args = append(args, value)
		values = append(values, fmt.Sprintf("$%d", len(args)))
	}

	return fmt.Sprintf(
		insertSQL,
		metadata.TableName,
		strings.Join(columns, ","),
		strings.Join(values, ","),
		metadata.Select(),
	), args, nil
}

func (Driver) IsUniqueViolation(err error) bool {
	return strings.Contains(err.Error(), uniqueViolation)
}

func toWhere(filter cr.Filter) (string, []any) {
	s, args := filter.ToSQL()
	if s == "" {
		return "", nil
	}

	var (
		where strings.Builder
		index int
	)

	where.WriteString(" WHERE ")
	for i := 0; i < len(s); i++ {
		if i > 0 && (i+2) < len(s) && s[i-1] == ' ' &&
			(s[i] == 'I' || s[i] == 'i') &&
			(s[i+1] == 'N' || s[i+1] == 'n') &&
			(s[i+2] == '?' || s[i+2] == ' ' || s[i+2] == '(') {
			where.WriteString("= ANY")
			i++
			continue
		}

		if s[i] == '?' {
			index++
			where.WriteString(fmt.Sprintf("$%d", index))
			continue
		}

		where.WriteByte(s[i])
	}

	return where.String(), args
}
