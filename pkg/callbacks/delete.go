package callbacks

import (
	"context"
	"github.com/inksnw/gorm-arango/pkg/conn"
	"gorm.io/gorm"
)

func Delete(db *gorm.DB) {
	if db.Error != nil {
		return
	}
	connection := db.Statement.ConnPool.(*conn.ConnPool)
	collection, err := connection.Database.Collection(db.Statement.Context, db.Statement.Table)
	if err != nil {
		db.AddError(err)
		return
	}
	aql := BuildAQL(db)

	metas, err := conn.QueryAll(context.TODO(), connection, aql, "delete", db)
	_, _, err = collection.RemoveDocuments(db.Statement.Context, metas.Keys())
	if err != nil {
		db.AddError(err)
		return
	}
	db.RowsAffected = int64(len(metas))
}
