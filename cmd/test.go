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
	query := db.WithContext(context.TODO()).Model(&Resource{}).Select("object")
	var resources1 [][]byte
	var resources2 []json.RawMessage
	var resources3 BytesList
	var resources4 []Resource
	var resources5 Resource

	where := map[string]any{"cluster": "cluster-example", "namespace": "clusterpedia-system"}
	query.Where(where).Where("kind = ?", "Pod").Find(&resources1)
	query.Where(where).Where("kind = ?", "Pod").Find(&resources2)
	query.Where(where).Where("kind = ?", "Pod").Find(&resources3)
	db.WithContext(context.TODO()).Model(&Resource{}).Where(where).Where("kind = ?", "Pod").Find(&resources4)
	db.WithContext(context.TODO()).Model(&Resource{}).Where(where).Where("kind1 = ? and kind2 = ? and kind3 = ?", 1, "Pod2", "Pod3").Find(&resources5)
	fmt.Println(len(resources1))
	fmt.Println(len(resources2))
	fmt.Println(len(resources3))
	fmt.Println(len(resources4))
	fmt.Println(resources5)

}
