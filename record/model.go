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
		Count() *Model
		Create() *Model
		Delete() *Model
		Exec(ctx context.Context, useReplica bool) (int64, error)
		Find() *Model
		Limit(limit int) *Model
		Offset(offset int) *Model
		Order(order string) *Model
		Select(columns string) *Model
		SQL() string
		Update() *Model
		Where(condition string, args ...interface{}) *Model
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
		where                string
		whereArgs            []interface{}
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
	m.queryType = "all"
	m.queryBuilder.WriteString("SELECT * FROM ")
	m.queryBuilder.WriteString(m.tableName)
	m.queryBuilder.WriteString(";")

	return m
}

// Count returns the total count of matching records. Note that this can cause
// performance issue if there are too many data rows in the model's table.
func (m *Model) Count() *Model {
	m.queryType = "count"
	m.queryBuilder.WriteString("SELECT COUNT(")

	if m.selectColumns != "" {
		m.queryBuilder.WriteString(m.selectColumns)
	} else {
		m.queryBuilder.WriteString("*")
	}

	m.queryBuilder.WriteString(") FROM ")
	m.queryBuilder.WriteString(m.tableName)

	if m.where != "" {
		m.queryBuilder.WriteString(" WHERE ")
		m.queryBuilder.WriteString(m.where)
	}

	m.queryBuilder.WriteString(";")

	return m
}

// Create inserts the model object(s) into the database.
func (m *Model) Create() *Model {
	m.queryType = "create"
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

	m.queryBuilder.WriteString(";")

	return m
}

// Exec can execute the query with/without context/replica which will return
// the affected rows and error if there is any.
func (m *Model) Exec(ctx context.Context, useReplica bool) (int64, error) {
	var (
		count               int64
		err                 error
		result              sql.Result
		rows                *Rows
		db, master, replica DBer
	)

	if len(m.masters) > 0 {
		master = m.masters[rand.Intn(len(m.masters))]
	}

	if len(m.replicas) > 0 {
		replica = m.replicas[rand.Intn(len(m.replicas))]
	}

	if master == nil {
		return int64(0), ErrModelMissingMasterDB
	}

	if useReplica && replica == nil {
		return int64(0), ErrModelMissingReplicaDB
	}

	if m.queryBuilder.String() == "" {
		return int64(0), ErrModelEmptyQueryBuilder
	}

	db = master
	if useReplica && replica != nil {
		db = replica
	}

	switch m.queryType {
	case "count":
		if m.tx != nil {
			if ctx != nil {
				err = m.tx.GetContext(ctx, &count, m.queryBuilder.String(), m.whereArgs...)
			} else {
				err = m.tx.Get(&count, m.queryBuilder.String(), m.whereArgs...)
			}
		} else {
			if ctx != nil {
				err = db.GetContext(ctx, &count, m.queryBuilder.String(), m.whereArgs...)
			} else {
				err = db.Get(&count, m.queryBuilder.String(), m.whereArgs...)
			}
		}
	case "create":
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
					result, err = db.NamedExecContext(ctx, m.queryBuilder.String(), m.dest)
				} else {
					result, err = db.NamedExec(m.queryBuilder.String(), m.dest)
				}
			}

			if err != nil {
				return int64(0), err
			}

			lastInsertID, err := result.LastInsertId()
			if err != nil {
				return int64(0), err
			}

			count, err = result.RowsAffected()
			if err != nil {
				return int64(0), err
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
					rows, err = db.NamedQueryContext(ctx, m.queryBuilder.String(), m.dest)
				} else {
					rows, err = db.NamedQuery(m.queryBuilder.String(), m.dest)
				}
			}

			if err != nil {
				return int64(0), err
			}

			defer rows.Close()

			switch m.destKind {
			case reflect.Array, reflect.Slice:
				v := reflect.ValueOf(m.dest)
				count = int64(v.Len())

				for i := 0; i < v.Len(); i++ {
					err = m.scanPrimaryKeys(rows, v.Index(i))

					if err != nil {
						return int64(0), err
					}
				}
			case reflect.Ptr:
				count = int64(1)
				err = m.scanPrimaryKeys(rows, reflect.ValueOf(m.dest).Elem())

				if err != nil {
					return int64(0), err
				}
			}
		}
	case "all", "find":
		switch m.destKind {
		case reflect.Array, reflect.Slice:
			if m.tx != nil {
				if ctx != nil {
					err = m.tx.SelectContext(ctx, m.dest, m.queryBuilder.String(), m.whereArgs...)
				} else {
					err = m.tx.Select(m.dest, m.queryBuilder.String(), m.whereArgs...)
				}
			} else {
				if ctx != nil {
					err = db.SelectContext(ctx, m.dest, m.queryBuilder.String(), m.whereArgs...)
				} else {
					err = db.Select(m.dest, m.queryBuilder.String(), m.whereArgs...)
				}
			}

			count = int64(reflect.ValueOf(m.dest).Elem().Len())
		case reflect.Ptr:
			if m.tx != nil {
				if ctx != nil {
					err = m.tx.GetContext(ctx, m.dest, m.queryBuilder.String(), m.whereArgs...)
				} else {
					err = m.tx.Get(m.dest, m.queryBuilder.String(), m.whereArgs...)
				}
			} else {
				if ctx != nil {
					err = db.GetContext(ctx, m.dest, m.queryBuilder.String(), m.whereArgs...)
				} else {
					err = db.Get(m.dest, m.queryBuilder.String(), m.whereArgs...)
				}
			}

			if err == nil {
				count = 1
			}
		}
	}

	return count, err
}

