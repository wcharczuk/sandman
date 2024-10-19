package dbutil

import (
	"bytes"
	"fmt"
	"reflect"
	"text/template"

	"go.charczuk.com/sdk/db"
)

// QueryFormatter is a type that can format queries.
type QueryFormatter struct {
	vars           map[string]any
	templates      []string
	funcs          template.FuncMap
	models         map[string]any
	modelTypeMeta  map[string]*db.TypeMeta
	modelTableName map[string]string

	t *template.Template
}

// WithTemplates adds snippets or template to the formatter.
func (qf *QueryFormatter) WithTemplates(templates ...string) *QueryFormatter {
	qf.templates = append(qf.templates, templates...)
	return qf
}

// WithFunc adds a view func to the formatter.
func (qf *QueryFormatter) WithFunc(name string, fn any) *QueryFormatter {
	if qf.funcs == nil {
		qf.funcs = make(template.FuncMap)
	}
	qf.funcs[name] = fn
	return qf
}

// WithModels registers model types for use in view func helpers.
func (qf *QueryFormatter) WithModels(models ...any) *QueryFormatter {
	if qf.models == nil {
		qf.models = make(map[string]any)
	}
	if qf.modelTypeMeta == nil {
		qf.modelTypeMeta = make(map[string]*db.TypeMeta)
	}
	if qf.modelTableName == nil {
		qf.modelTableName = make(map[string]string)
	}
	for _, m := range models {
		modelName := reflect.TypeOf(m).Name()
		qf.models[modelName] = m
		qf.modelTypeMeta[modelName] = db.TypeMetaFor(m)
		qf.modelTableName[modelName] = db.TableName(m)
	}
	return qf
}

// MustFormatQuery formats a query with a given variadic set of options and panics on error.
func (qf *QueryFormatter) MustFormat(body string, vars any) (output string) {
	var err error
	output, err = qf.FormatQuery(body, vars)
	if err != nil {
		panic(err)
	}
	return output
}

func (qf *QueryFormatter) Initialize() *QueryFormatter {
	if qf.t == nil {
		qf.t = template.New("")
		if qf.funcs == nil {
			qf.funcs = make(template.FuncMap)
		}
		qf.funcs["columns"] = func(modelName string) (string, error) {
			columns, ok := qf.modelTypeMeta[modelName]
			if !ok {
				return "", fmt.Errorf("invalid model; %s not found", modelName)
			}
			return db.ColumnNamesCSV(columns.Columns()), nil
		}
		qf.funcs["columns_alias"] = func(modelName, alias string) (string, error) {
			columns, ok := qf.modelTypeMeta[modelName]
			if !ok {
				return "", fmt.Errorf("invalid model; %s not found", modelName)
			}
			return db.ColumnNamesFromAliasCSV(columns.Columns(), alias), nil
		}
		qf.funcs["columns_prefix_alias"] = func(modelName, prefix, alias string) (string, error) {
			columns, ok := qf.modelTypeMeta[modelName]
			if !ok {
				return "", fmt.Errorf("invalid model; %s not found", modelName)
			}
			return db.ColumnNamesWithPrefixFromAliasCSV(columns.Columns(), prefix, alias), nil
		}
		qf.funcs["table"] = func(modelName string) (string, error) {
			tableName, ok := qf.modelTableName[modelName]
			if !ok {
				return "", fmt.Errorf("invalid model; %s not found", modelName)
			}
			return tableName, nil
		}
		qf.t = qf.t.Funcs(qf.funcs)
		var err error
		for _, template := range qf.templates {
			qf.t, err = qf.t.Parse(template)
			if err != nil {
				err = fmt.Errorf("cannot parse template; %s: %w", template, err)
				panic(err)
			}
		}
	}
	return qf
}

// FormatQuery formats a query with a given variadic set of options.
func (qf *QueryFormatter) FormatQuery(body string, vars any) (output string, err error) {
	qf.Initialize()
	var tmpl *template.Template
	tmpl, err = qf.t.Parse(body)
	if err != nil {
		err = fmt.Errorf("cannot parse query; %s: %w", body, err)
		return
	}
	buf := new(bytes.Buffer)
	if err = tmpl.Execute(buf, vars); err != nil {
		err = fmt.Errorf("cannot execute query; %s: %w", body, err)
		return
	}
	output = buf.String()
	return
}
