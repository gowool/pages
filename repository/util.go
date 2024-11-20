package repository

import (
	"fmt"
	"time"

	"github.com/gowool/cr"
)

func LifeSpanConditions(alias string, now time.Time) []any {
	if now.IsZero() {
		return nil
	}
	now = now.Truncate(60 * time.Second)

	if alias != "" && alias[len(alias)-1] != '.' {
		alias = fmt.Sprintf("%s.", alias)
	}

	return []any{
		fmt.Sprintf("%spublished IS NOT NULL", alias),
		cr.Condition{Column: fmt.Sprintf("%spublished", alias), Operator: cr.OpLte, Value: now},
		cr.Filter{
			Operator: cr.OpOR,
			Conditions: []any{
				fmt.Sprintf("%sexpired IS NULL", alias),
				cr.Condition{Column: fmt.Sprintf("%sexpired", alias), Operator: cr.OpGt, Value: now},
			},
		},
	}
}

func LifeSpanSort(alias string) cr.SortBy {
	if alias != "" && alias[len(alias)-1] != '.' {
		alias = fmt.Sprintf("%s.", alias)
	}

	return cr.SortBy{
		{Column: fmt.Sprintf("%spublished", alias), Order: "ASC NULLS LAST"},
		{Column: fmt.Sprintf("%sexpired", alias), Order: "ASC NULLS LAST"},
	}
}
