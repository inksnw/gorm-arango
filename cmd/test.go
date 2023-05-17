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
type Bytes json.RawMessage
type BytesList []Bytes

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
	var resources4 []Resource
	var resources5 []Resource

	where := map[string]any{"cluster": "cluster-example", "namespace": "clusterpedia-system"}
	db.WithContext(context.TODO()).Model(&Resource{}).Where(where).Where("kind = ? and kind1 = ?", "Pod", true).Find(&resources4)
	db.WithContext(context.TODO()).Model(&Resource{}).Where(where).Where("name IN ?", []string{"clusterpedia-controller-manager", "jinzhu 2"}).Find(&resources5)
	fmt.Println(len(resources4))
	fmt.Println(len(resources5))

}
