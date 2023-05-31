package conn

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"gorm.io/gorm"
	"reflect"
	//"sync"

	driver "github.com/arangodb/go-driver"
)

type ConnPool struct {
	Connection driver.Connection
	Database   driver.Database
}

func (connPool *ConnPool) GetDBConn() (*sql.DB, error) {
	//TODO implement me
	return &sql.DB{}, nil
}

func (connPool *ConnPool) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	// TODO: Implements
	return nil, nil
}

func (connPool *ConnPool) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	// TODO: Implements
	return nil, nil
}

func (connPool *ConnPool) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	db := args[0].(*gorm.DB)
	_, err := QueryAll(ctx, query, "query", db)
	return nil, err
}

func (connPool *ConnPool) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	// TODO: Implements
	return nil
}

func CheckRaw(in any) (result bool) {
	value := reflect.ValueOf(in)
	elemType := value.Type().Elem()
	if elemType.Kind() != reflect.Slice {
		return false
	}
	elemElemType := elemType.Elem()
	return elemElemType.Kind() == reflect.Uint8
}

func QueryAll(ctx context.Context, query string, action string, db *gorm.DB) (metaSlice driver.DocumentMetaSlice, err error) {
	connPool := db.Statement.ConnPool.(*ConnPool)

	cursor, err := connPool.Database.Query(ctx, query, nil)
	if err != nil {
		return nil, err
	}
	defer cursor.Close()
	isSlice := db.Statement.ReflectValue.Kind() == reflect.Slice || db.Statement.ReflectValue.Kind() == reflect.Array

	if !isSlice && action != "delete" {
		meta, err := cursor.ReadDocument(ctx, db.Statement.Dest)
		if driver.IsNoMoreDocuments(err) {
			return nil, errors.New("document not found")
		}
		metaSlice = append(metaSlice, meta)
		return metaSlice, nil
	}

	results := make([]any, 0)
	for {
		elemType := NewInstanceOfSliceType(db.Statement.Dest)
		r := reflect.New(elemType).Interface()
		var meta driver.DocumentMeta
		if CheckRaw(r) {
			rawMessage := json.RawMessage{}
			meta, err = cursor.ReadDocument(ctx, &rawMessage)
			if driver.IsNoMoreDocuments(err) {
				break
			}
			if err != nil {
				return nil, err
			}
			rVal := reflect.ValueOf(r).Elem()
			rawMessageVal := reflect.ValueOf(rawMessage)
			convertedVal := rawMessageVal.Convert(rVal.Type())
			rVal.Set(convertedVal)

		} else {
			meta, err = cursor.ReadDocument(ctx, r)
			if driver.IsNoMoreDocuments(err) {
				break
			}
			if err != nil {
				return nil, err
			}
		}

		results = append(results, r)
		metaSlice = append(metaSlice, meta)
	}

	db.Statement.Dest = results
	db.RowsAffected = int64(len(results))

	return metaSlice, nil
}

func NewInstanceOfSliceType(arr any) reflect.Type {
	val := reflect.ValueOf(arr)
	if val.Kind() == reflect.Ptr {
		return NewInstanceOfSliceType(val.Elem().Interface())
	}
	if val.Kind() == reflect.Struct {
		return reflect.TypeOf(arr)
	}
	return reflect.TypeOf(arr).Elem()
}
