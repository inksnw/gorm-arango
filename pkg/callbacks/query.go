package callbacks

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
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
	err = scan(db)
	if err != nil {
		db.AddError(err)
		return
	}
}

func BuildAQL(db *gorm.DB) string {

	db.Statement.Build("ORDER BY")
	sort := db.Statement.SQL.String()
	sort = strings.ReplaceAll(sort, "SORT .", "")
	db.Statement.SQL.Reset()
	firstPart := fmt.Sprintf("FOR doc IN %s  ", db.Statement.Table)
	db.Statement.SQL.WriteString(firstPart)
	db.Statement.Build("WHERE", "LIMIT")
	db.Statement.SQL.WriteString(sort)
	returnPart := selectColumn(db.Statement.Selects)
	db.Statement.SQL.WriteString(returnPart)
	db.Logger.Info(context.TODO(), db.Statement.SQL.String())

	return db.Statement.SQL.String()
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

// This method is based on gorm.Scan() method.
func scan(db *gorm.DB) error {
	isSlice := db.Statement.ConnPool.(*conn.ConnPool).Return.IsSlice
	if !isSlice {
		return nil
	}
	elemType := db.Statement.ConnPool.(*conn.ConnPool).Return.ElemType
	model := reflect.New(elemType).Interface()
	mapModel, err := any2Map(model)
	if err != nil {
		return err
	}

	list := db.Statement.ConnPool.(*conn.ConnPool).Return.Dest.([]any)
	for _, row := range list {
		db.RowsAffected++
		if len(mapModel) == 0 {
			elem := reflect.ValueOf(row)
			db.Statement.ReflectValue.Set(reflect.Append(db.Statement.ReflectValue, elem.Elem()))
			continue
		}
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
}
