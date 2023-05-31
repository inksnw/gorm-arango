package main

import (
	"context"
	"encoding/json"
	"fmt"
	arango "github.com/inksnw/gorm-arango"
	"gorm.io/datatypes"
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
	var resources4 []Resource
	var resources5 []string
	var resources6 []json.RawMessage
	type Bytes datatypes.JSON
	type BytesList []Bytes

	resources7 := &BytesList{}

	where := map[string]any{"cluster": "cluster-example", "namespace": "clusterpedia-system"}
	db.WithContext(context.TODO()).Model(&Resource{}).Select("uid").Where(where).Where("kind = ? ", "Pod").Find(&resources5)
	for _, i := range resources5 {
		fmt.Println(i)
	}
	db.WithContext(context.TODO()).Model(&Resource{}).Where(where).Where("kind = ? ", "Pod").Find(&resources4)
	for _, i := range resources5 {
		fmt.Println(i)
	}
	db.WithContext(context.TODO()).Model(&Resource{}).Select("object").Where(where).Where("kind = ? ", "Pod").Find(&resources6)
	fmt.Println(len(resources6))
	db.WithContext(context.TODO()).Model(&Resource{}).Select("object").Where(where).Where("kind = ? ", "Pod").Find(&resources7)

}
