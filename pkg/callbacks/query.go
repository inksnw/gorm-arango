package callbacks

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"reflect"

	"time"

	"github.com/inksnw/gorm-arango/pkg/conn"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func Query(db *gorm.DB) {
	aql := BuildAQL(db)
	_, err := db.Statement.ConnPool.QueryContext(db.Statement.Context, aql, db)
	if err != nil {
		db.AddError(err)
		return
	}
	isSlice := db.Statement.ReflectValue.Kind() == reflect.Slice || db.Statement.ReflectValue.Kind() == reflect.Array
	if isSlice {
		err = scan(db)
		if err != nil {
			db.AddError(err)
			return
		}
	}

}

func BuildAQL(db *gorm.DB) string {

	db.Statement.Build("WHERE")
	Where := db.Statement.SQL.String()
	db.Statement.SQL.Reset()
	db.Statement.Build("LIMIT")
	Limit := db.Statement.SQL.String()
	db.Statement.SQL.Reset()
	db.Statement.Build("ORDER BY")
	order := db.Statement.SQL.String()
	db.Statement.SQL.Reset()
	returnPart := selectColumn(db.Statement.Selects)

	firstPart := fmt.Sprintf("for doc in %s filter ", db.Statement.Table)
	all := fmt.Sprintf("%s %s %s %s %s", firstPart, Where, Limit, order, returnPart)
	db.Statement.SQL.WriteString(all)

	sql := db.Statement.SQL.String()
	db.Logger.Info(context.TODO(), sql)

	return sql
}

func selectColumn(in []string) (result string) {
	if len(in) == 0 {
		return fmt.Sprintf(" RETURN doc")
	}
	if len(in) == 1 {
		return fmt.Sprintf(" RETURN doc.%s", in[0])
	}

	for idx, column := range in {
		if idx != len(in)-1 {
			result = result + fmt.Sprintf("\"%s\":doc.%s,", column, column)
		} else {
			result = result + fmt.Sprintf("\"%s\":doc.%s", column, column)
		}
	}
	return " RETURN {" + result + "}"
}

func any2Map(in any) (map[string]any, error) {
	data, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}
	var mapModel map[string]any
	err = json.Unmarshal(data, &mapModel)
	return mapModel, err
}
func checkElementType(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Struct, reflect.Map:
		return false
	default:
		return true
	}
}

// This method is based on gorm.Scan() method.
func scan(db *gorm.DB) error {

	elemType := conn.NewInstanceOfSliceType(db.Statement.Dest)
	list := db.Statement.Dest.([]any)

	switch elemType.Kind() {
	case reflect.Struct, reflect.Map:
		for _, row := range list {
			data, err := any2Map(row)
			if err != nil {
				return err
			}
			reflectValueType := db.Statement.ReflectValue.Type().Elem()
			elem := reflect.New(reflectValueType)
			for _, field := range db.Statement.Schema.Fields {
				value, err := json.Marshal(data[field.Name])
				if err != nil {
					return err
				}
				var v any
				if field.DataType == schema.Time {
					v = time.Unix(int64(binary.BigEndian.Uint64(value)), 0)
				} else {
					v = value
				}
				err = field.Set(context.TODO(), elem, v)
				if err != nil {
					return err
				}
			}
			db.Statement.ReflectValue.Set(reflect.Append(db.Statement.ReflectValue, elem.Elem()))
		}
		db.RowsAffected = int64(len(list))
		return nil
	default:
		for _, row := range list {
			elem := reflect.ValueOf(row)
			db.Statement.ReflectValue.Set(reflect.Append(db.Statement.ReflectValue, elem.Elem()))
		}
		db.RowsAffected = int64(len(list))
		return nil
	}

}
