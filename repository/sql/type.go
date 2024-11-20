package sql

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/gowool/pages/internal"
	"github.com/gowool/pages/model"
)

type Metas []model.Meta

func NewMetas(metas []model.Meta) Metas {
	if metas == nil {
		metas = Metas{}
	}
	return metas
}

func (m *Metas) Scan(src any) error {
	switch src := src.(type) {
	case string:
		return json.Unmarshal(internal.Bytes(src), m)
	case []byte:
		return json.Unmarshal(src, m)
	default:
		return errors.New("invalid src type for Metas")
	}
}

func (m *Metas) Value() (driver.Value, error) {
	if m == nil || len(*m) == 0 {
		return "[]", nil
	}
	raw, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return internal.String(raw), nil
}

type StrMap map[string]string

func NewStrMap(strMap map[string]string) StrMap {
	if strMap == nil {
		strMap = make(map[string]string)
	}
	return strMap
}

func (m *StrMap) Scan(src any) error {
	switch src := src.(type) {
	case string:
		return json.Unmarshal(internal.Bytes(src), m)
	case []byte:
		return json.Unmarshal(src, m)
	default:
		return errors.New("invalid src type for StrMap")
	}
}

func (m *StrMap) Value() (driver.Value, error) {
	if m == nil || len(*m) == 0 {
		return "{}", nil
	}
	raw, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return internal.String(raw), nil
}
