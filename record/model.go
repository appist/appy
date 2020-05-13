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
		Begin() error
		BeginContext(ctx context.Context, opts *sql.TxOptions) error
		Commit() error
		Count() *Model
		Create() *Model
		Delete() *Model
		Exec(opts ...ExecOption) (int64, error)
		Find() *Model
		Limit(limit int) *Model
		Offset(offset int) *Model
		Order(order string) *Model
		Rollback() error
		Select(columns string) *Model
		SQL() string
		Tx() Txer
		Update() *Model
		UpdateAll(set string, args ...interface{}) *Model
		Where(condition string, args ...interface{}) *Model
	}

	// Model is the layer that represents business data and logic.
	Model struct {
		adapter, autoIncrement, tableName                            string
		attrs                                                        map[string]*modelAttr
		dest, scanDest                                               interface{}
		destKind                                                     reflect.Kind
		masters, replicas                                            []DBer
		primaryKeys                                                  []string
		queryBuilder                                                 strings.Builder
		tx                                                           Txer
		limit, offset                                                int
		action, group, having, order, selectColumns, timezone, where string
		args, havingArgs, whereArgs                                  []interface{}
	}

	// ModelOption is used to initialise a model with additional configurations.
	ModelOption struct {
		tx Txer
	}

	// ExecOption indicates how a query should be executed.
	ExecOption struct {
		// Context can be used to set the query timeout.
		Context context.Context

		// UseReplica indicates if the query should use replica. Note that there
		// could be replica lag which won't allow recent inserted/updated data to
		// be returned correctly.
		UseReplica bool
	}

	modelAttr struct {
		autoIncrement bool
		ignored       bool
		primaryKey    bool
		stFieldName   string
		stFieldType   reflect.Type
	}
)

func init() {
	// For model to choose random masters/replicas.
	rand.Seed(time.Now().Unix())
}

// NewModel initializes a model that represents business data and logic.
func NewModel(dbManager *Engine, dest interface{}, opts ...ModelOption) Modeler {
	destType := reflect.TypeOf(dest)
	destKind := destType.Kind()
	destElem := destType.Elem()

	if destElem.Kind() == reflect.Array || destElem.Kind() == reflect.Slice {
		destKind = destElem.Kind()
		destElem = destElem.Elem()
	}

	model := &Model{
		adapter:       "",
		attrs:         map[string]*modelAttr{},
		dest:          dest,
		destKind:      destKind,
		masters:       []DBer{},
		replicas:      []DBer{},
		autoIncrement: "",
		primaryKeys:   []string{"id"},
		tableName:     support.ToSnakeCase(support.Plural(destElem.Name())),
		timezone:      "utc",
	}

	if len(opts) > 0 {
		model.tx = opts[0].tx
	}

	for i := 0; i < destElem.NumField(); i++ {
		field := destElem.Field(i)

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

			tz := field.Tag.Get("timezone")
			if tz != "" {
				model.timezone = tz
			}

			autoIncrement := field.Tag.Get("autoIncrement")
			if autoIncrement != "" {
				model.autoIncrement = autoIncrement
			}

			pks, ok := field.Tag.Lookup("primaryKeys")
			if ok && pks == "" {
				model.primaryKeys = []string{}
			}

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

			if dbColumn == model.autoIncrement {
				attr.autoIncrement = true
			}

			if support.ArrayContains(model.primaryKeys, dbColumn) {
				attr.primaryKey = true
			}

			model.attrs[dbColumn] = &attr
		}
	}

	return model
}

// All returns all records from the model's table. Use an array/slice of the
// struct to scan all the records. Otherwise, only the 1st record will be
// scanned into the single struct.
//
// Note that this can cause performance issue if there are too many data rows
// in the model's table.
func (m *Model) All() *Model {
	m.action = "all"
	m.queryBuilder.WriteString("SELECT * FROM ")
	m.queryBuilder.WriteString(m.tableName)
	m.queryBuilder.WriteString(";")

	return m
}

// Begin starts a transaction. The default isolation level is dependent on the
// driver.
func (m *Model) Begin() error {
	var err error

	if m.tx == nil {
		m.tx, err = m.masters[rand.Intn(len(m.masters))].Begin()
	}

	return err
}

// BeginContext starts a transaction.
//
// The provided context is used until the transaction is committed or rolled
// back. If the context is canceled, the sql package will roll back the
// transaction. Tx.Commit will return an error if the context provided to
// BeginContext is canceled.
//
// The provided TxOptions is optional and may be nil if defaults should be used.
// If a non-default isolation level is used that the driver doesn't support, an
// error will be returned.
func (m *Model) BeginContext(ctx context.Context, opts *sql.TxOptions) error {
	var err error

	if m.tx == nil {
		m.tx, err = m.masters[rand.Intn(len(m.masters))].BeginContext(ctx, opts)
	}

	return err
}

