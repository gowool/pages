package fs

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/gowool/cr"
	"github.com/spf13/cast"
	"golang.org/x/exp/constraints"

	"github.com/gowool/pages/internal"
	"github.com/gowool/pages/model"
	"github.com/gowool/pages/repository"
)

type Template struct {
	Path string // without left / (slash) or . (dot) or @
	FS   fs.FS
	Info fs.FileInfo
}

func (t Template) ID() int64 {
	return -internal.ToInt64(t.Name())
}

func (t Template) Name() string {
	return fmt.Sprintf("@%s", t.Path)
}

func (t Template) Model() (model.Template, error) {
	content, err := fs.ReadFile(t.FS, t.Path)
	if err != nil {
		return model.Template{}, err
	}

	return model.Template{
		ID:      t.ID(),
		Name:    t.Name(),
		Content: internal.String(content),
		Type:    model.TemplateFS,
		Enabled: true,
		Created: fileCreated(t.Info),
		Updated: t.Info.ModTime(),
	}, nil
}

type TemplateRepository struct {
	repository.Template
	ext  string
	fs   fs.FS
	data map[string]model.Template
}

func NewTemplateRepository(inner repository.Template, fsys fs.FS, ext ...string) *TemplateRepository {
	if len(ext) == 0 || ext[0] == "" {
		ext = []string{".gohtml"}
	}
	return &TemplateRepository{Template: inner, ext: ext[0], fs: fsys, data: make(map[string]model.Template)}
}

func (r *TemplateRepository) FindByID(ctx context.Context, id int64) (model.Template, error) {
	template, err := r.Template.FindByID(ctx, id)
	if err == nil {
		return template, nil
	}

	err1 := fs.WalkDir(r.fs, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if filepath.Ext(path) != r.ext {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		if !info.Mode().IsRegular() {
			return nil
		}

		t := &Template{
			Path: path,
			Info: info,
			FS:   r.fs,
		}

		if t.ID() == id {
			template, err = t.Model()
			if err != nil {
				return err
			}
			return fs.SkipAll
		}
		return nil
	})

	if err1 != nil && !errors.Is(err1, fs.SkipAll) {
		err = errors.Join(sql.ErrNoRows, repository.ErrNotFound, err, err1)
	}

	if template.ID == 0 {
		return template, err
	}
	return template, nil
}

func (r *TemplateRepository) FindByName(ctx context.Context, name string) (model.Template, error) {
	template, err := r.Template.FindByName(ctx, name)
	if err == nil {
		return template, nil
	}

	path := strings.TrimLeft(name, "@/.")

	info, err1 := fs.Stat(r.fs, path)
	if err1 != nil {
		return template, errors.Join(sql.ErrNoRows, repository.ErrNotFound, err, err1)
	}

	if info.IsDir() || !info.Mode().IsRegular() {
		return template, errors.Join(sql.ErrNoRows, repository.ErrNotFound, err, fmt.Errorf("%s is not a file", name))
	}

	t := Template{
		Path: path,
		Info: info,
		FS:   r.fs,
	}

	template, err1 = t.Model()
	if err1 != nil {
		return template, errors.Join(sql.ErrNoRows, repository.ErrNotFound, err, err1)
	}
	return template, nil
}

func (r *TemplateRepository) FindAndCount(ctx context.Context, criteria *cr.Criteria) ([]model.Template, int, error) {
	if criteria == nil {
		criteria = &cr.Criteria{}
	}

	data, total, err := r.Template.FindAndCount(ctx, criteria)
	if err != nil {
		return nil, 0, err
	}

	size := -len(data)
	if criteria.Size != nil {
		size += *criteria.Size
	}
	items, count, err := r.fsDiff(size, criteria.Filter)
	if err != nil {
		return nil, 0, err
	}

	data = append(data, items...)

	sortBy(data, criteria.SortBy)

	return data, total + count, nil
}

func (r *TemplateRepository) Find(ctx context.Context, criteria *cr.Criteria) ([]model.Template, error) {
	if criteria == nil {
		criteria = &cr.Criteria{}
	}

	data, err := r.Template.Find(ctx, criteria)
	if err != nil {
		return nil, err
	}

	size := -len(data)
	if criteria.Size != nil {
		size += *criteria.Size
	}
	items, _, err := r.fsDiff(size, criteria.Filter)
	if err != nil {
		return nil, err
	}

	data = append(data, items...)

	sortBy(data, criteria.SortBy)

	return data, nil
}

