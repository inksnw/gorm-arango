package callbacks

import (
	"encoding/json"
	"github.com/arangodb/go-driver"
	"reflect"
	"time"

	"github.com/inksnw/gorm-arango/pkg/conn"
	"gorm.io/gorm"
)

func Update(db *gorm.DB) {
	if db.Error != nil {
		return
	}
	connection := db.Statement.ConnPool.(*conn.ConnPool)
	collection, err := connection.Database.Collection(db.Statement.Context, db.Statement.Table)
	if err != nil {
		db.AddError(err)
		return
	}
	js, err := json.Marshal(db.Statement.Dest)
	if err != nil {
		db.AddError(err)
		return
	}
	data := make(map[string]any)
	err = json.Unmarshal(js, &data)
	if err != nil {
		db.AddError(err)
		return
	}

	data = keysToSnakeCase(data)
	data["updated_at"] = time.Now()
	delete(data, "id")
	delete(data, "created_at")
	delete(data, "deleted_at")

	document, err := MetaInfo(db, collection)
	if err != nil {
		db.AddError(err)
		return
	}

	if _, err := collection.UpdateDocument(db.Statement.Context, document.Key, data); err != nil {
		db.AddError(err)
	}
	db.RowsAffected = 1
}

func MetaInfo(db *gorm.DB, collection driver.Collection) (driver.DocumentMeta, error) {
	connection := db.Statement.ConnPool.(*conn.ConnPool)
	aql := BuildAQL(db)
	cursor, err := collection.Database().Query(db.Statement.Context, aql, nil)
	defer cursor.Close()
	if err != nil {
		return driver.DocumentMeta{}, err
	}
	r := reflect.New(connection.Return.ElemType).Interface()
	document, err := cursor.ReadDocument(db.Statement.Context, r)
	if err != nil {
		return driver.DocumentMeta{}, err
	}
	return document, err
}
