package callbacks

import (
	"gorm.io/gorm"
)

func RegisterDefaultCallbacks(db *gorm.DB) {
	db.Callback().Create().Register("arango:create", Create)
	db.Callback().Query().Register("arango:query", Query)
	db.Callback().Update().Register("arango:update", Update)
	db.Callback().Delete().Register("arango:delete", Delete)
}
