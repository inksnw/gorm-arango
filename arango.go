package gorm_arango

import (
	"context"
	"fmt"
	"github.com/cenkalti/backoff/v4"
	"reflect"
	"time"

	"github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
	"github.com/inksnw/gorm-arango/pkg/callbacks"
	"github.com/inksnw/gorm-arango/pkg/clause"
	"github.com/inksnw/gorm-arango/pkg/conn"
	"github.com/inksnw/gorm-arango/pkg/errors"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	gormClause "gorm.io/gorm/clause"
	"gorm.io/gorm/migrator"
	"gorm.io/gorm/schema"
)

const DriverName = "gorm-arango"

type Dialector struct {
	DriverName string
	Config     *Config
	Conn       gorm.ConnPool
	Connection driver.Connection
	Client     driver.Client
	Database   driver.Database
}

func Open(config *Config) gorm.Dialector {
	return &Dialector{Config: config}
}

func (d Dialector) Name() string {
	return "arango"
}

func (d Dialector) DatabaseExists(ctx context.Context, databaseName string) (exists bool, err error) {
	databases, err := d.Client.Databases(ctx)
	if err != nil {
		return false, err
	}
	for _, i := range databases {
		if i.Name() == databaseName {
			exists = true
		}
	}
	return exists, nil
}

func (d Dialector) CreateDatabaseIfNeeded(ctx context.Context, databaseName string) (database driver.Database, err error) {
	exists, err := d.DatabaseExists(ctx, databaseName)
	if exists {
		database, err = d.Client.Database(ctx, databaseName)
		return database, err
	}
	database, err = d.Client.CreateDatabase(ctx, databaseName, nil)
	return database, err
}

func (d Dialector) Initialize(db *gorm.DB) error {
	ctx, cancel := d.setupContext()
	defer cancel()
	d.DriverName = DriverName
	logEntry := logrus.WithFields(logrus.Fields{})
	logEntry.Debug("Connecting to ArangoDB server...")

	connection, err := http.NewConnection(http.ConnectionConfig{Endpoints: []string{d.Config.URI}})
	if err != nil {
		logEntry.WithError(err).Error("ArangoDB connection creation failed")
		return err
	}
	d.Connection = connection
	d.Client, err = driver.NewClient(driver.ClientConfig{
		Connection:     connection,
		Authentication: driver.BasicAuthentication(d.Config.User, d.Config.Password),
	})
	if err != nil {
		logEntry.WithError(err).Error("ArangoDB client creation failed")
		return err
	}

	expBackoff := backoff.WithMaxRetries(backoff.NewExponentialBackOff(), d.Config.MaxConnectionRetries)
	var database driver.Database
	operation := func() error {
		var err error
		database, err = d.CreateDatabaseIfNeeded(ctx, d.Config.Database)
		if err != nil {
			nextBackOff := expBackoff.NextBackOff()
			logEntry.WithError(err).Errorf("ArangoDB opening database connection failed. Retrying in %v...", nextBackOff)
			return errors.ErrOpeningDatabaseConnectionFailedWithRetry(fmt.Sprintf("Retrying in %v...", nextBackOff))
		}
		return err
	}
	err = backoff.Retry(operation, expBackoff)
	if err != nil {
		logEntry.WithError(err).Error("ArangoDB opening database connection failed")
		return errors.ErrOpeningDatabaseConnectionFailed
	}
	d.Database = database

	if d.Conn != nil {
		db.ConnPool = d.Conn
	} else {
		db.ConnPool = reflect.ValueOf(&conn.ConnPool{Connection: connection, Database: database}).Interface().(gorm.ConnPool)
	}

	clause.RegisterDefaultClauses(db)
	callbacks.RegisterDefaultCallbacks(db)
	db.Dialector = d

	return nil
}

func (d Dialector) Migrator(db *gorm.DB) gorm.Migrator {
	return Migrator{migrator.Migrator{Config: migrator.Config{
		DB:                          db,
		Dialector:                   d,
		CreateIndexAfterCreateTable: true,
	}}}
}

func (d Dialector) DataTypeOf(field *schema.Field) string {
	// TODO: Implement
	return string(field.DataType)
}

func (d Dialector) DefaultValueOf(field *schema.Field) gormClause.Expression {
	return gormClause.Expr{SQL: ""}
}

func (d Dialector) BindVarTo(writer gormClause.Writer, stmt *gorm.Statement, v any) {

	switch ins := v.(type) {
	case string:
		value := fmt.Sprintf("'%v'", v)
		writer.WriteString(value)
	case time.Time:
		value := fmt.Sprintf("'timestamp:%d'", ins.UnixMilli())
		writer.WriteString(value)
	case bool:
		value := fmt.Sprintf("%t", v)
		writer.WriteString(value)
	case *gorm.DB:
		stmt.Logger.Error(context.TODO(), "not support sub query")
		writer.WriteString("")

	default:
		value := fmt.Sprintf("%v", v)
		writer.WriteString(value)
	}
}

func (d Dialector) QuoteTo(writer gormClause.Writer, str string) {
	// TODO: Implement
	//panic("Implement me")
}

func (d Dialector) Explain(sql string, vars ...any) string {
	// TODO: Implement
	//panic("Implement me")
	return ""
}

func (d Dialector) setupContext() (context.Context, context.CancelFunc) {
	return context.WithDeadline(context.Background(), time.Now().Add(d.Config.Timeout*time.Second))
}
