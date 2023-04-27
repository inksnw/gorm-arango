package gorm_arango

import (
	driver "github.com/arangodb/go-driver"
	"github.com/inksnw/gorm-arango/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/migrator"
)

type Migrator struct {
	migrator.Migrator
}

func (m Migrator) AutoMigrate(values ...any) error {
	for _, value := range m.ReorderModels(values, true) {
		tx := m.DB.Session(&gorm.Session{})
		if !tx.Migrator().HasTable(value) {
			if err := tx.Migrator().CreateTable(value); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m Migrator) CurrentDatabase() (name string) {
	if dialector, ok := m.DB.Dialector.(Dialector); ok {
		name = dialector.Database.Name()
	}
	return
}

func (m Migrator) CreateTable(values ...any) error {
	return m.RunWithValue(values[0], func(stmt *gorm.Statement) error {
		if dialector, ok := m.DB.Dialector.(Dialector); ok {
			if dialector.Database == nil {
				return errors.ErrDatabaseConnectionFailed
			}
			_, err := dialector.Database.CreateCollection(stmt.Context, stmt.Table, &driver.CreateCollectionOptions{})
			return err
		}
		return errors.ErrDatabaseConnectionFailed
	})
}

func (m Migrator) DropTable(values ...any) error {
	values = m.ReorderModels(values, false)
	for i := len(values) - 1; i >= 0; i-- {
		if err := m.RunWithValue(values[i], func(stmt *gorm.Statement) error {
			if dialector, ok := m.DB.Dialector.(Dialector); ok {
				if hasTable := m.HasTable(stmt.Table); hasTable {
					collection, err := dialector.Database.Collection(stmt.Context, stmt.Table)
					if err != nil {
						return err
					}
					return collection.Remove(stmt.Context)
				}
				return nil
			}
			return errors.ErrDatabaseConnectionFailed
		}); err != nil {
			return err
		}
	}
	return nil
}

func (m Migrator) HasTable(value any) bool {
	var hasTable bool
	var err error

	err = m.RunWithValue(value, func(stmt *gorm.Statement) error {
		if dialector, ok := m.DB.Dialector.(Dialector); ok {
			hasTable, err = dialector.Database.CollectionExists(stmt.Context, stmt.Table)
			return err
		}
		return errors.ErrDatabaseConnectionFailed
	})
	if err != nil {
		panic(err)
	}

	return hasTable
}