// Commit commits the transaction.
func (m *Model) Commit() error {
	var err error

	if m.tx != nil {
		err = m.tx.Commit()

		if err == nil {
			m.tx = nil
		}
	}

	return err
}

// Rollback aborts the transaction.
func (m *Model) Rollback() error {
	var err error

	if m.tx != nil {
		err = m.tx.Rollback()

		if err == nil {
			m.tx = nil
		}
	}

	return err
}

// Count returns the total count of matching records. Note that this can cause
// performance issue if there are too many data rows in the model's table.
func (m *Model) Count() *Model {
	m.action = "count"
	m.args = []interface{}{}
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
		m.args = append(m.args, m.whereArgs...)
	}

	m.queryBuilder.WriteString(";")

	return m
}

// Create inserts the model object(s) into the database.
func (m *Model) Create() *Model {
	m.action = "create"
	m.queryBuilder.WriteString("INSERT INTO ")
	m.queryBuilder.WriteString(m.tableName)
	m.queryBuilder.WriteString(" (")

	columns := []string{}
	values := []string{}
	for column, attr := range m.attrs {
		if attr.ignored || attr.autoIncrement {
			continue
		}

		columns = append(columns, column)
		values = append(values, ":"+column)
	}

	m.queryBuilder.WriteString(strings.Join(columns, ", "))
	m.queryBuilder.WriteString(") VALUES (")
	m.queryBuilder.WriteString(strings.Join(values, ", "))
	m.queryBuilder.WriteString(")")

	if len(m.primaryKeys) > 0 && m.adapter == "postgres" {
		m.queryBuilder.WriteString(" RETURNING ")
		m.queryBuilder.WriteString(strings.Join(m.primaryKeys, ", "))
	}

	m.queryBuilder.WriteString(";")

	return m
}

// Delete deletes the records from the database.
func (m *Model) Delete() *Model {
	m.action = "delete"
	m.args = []interface{}{}

	m.queryBuilder.WriteString("DELETE FROM ")
	m.queryBuilder.WriteString(m.tableName)

	if m.where == "" {
		m.buildWhereWithPrimaryKeys()
	}

	if m.where != "" {
		m.queryBuilder.WriteString(" WHERE ")
		m.queryBuilder.WriteString(m.where)
		m.args = append(m.args, m.whereArgs...)
	}

	m.queryBuilder.WriteString(";")

	return m
}