func (r *TemplateRepository) fsDiff(n int, f cr.Filter) ([]model.Template, int, error) {
	templates, err := r.walk()
	if err != nil {
		return nil, 0, err
	}

	templates = filter(templates, f)

	size := len(templates)
	if n >= 0 {
		size = min(n, size)
	}
	data := make([]model.Template, size)

	for i := 0; i < len(data); i++ {
		item, err := templates[i].Model()
		if err != nil {
			return nil, 0, err
		}
		data[i] = item
	}

	return data, len(templates), nil
}

func (r *TemplateRepository) walk() ([]*Template, error) {
	templates := make([]*Template, 0, 10)
	if err := fs.WalkDir(r.fs, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if filepath.Ext(path) != r.ext {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		if !info.Mode().IsRegular() {
			return nil
		}

		templates = append(templates, &Template{
			Path: path,
			Info: info,
			FS:   r.fs,
		})

		return nil
	}); err != nil {
		return nil, err
	}

	return slices.Clip(templates), nil
}

func sortBy(templates []model.Template, sBy cr.SortBy) {
	if len(sBy) == 0 {
		return
	}

	sort.Slice(templates, func(i, j int) bool {
		for _, s := range sBy {
			parts := strings.SplitN(s.Column, ".", 2)
			col := parts[0]
			if len(parts) > 1 {
				col = parts[1]
			}

			switch col {
			case "id":
				if c := compare(templates[i].ID, templates[j].ID); c != nil {
					return c(s.Order)
				}
			case "name":
				if c := compare(templates[i].Name, templates[j].Name); c != nil {
					return c(s.Order)
				}
			case "type":
				if c := compare(templates[i].Type, templates[j].Type); c != nil {
					return c(s.Order)
				}
			case "enabled":
				if templates[i].Enabled == templates[j].Enabled {
					continue
				}
				if s.Order == "ASC" {
					return templates[i].Enabled && !templates[j].Enabled
				}
				return !templates[i].Enabled && templates[j].Enabled
			case "created":
				if c := compareDates(templates[i].Created, templates[j].Created); c != nil {
					return c(s.Order)
				}
			case "updated":
				if c := compareDates(templates[i].Updated, templates[j].Updated); c != nil {
					return c(s.Order)
				}
			}
		}

		return false
	})
}

func compare[T constraints.Ordered](a, b T) func(string) bool {
	if a == b {
		return nil
	}
	return func(order string) bool {
		if order == "ASC" {
			return a < b
		}
		return a > b
	}
}

func compareDates(a, b time.Time) func(string) bool {
	if a.IsZero() && b.IsZero() || a.Equal(b) {
		return nil
	}
	return func(order string) bool {
		if order == "ASC" {
			return a.Before(b)
		}
		return a.After(b)
	}
}

func filter(templates []*Template, f cr.Filter) []*Template {
	if len(f.Conditions) == 0 {
		return templates
	}

	return internal.Filter(templates, func(item *Template) bool {
		result := f.Operator != cr.OpOR

		for _, cond := range f.Conditions {
			switch c := cond.(type) {
			case cr.Condition:
				parts := strings.SplitN(c.Column, ".", 2)
				if len(parts) == 1 && parts[0] != "name" {
					// support only name column
					return false
				} else if len(parts) > 1 && parts[1] != "name" {
					// support only name column
					return false
				}
				r := true
				switch c.Operator {
				case cr.OpEqual:
					r = item.Name() == fmt.Sprintf("%v", c.Value)
				case cr.OpNotEqual:
					r = item.Name() != fmt.Sprintf("%v", c.Value)
				case cr.OpIN:
					r = slices.Contains(cast.ToStringSlice(c.Value), item.Name())
				case cr.OpNOT.Append(cr.OpIN):
					r = !slices.Contains(cast.ToStringSlice(c.Value), item.Name())
				case cr.OpLIKE:
					r = like(item.Name(), c.Value)
				case cr.OpNOT.Append(cr.OpLIKE):
					r = !like(item.Name(), c.Value)
				default:
					if f.Operator != cr.OpOR {
						return false
					}
				}

				if f.Operator == cr.OpOR {
					if r {
						return true
					}
				} else {
					result = result && r

					if !result {
						return false
					}
				}
			default:
				// not support
				return false
			}
		}

		return result
	})
}

func like(s string, search any) bool {
	s = strings.ToLower(s)
	substr := strings.ToLower(fmt.Sprintf("%v", search))

	if substr[0] == '%' && substr[len(substr)-1] == '%' {
		return strings.Contains(s, substr[1:len(substr)-1])
	}

	if substr[0] == '%' {
		return strings.HasSuffix(s, substr[1:])
	}

	if substr[len(substr)-1] == '%' {
		return strings.HasPrefix(s, substr[:len(substr)-1])
	}

	return s == substr
}
