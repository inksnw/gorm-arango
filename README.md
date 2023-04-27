# gorm-arango

how to use

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	arango "github.com/inksnw/gorm-arango"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"time"
)

type Resource struct {
	Cluster   string
	Namespace string
	Name      string
	CreatedAt time.Time
	Object    json.RawMessage
}

func main() {
	arangoConf := &arango.Config{
		URI:      "http://192.168.50.54:8529",
		Database: "clusterpedia",
		Timeout:  5,
	}
	dialector := arango.Open(arangoConf)
	customLogger := logger.Default.LogMode(logger.Info)
	db, err := gorm.Open(dialector, &gorm.Config{Logger: customLogger})
	if err != nil {
		panic(err)
	}
	query := db.WithContext(context.TODO()).Model(&Resource{}).Select([]string{"name", "namespace"})
	var resources []json.RawMessage

	query.Where(map[string]any{
		"cluster":   "cluster-example",
		"namespace": "clusterpedia-system",
	}).Where("kind = ?", "Pod").Find(&resources)
	for _, i := range resources {
		fmt.Println(string(i))
	}
}
```