// Exec can execute the query with/without context/replica which will return
// the affected rows and error if there is any.
func (m *Model) Exec(opts ...ExecOption) (int64, error) {
	var (
		count               int64
		err                 error
		result              sql.Result
		rows                *Rows
		stmt                *Stmt
		db, master, replica DBer
	)

	opt := ExecOption{}
	if len(opts) > 0 {
		opt = opts[0]
	}

	if len(m.masters) > 0 {
		master = m.masters[rand.Intn(len(m.masters))]
	}

	if len(m.replicas) > 0 {
		replica = m.replicas[rand.Intn(len(m.replicas))]
	}

	if master == nil {
		return int64(0), ErrModelMissingMasterDB
	}

	if opt.UseReplica && replica == nil {
		return int64(0), ErrModelMissingReplicaDB
	}

	if m.queryBuilder.String() == "" {
		return int64(0), ErrModelEmptyQueryBuilder
	}

	db = master
	if opt.UseReplica && replica != nil {
		db = replica
	}

	query := m.queryBuilder.String()
	if m.adapter == "postgres" {
		var builder strings.Builder

		count := 0
		for _, char := range query {
			if char == '?' {
				builder.WriteString("$")
				builder.WriteString(strconv.Itoa(count + 1))
				count++
				continue
			}

			builder.WriteString(string(char))
		}

		query = builder.String()
	}

	// Reset the buffer so that the model instance can be re-used to execute
	// another query.
	m.queryBuilder.Reset()

	now := time.Now()
	switch m.timezone {
	case "local":
		now = now.Local()
	default:
		now = now.UTC()
	}

	switch m.action {
	case "delete", "update_all":
		if m.tx != nil {
			if opt.Context != nil {
				stmt, err = m.tx.PrepareContext(opt.Context, query)
				if err != nil {
					return int64(0), err
				}

				result, err = stmt.ExecContext(opt.Context, m.args...)
			} else {
				stmt, err = m.tx.Prepare(query)
				if err != nil {
					return int64(0), err
				}

				result, err = stmt.Exec(m.args...)
			}
		} else {
			if opt.Context != nil {
				stmt, err = db.PrepareContext(opt.Context, query)
				if err != nil {
					return int64(0), err
				}

				result, err = stmt.ExecContext(opt.Context, m.args...)
			} else {
				stmt, err = db.Prepare(query)
				if err != nil {
					return int64(0), err
				}

				result, err = stmt.Exec(m.args...)
			}
		}

		if err != nil {
			return int64(0), err
		}

		count, err = result.RowsAffected()
		if err != nil {
			return int64(0), err
		}
	case "count":
		if m.tx != nil {
			if opt.Context != nil {
				stmt, err = m.tx.PrepareContext(opt.Context, query)
				if err != nil {
					return int64(0), err
				}

				err = stmt.GetContext(opt.Context, &count, m.args...)
			} else {
				stmt, err = m.tx.Prepare(query)
				if err != nil {
					return int64(0), err
				}

				err = stmt.Get(&count, m.args...)
			}
		} else {
			if opt.Context != nil {
				stmt, err = db.PrepareContext(opt.Context, query)
				if err != nil {
					return int64(0), err
				}

				err = stmt.GetContext(opt.Context, &count, m.args...)
			} else {
				stmt, err = db.Prepare(query)
				if err != nil {
					return int64(0), err
				}

				err = stmt.Get(&count, m.args...)
			}
		}
	case "create":
		dest := m.dest

		switch m.destKind {
		case reflect.Array, reflect.Slice:
			v := reflect.ValueOf(dest).Elem()

			for i := 0; i < v.Len(); i++ {
				field := v.Index(i).FieldByName("CreatedAt")

				if field.IsValid() {
					field.Set(reflect.ValueOf(&now))
				}
			}

			dest = reflect.ValueOf(dest).Elem().Interface()
		case reflect.Ptr:
			v := reflect.ValueOf(dest).Elem()
			field := v.FieldByName("CreatedAt")

			if field.IsValid() {
				field.Set(reflect.ValueOf(&now))
			}
		}

		switch m.adapter {
		case "mysql":
			if m.tx != nil {
				if opt.Context != nil {
					result, err = m.tx.NamedExecContext(opt.Context, query, dest)
				} else {
					result, err = m.tx.NamedExec(query, dest)
				}
			} else {
				if opt.Context != nil {
					result, err = db.NamedExecContext(opt.Context, query, dest)
				} else {
					result, err = db.NamedExec(query, dest)
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

			if m.autoIncrement != "" {
				switch m.destKind {
				case reflect.Array, reflect.Slice:
					v := reflect.ValueOf(m.dest).Elem()

					for i := 0; i < v.Len(); i++ {
						v.Index(i).FieldByName(m.attrs[m.autoIncrement].stFieldName).SetInt(lastInsertID + int64(i))
					}
				case reflect.Ptr:
					reflect.ValueOf(m.dest).Elem().FieldByName(m.attrs[m.autoIncrement].stFieldName).SetInt(lastInsertID)
				}
			}
		case "postgres":
			if m.tx != nil {
				if opt.Context != nil {
					rows, err = m.tx.NamedQueryContext(opt.Context, query, dest)
				} else {
					rows, err = m.tx.NamedQuery(query, dest)
				}
			} else {
				if opt.Context != nil {
					rows, err = db.NamedQueryContext(opt.Context, query, dest)
				} else {
					rows, err = db.NamedQuery(query, dest)
				}
			}

			if err != nil {
				return int64(0), err
			}

			defer rows.Close()

			switch m.destKind {
			case reflect.Array, reflect.Slice:
				v := reflect.ValueOf(m.dest).Elem()
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
	case "all", "find", "scan":
		switch m.destKind {
		case reflect.Array, reflect.Slice:
			if m.tx != nil {
				if opt.Context != nil {
					stmt, err = m.tx.PrepareContext(opt.Context, query)
					if err != nil {
						return int64(0), err
					}

					err = stmt.SelectContext(opt.Context, m.dest, m.args...)
				} else {
					stmt, err = m.tx.Prepare(query)
					if err != nil {
						return int64(0), err
					}

					err = stmt.Select(m.dest, m.args...)
				}
			} else {
				if opt.Context != nil {
					stmt, err = db.PrepareContext(opt.Context, query)
					if err != nil {
						return int64(0), err
					}

					err = stmt.SelectContext(opt.Context, m.dest, m.args...)
				} else {
					stmt, err = db.Prepare(query)
					if err != nil {
						return int64(0), err
					}

					err = stmt.Select(m.dest, m.args...)
				}
			}

			count = int64(reflect.ValueOf(m.dest).Elem().Len())
		case reflect.Ptr:
			if m.tx != nil {
				if opt.Context != nil {
					stmt, err = m.tx.PrepareContext(opt.Context, query)
					if err != nil {
						return int64(0), err
					}

					err = stmt.GetContext(opt.Context, m.dest, m.args...)
				} else {
					stmt, err = m.tx.Prepare(query)
					if err != nil {
						return int64(0), err
					}

					err = stmt.Get(m.dest, m.args...)
				}
			} else {
				if opt.Context != nil {
					stmt, err = db.PrepareContext(opt.Context, query)
					if err != nil {
						return int64(0), err
					}

					err = stmt.GetContext(opt.Context, m.dest, m.args...)
				} else {
					stmt, err = db.Prepare(query)
					if err != nil {
						return int64(0), err
					}

					err = stmt.Get(m.dest, m.args...)
				}
			}

			if err == sql.ErrNoRows {
				err = nil
			}

			for _, pk := range m.primaryKeys {
				if !reflect.ValueOf(m.dest).Elem().FieldByName(m.attrs[pk].stFieldName).IsZero() {
					count = 1
					break
				}
			}
		}
	}

	return count, err
}

// Find retrieves the records from the database.
func (m *Model) Find() *Model {
	m.action = "find"
	m.args = []interface{}{}
	m.queryBuilder.WriteString("SELECT ")

	if m.selectColumns != "" {
		m.queryBuilder.WriteString(m.selectColumns)
	} else {
		m.queryBuilder.WriteString("*")
	}

	m.queryBuilder.WriteString(" FROM ")
	m.queryBuilder.WriteString(m.tableName)

	if m.where == "" {
		m.buildWhereWithPrimaryKeys()
		m.resetDest()
	}

	if m.where != "" {
		m.queryBuilder.WriteString(" WHERE ")
		m.queryBuilder.WriteString(m.where)
		m.args = append(m.args, m.whereArgs...)
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

// Group indicates how to group rows into subgroups based on values of columns
// or expressions.
func (m *Model) Group(group string) *Model {
	m.group = group

	return m
}

// Having indicates the filter conditions for a group of rows.
func (m *Model) Having(having string, args ...interface{}) *Model {
	m.having = having
	m.havingArgs = args

	return m
}

// Limit indicates the number limit of records to retrieve from the database.
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

// Scan allows custom select result being scanned into the specified dest.
func (m *Model) Scan(dest interface{}) *Model {
	m.action = "scan"
	m.args = []interface{}{}

	m.dest = dest
	tmpKind := reflect.TypeOf(dest).Elem().Kind()
	if tmpKind == reflect.Array || tmpKind == reflect.Slice {
		m.destKind = tmpKind
	}

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
		m.args = append(m.args, m.whereArgs...)
	}

	if m.group != "" {
		m.queryBuilder.WriteString(" GROUP BY ")
		m.queryBuilder.WriteString(m.group)
	}

	if m.having != "" {
		m.queryBuilder.WriteString(" HAVING ")
		m.queryBuilder.WriteString(m.having)
		m.args = append(m.args, m.havingArgs...)
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

// Select selects only a subset of fields from the result set.
func (m *Model) Select(columns string) *Model {
	m.selectColumns = columns

	return m
}

// SQL returns the SQL string.
func (m *Model) SQL() string {
	return m.queryBuilder.String()
}

// Tx returns the transaction connection.
func (m *Model) Tx() Txer {
	return m.tx
}

// Update updates the model object(s) into the database.
func (m *Model) Update() *Model {
	m.action = "update"
	m.queryBuilder.WriteString("UPDATE ")
	m.queryBuilder.WriteString(m.tableName)
	m.queryBuilder.WriteString(" SET ")

	sets := []string{}
	for column, attr := range m.attrs {
		if attr.ignored || attr.autoIncrement {
			continue
		}

		sets = append(sets, column+" = :"+column)
	}

	m.queryBuilder.WriteString(strings.Join(sets, ", "))

	if len(m.primaryKeys) > 0 && m.adapter == "postgres" {
		m.queryBuilder.WriteString(" RETURNING ")
		m.queryBuilder.WriteString(strings.Join(m.primaryKeys, ", "))
	}

	m.queryBuilder.WriteString(";")

	return m
}

// UpdateAll updates all the records that match where condition.
func (m *Model) UpdateAll(set string, args ...interface{}) *Model {
	m.action = "update_all"
	m.args = []interface{}{}

	m.queryBuilder.WriteString("UPDATE ")
	m.queryBuilder.WriteString(m.tableName)
	m.queryBuilder.WriteString(" SET ")
	m.queryBuilder.WriteString(set)
	m.args = append(m.args, args...)

	if m.where == "" {
		m.buildWhereWithPrimaryKeys()
	}

	if m.where != "" {
		m.queryBuilder.WriteString(" WHERE ")
		m.queryBuilder.WriteString(m.where)
		m.args = append(m.args, m.whereArgs...)
	}

	if len(m.primaryKeys) > 0 && m.adapter == "postgres" {
		m.queryBuilder.WriteString(" RETURNING ")
		m.queryBuilder.WriteString(strings.Join(m.primaryKeys, ", "))
	}

	m.queryBuilder.WriteString(";")

	return m
}

// Where indicates the condition of which records to return.
func (m *Model) Where(condition string, args ...interface{}) *Model {
	m.where, m.whereArgs = m.rebind(condition, args...)

	return m
}

func (m *Model) buildWhereWithPrimaryKeys() {
	if len(m.primaryKeys) < 1 {
		return
	}

	var builder strings.Builder
	args := []interface{}{}
	dest := reflect.ValueOf(m.dest)

	switch m.destKind {
	case reflect.Array, reflect.Slice:
		if len(m.primaryKeys) > 1 {
			builder.WriteString("(")
		}

		builder.WriteString(strings.Join(m.primaryKeys, ", "))

		if len(m.primaryKeys) > 1 {
			builder.WriteString(")")
		}

		builder.WriteString(" IN (")

		dest = dest.Elem()
		for i := 0; i < dest.Len(); i++ {
			elem := dest.Index(i)
			pkValues := []interface{}{}

			for _, pk := range m.primaryKeys {
				pkValue := elem.FieldByName(m.attrs[pk].stFieldName)

				if !pkValue.IsZero() {
					pkValues = append(pkValues, pkValue.Interface())
				}
			}

			if len(pkValues) == len(m.primaryKeys) {
				if len(m.primaryKeys) > 1 {
					builder.WriteString("(")
				}

				builder.WriteString(strings.Trim(strings.Repeat("?, ", len(m.primaryKeys)), ", "))

				if len(m.primaryKeys) > 1 {
					builder.WriteString(")")
				}

				args = append(args, pkValues...)
			}

			if i < dest.Len()-1 {
				builder.WriteString(", ")
			}
		}

		builder.WriteString(")")

		m.where = builder.String()
		m.whereArgs = args
	case reflect.Ptr:
		dest = dest.Elem()
		wheres := []string{}

		for _, pk := range m.primaryKeys {
			if !dest.FieldByName(m.attrs[pk].stFieldName).IsZero() {
				wheres = append(wheres, pk+" = ?")
				args = append(args, dest.FieldByName(m.attrs[pk].stFieldName).Interface())
			}
		}

		builder.WriteString(strings.Join(wheres, " AND "))
		m.where, m.whereArgs = m.rebind(builder.String(), args...)
	}
}

func (m *Model) rebind(query string, args ...interface{}) (string, []interface{}) {
	var builder strings.Builder
	newArgs := []interface{}{}
	count := 0

	for _, char := range query {
		var kind reflect.Kind

		if count < len(args) {
			kind = reflect.TypeOf(args[count]).Kind()
		}

		if char == '?' {
			if kind == reflect.Array || kind == reflect.Slice {
				arrayArg := reflect.ValueOf(args[count])
				builder.WriteString(strings.Trim(strings.Repeat("?, ", arrayArg.Len()), ", "))

				for i := 0; i < arrayArg.Len(); i++ {
					newArgs = append(newArgs, arrayArg.Index(i).Interface())
				}
			} else {
				builder.WriteString(string(char))

				if len(args) > 0 && args[count] != nil {
					newArgs = append(newArgs, args[count])
				}
			}

			count++
			continue
		}

		builder.WriteString(string(char))
	}

	return builder.String(), newArgs
}

func (m *Model) resetDest() {
	switch m.destKind {
	case reflect.Array, reflect.Slice:
		v := reflect.ValueOf(m.dest)
		v.Elem().Set(reflect.MakeSlice(v.Type().Elem(), 0, v.Elem().Cap()))
	case reflect.Ptr:
		v := reflect.ValueOf(m.dest)
		v.Elem().Set(reflect.New(v.Type().Elem()).Elem())
	}
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
