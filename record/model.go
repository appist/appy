package record

import (
	"context"
	"database/sql"
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
		Exec(ctx context.Context) error
		SQL() string
		Update() *Model
		Where() *Model
	}

	// Model is the layer that represents business data and logic.
	Model struct {
		adapter              string
		attrs                map[string]modelAttr
		autoIncrementStField string
		dest                 interface{}
		destKind             reflect.Kind
		masters              []DBer
		replicas             []DBer
		tableName            string
		primaryKeys          []string
		queryBuilder         strings.Builder
		queryType            string
		tx                   Txer
		selectColumns        string
		limit                int
		offset               int
		order                string
		wheres               []string
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

	model := &Model{
		adapter:     "",
		attrs:       map[string]modelAttr{},
		dest:        dest,
		destKind:    destKind,
		masters:     []DBer{},
		replicas:    []DBer{},
		primaryKeys: []string{},
		tableName:   support.ToSnakeCase(support.Plural(e.Name())),
	}

	for i := 0; i < e.NumField(); i++ {
		field := e.Field(i)

		switch field.Type.String() {
		case "record.Modeler":
			for _, name := range strings.Split(field.Tag.Get("masters"), ",") {
				if dbManager.DB(name) != nil {
					model.masters = append(model.masters, dbManager.DB(name))
				}
			}

			if len(model.masters) > 0 {
				model.adapter = model.masters[0].Config().Adapter
			}

			for _, name := range strings.Split(field.Tag.Get("replicas"), ",") {
				if dbManager.DB(name) != nil {
					model.replicas = append(model.replicas, dbManager.DB(name))
				}
			}

			if model.adapter == "" && len(model.replicas) > 0 {
				model.adapter = model.replicas[0].Config().Adapter
			}

			tblName := field.Tag.Get("tableName")
			if tblName != "" {
				model.tableName = tblName
			}

			pks := field.Tag.Get("primaryKeys")
			if pks != "" {
				model.primaryKeys = strings.Split(pks, ",")
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
				continue
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
						continue
					}

					attr.autoIncrement = autoIncrement

					if autoIncrement {
						model.autoIncrementStField = field.Name
					}
				}
			}

			model.attrs[dbColumn] = attr
		}
	}

	return model
}

// All returns all records from the model's table. Note that this can cause
// performance issue if there are too many data rows in the model's table.
func (m *Model) All() *Model {
	m.queryType = "select"
	m.queryBuilder.WriteString("SELECT * FROM ")
	m.queryBuilder.WriteString(m.tableName)

	return m
}

// Create inserts the model object(s) into the database.
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

// Exec executes the query with/without context and returns error if there is
// any.
func (m *Model) Exec(ctx context.Context) error {
	var (
		err             error
		result          sql.Result
		rows            *Rows
		master, replica DBer
	)

	if len(m.masters) > 0 {
		master = m.masters[rand.Intn(len(m.masters))]
	}

	if len(m.replicas) > 0 {
		replica = m.replicas[rand.Intn(len(m.replicas))]
	}

	if master == nil && replica == nil {
		return ErrMissingModelDB
	}

	switch m.queryType {
	case "insert":
		if m.destKind == reflect.Array || m.destKind == reflect.Slice {
			m.dest = reflect.ValueOf(m.dest).Elem().Interface()
		}

		switch m.adapter {
		case "mysql":
			if m.tx != nil {
				if ctx != nil {
					result, err = m.tx.NamedExecContext(ctx, m.queryBuilder.String(), m.dest)
				} else {
					result, err = m.tx.NamedExec(m.queryBuilder.String(), m.dest)
				}
			} else {
				if ctx != nil {
					result, err = master.NamedExecContext(ctx, m.queryBuilder.String(), m.dest)
				} else {
					result, err = master.NamedExec(m.queryBuilder.String(), m.dest)
				}
			}

			if err != nil {
				return err
			}

			lastInsertID, err := result.LastInsertId()
			if err != nil {
				return err
			}

			if m.autoIncrementStField != "" {
				switch m.destKind {
				case reflect.Array, reflect.Slice:
					v := reflect.ValueOf(m.dest)

					for i := 0; i < v.Len(); i++ {
						v.Index(i).FieldByName(m.autoIncrementStField).SetInt(lastInsertID + int64(i))
					}
				case reflect.Ptr:
					reflect.ValueOf(m.dest).Elem().FieldByName(m.autoIncrementStField).SetInt(lastInsertID)
				}
			}
		case "postgres":
			if m.tx != nil {
				if ctx != nil {
					rows, err = m.tx.NamedQueryContext(ctx, m.queryBuilder.String(), m.dest)
				} else {
					rows, err = m.tx.NamedQuery(m.queryBuilder.String(), m.dest)
				}
			} else {
				if ctx != nil {
					rows, err = master.NamedQueryContext(ctx, m.queryBuilder.String(), m.dest)
				} else {
					rows, err = master.NamedQuery(m.queryBuilder.String(), m.dest)
				}
			}

			if err != nil {
				return err
			}

			defer rows.Close()

			switch m.destKind {
			case reflect.Array, reflect.Slice:
				v := reflect.ValueOf(m.dest)

				for i := 0; i < v.Len(); i++ {
					err = m.scanPrimaryKeys(rows, v.Index(i))

					if err != nil {
						return err
					}
				}
			case reflect.Ptr:
				err = m.scanPrimaryKeys(rows, reflect.ValueOf(m.dest).Elem())

				if err != nil {
					return err
				}
			}
		}
	case "select":
		if m.tx != nil {
			if ctx != nil {
				err = m.tx.SelectContext(ctx, m.dest, m.queryBuilder.String())
			} else {
				err = m.tx.Select(m.dest, m.queryBuilder.String())
			}
		} else {
			db := replica

			if db == nil {
				db = master
			}

			if ctx != nil {
				err = db.SelectContext(ctx, m.dest, m.queryBuilder.String())
			} else {
				err = db.Select(m.dest, m.queryBuilder.String())
			}
		}
	}

	return err
}

// Limit indicates the number of recrods to retrieve from the database.
func (m *Model) Limit(limit int) *Model {
	m.limit = limit

	return m
}

// Offset indicates the number of records to skip before starting to return
// the records.
func (m *Model) Offset(offset int) *Model {
	m.offset = offset

	return m
}

// Order indicates the specific order to retrieve records from the database.
func (m *Model) Order(order string) *Model {
	m.order = order

	return m
}

// Select selects only a subset of fields from the result set.
func (m *Model) Select(selectColumns string) *Model {
	m.selectColumns = selectColumns

	return m
}

// SQL returns the SQL string.
func (m *Model) SQL() string {
	return m.queryBuilder.String()
}

// Update updates the model object(s) in the database.
func (m *Model) Update() *Model {
	return m
}

// Where indicates the condition of which records to return.
func (m *Model) Where() *Model {
	return m
}

func (m *Model) scanPrimaryKeys(rows *Rows, v reflect.Value) error {
	if !rows.Next() {
		return nil
	}

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
		v.FieldByName(m.attrs[primaryKey].stFieldName).Set(reflect.ValueOf(values[idx]).Elem())
	}

	return nil
}
