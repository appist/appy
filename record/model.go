package record

import (
	"context"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/appist/appy/support"
)

type (
	// Modeler implements all Model methods.
	Modeler interface {
		All() *Model
		Create() *Model
		Exec() error
		Update() *Model
	}

	// Model is the layer that represents business data and logic.
	Model struct {
		adapter      string
		attrs        map[string]modelAttr
		ctx          context.Context
		dest         interface{}
		destKind     reflect.Kind
		masters      []DBer
		replicas     []DBer
		tableName    string
		primaryKeys  []string
		queryBuilder strings.Builder
		queryType    string
		tx           Txer
		selectFields string
		limit        int
		offset       int
		order        string
		wheres       []string
	}

	modelAttr struct {
		autoIncrement bool
		ignored       bool
		stFieldName   string
		stFieldType   reflect.Type
	}
)

func init() {
	// For model to choose random masters/replicas.
	rand.Seed(time.Now().Unix())
}

// NewModel initializes a model that represents business data and logic.
func NewModel(dbManager *Engine, dest interface{}) Modeler {
	t := reflect.TypeOf(dest)
	e := t.Elem()
	destKind := t.Kind()

	if e.Kind() == reflect.Array || e.Kind() == reflect.Slice {
		destKind = e.Kind()
		e = e.Elem()
	}

	adapter := ""
	attrs := map[string]modelAttr{}
	masters := []DBer{}
	replicas := []DBer{}
	primaryKeys := []string{}
	tableName := support.ToSnakeCase(support.Plural(e.Name()))

	for i := 0; i < e.NumField(); i++ {
		field := e.Field(i)

		switch field.Type.String() {
		case "record.Modeler":
			for _, name := range strings.Split(field.Tag.Get("masters"), ",") {
				if dbManager.DB(name) != nil {
					masters = append(masters, dbManager.DB(name))
				}
			}

			if len(masters) > 0 {
				adapter = masters[0].Config().Adapter
			}

			for _, name := range strings.Split(field.Tag.Get("replicas"), ",") {
				if dbManager.DB(name) != nil {
					replicas = append(replicas, dbManager.DB(name))
				}
			}

			if adapter == "" && len(replicas) > 0 {
				adapter = replicas[0].Config().Adapter
			}

			tblName := field.Tag.Get("tableName")
			if tblName != "" {
				tableName = tblName
			}

			pks := field.Tag.Get("primaryKeys")
			if pks != "" {
				primaryKeys = strings.Split(pks, ",")
			}
		default:
			dbColumn := support.ToSnakeCase(field.Name)
			attr := modelAttr{
				autoIncrement: false,
				ignored:       false,
				stFieldName:   field.Name,
				stFieldType:   field.Type,
			}

			// SQLX uses db tag to retrieve the column name.
			dbTag := field.Tag.Get("db")
			if dbTag == "-" {
				attr.ignored = true
			}

			if dbTag != "" {
				dbColumn = dbTag
			}

			ormTag := field.Tag.Get("orm")
			ormAttrs := strings.Split(ormTag, ";")
			for _, ormAttr := range ormAttrs {
				splits := strings.Split(ormAttr, ":")

				switch splits[0] {
				case "auto_increment":
					autoIncrement, err := strconv.ParseBool(splits[1])
					if err != nil {
						break
					}

					attr.autoIncrement = autoIncrement
				}
			}

			attrs[dbColumn] = attr
		}
	}

	model := &Model{
		adapter:     adapter,
		attrs:       attrs,
		dest:        dest,
		destKind:    destKind,
		masters:     masters,
		replicas:    replicas,
		primaryKeys: primaryKeys,
		tableName:   tableName,
	}

	return model
}

func (m *Model) All() *Model {
	m.queryType = "select"
	m.queryBuilder.WriteString("SELECT * FROM ")
	m.queryBuilder.WriteString(m.tableName)

	return m
}

func (m *Model) Create() *Model {
	m.queryType = "insert"
	m.queryBuilder.WriteString("INSERT INTO ")
	m.queryBuilder.WriteString(m.tableName)
	m.queryBuilder.WriteString(" (")

	var (
		count  = 0
		values strings.Builder
	)

	for dbColumn, attr := range m.attrs {
		if attr.ignored || attr.autoIncrement {
			continue
		}

		values.WriteString(":")
		values.WriteString(dbColumn)
		m.queryBuilder.WriteString(dbColumn)

		if count < len(m.attrs)-2 {
			values.WriteString(", ")
			m.queryBuilder.WriteString(", ")
		}

		count++
	}

	m.queryBuilder.WriteString(") VALUES (")
	m.queryBuilder.WriteString(values.String())
	m.queryBuilder.WriteString(")")

	if len(m.primaryKeys) > 0 && m.adapter == "postgres" {
		m.queryBuilder.WriteString(" RETURNING ")
		m.queryBuilder.WriteString(strings.Join(m.primaryKeys, ", "))
	}

	return m
}

func (m *Model) Exec() error {
	var (
		err  error
		rows *Rows
	)

	switch m.queryType {
	case "insert":
		if m.destKind == reflect.Array || m.destKind == reflect.Slice {
			m.dest = reflect.ValueOf(m.dest).Elem().Interface()
		}

		if m.tx != nil {
			rows, err = m.tx.NamedQuery(m.queryBuilder.String(), m.dest)
		} else {
			rows, err = m.masters[rand.Intn(len(m.masters))].NamedQuery(m.queryBuilder.String(), m.dest)
		}

		if err != nil {
			return err
		}

		defer rows.Close()

		switch m.destKind {
		case reflect.Array, reflect.Slice:
			v := reflect.ValueOf(m.dest)

			for i := 0; i < v.Len(); i++ {
				if rows.Next() {
					columns, err := rows.Columns()
					if err != nil {
						return err
					}

					values := make([]interface{}, len(columns))
					for idx, column := range columns {
						values[idx] = reflect.New(m.attrs[column].stFieldType).Interface()
					}

					err = rows.Scan(values...)
					if err != nil {
						return err
					}

					for idx, primaryKey := range m.primaryKeys {
						v.Index(i).FieldByName(m.attrs[primaryKey].stFieldName).Set(reflect.ValueOf(values[idx]).Elem())
					}
				}
			}
		case reflect.Ptr:
			if rows.Next() {
				columns, err := rows.Columns()
				if err != nil {
					return err
				}

				values := make([]interface{}, len(columns))
				for idx, column := range columns {
					values[idx] = reflect.New(m.attrs[column].stFieldType).Interface()
				}

				err = rows.Scan(values...)
				if err != nil {
					return err
				}

				v := reflect.ValueOf(m.dest).Elem()
				for idx, primaryKey := range m.primaryKeys {
					v.FieldByName(m.attrs[primaryKey].stFieldName).Set(reflect.ValueOf(values[idx]).Elem())
				}
			}
		}
	case "select":
		if m.tx != nil {
			err = m.tx.Select(m.dest, m.queryBuilder.String())
		} else {
			err = m.masters[rand.Intn(len(m.masters))].Select(m.dest, m.queryBuilder.String())
		}
	}

	return err
}

func (m *Model) Limit(limit int) *Model {
	m.limit = limit

	return m
}

func (m *Model) Offset(offset int) *Model {
	m.offset = offset

	return m
}

func (m *Model) Order(order string) *Model {
	m.order = order

	return m
}

func (m *Model) Select(selectFields string) *Model {
	m.selectFields = selectFields

	return m
}

func (m *Model) SQL() string {
	return m.queryBuilder.String()
}

func (m *Model) Update() *Model {
	return m
}