// Delete deletes the records from the database.
func (m *Model) Delete() *Model {
	return m
}

// Find retrieves the records from the database.
func (m *Model) Find() *Model {
	m.queryType = "find"
	m.queryBuilder.WriteString("SELECT ")

	if m.selectColumns != "" {
		m.queryBuilder.WriteString(m.selectColumns)
	} else {
		m.queryBuilder.WriteString("*")
	}

	m.queryBuilder.WriteString(" FROM ")
	m.queryBuilder.WriteString(m.tableName)

	if m.where != "" {
		m.queryBuilder.WriteString(" WHERE ")
		m.queryBuilder.WriteString(m.where)
	}

	if m.order != "" {
		m.queryBuilder.WriteString(" ORDER BY ")
		m.queryBuilder.WriteString(m.order)
	}

	if m.limit != 0 {
		m.queryBuilder.WriteString(" LIMIT ")
		m.queryBuilder.WriteString(strconv.Itoa(m.limit))
	}

	if m.offset != 0 {
		m.queryBuilder.WriteString(" OFFSET ")
		m.queryBuilder.WriteString(strconv.Itoa(m.offset))
	}

	m.queryBuilder.WriteString(";")

	return m
}

// Limit indicates the number of records to retrieve from the database.
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
func (m *Model) Select(columns string) *Model {
	m.selectColumns = columns

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
func (m *Model) Where(condition string, args ...interface{}) *Model {
	m.whereArgs = args
	m.where = condition

	if strings.Contains(m.where, " IN (?)") && len(m.whereArgs) > 0 {
		var builder strings.Builder
		newArgs := []interface{}{}
		count := 0

		for _, char := range condition {
			var (
				arg  interface{}
				kind reflect.Kind
			)

			if count < len(m.whereArgs) {
				arg = m.whereArgs[count]
				kind = reflect.TypeOf(arg).Kind()
			}

			if char == '?' {
				if kind == reflect.Array || kind == reflect.Slice {
					args := reflect.ValueOf(arg)
					builder.WriteString(strings.Trim(strings.Repeat("?,", args.Len()), ","))

					for i := 0; i < args.Len(); i++ {
						newArgs = append(newArgs, args.Index(i).Interface())
					}
				} else {
					builder.WriteString(string(char))

					if arg != nil {
						newArgs = append(newArgs, arg)
					}
				}

				count++
				continue
			}

			builder.WriteString(string(char))
		}

		m.whereArgs = newArgs
		m.where = builder.String()
	}

	if m.adapter == "postgres" {
		var builder strings.Builder
		count := 0

		for _, char := range m.where {
			if char == '?' {
				builder.WriteString("$")
				builder.WriteString(strconv.Itoa(count + 1))
				count++
				continue
			}

			builder.WriteString(string(char))
		}

		m.where = builder.String()
	}

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
