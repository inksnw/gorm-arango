package callbacks

import (
	"encoding/json"
	"github.com/inksnw/gorm-arango/pkg/conn"
	"gorm.io/gorm"
	"strings"
	"time"
	"unicode"
)

func Create(db *gorm.DB) {
	if db.Error != nil {
		return
	}

	connection := db.Statement.ConnPool.(*conn.ConnPool)
	collection, err := connection.Database.Collection(db.Statement.Context, db.Statement.Table)
	if err != nil {
		db.AddError(err)
		return
	}
	now := time.Now()
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
	data["id"] = uint(now.UnixNano() / 1000)
	data["updated_at"] = now

	_, err = collection.CreateDocument(db.Statement.Context, data)
	if err != nil {
		db.AddError(err)
		return
	}
	db.RowsAffected = int64(1)
}

func CamelCaseToUnderscore(s string) string {
	if strings.Contains(s, "_") {
		return s
	}
	var output []rune
	for i, r := range s {
		if i != 0 && unicode.IsUpper(r) {
			if i+1 < len(s) && unicode.IsLower(rune(s[i+1])) || unicode.IsLower(rune(s[i-1])) {
				output = append(output, '_')
			}
		}
		output = append(output, unicode.ToLower(r))
	}
	return string(output)
}

func keysToSnakeCase(m map[string]any) map[string]any {
	newMap := make(map[string]any)
	for k, v := range m {
		snakeKey := CamelCaseToUnderscore(k)
		newMap[snakeKey] = v
	}
	return newMap
}
