package main

import (
	"database/sql/driver"
	"fmt"

	_ "github.com/jinzhu/gorm/dialects/postgres"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DbType string

const (
	PostgreSQL DbType = "PostgreSQL"
	Oracle     DbType = "Oracle"
	Ignite     DbType = "Ignite"
)

func (st *DbType) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		*st = DbType(b)
	}
	return nil
}

func (st DbType) Value() (driver.Value, error) {
	return string(st), nil
}

func (st DbType) TableName() string {
	return "admin.db_type"
}

type MonitoringDatabase struct {
	gorm.Model
	ID     uint
	Name   string
	DBType DbType `gorm:"type:admin.db_type;default:'PostgreSQL'"`
}

func main() {
	dsn := `postgresql://pcbo:pcbo@192.9.200.43:5432/powercard?connect_timeout=10&application_name=Powercard`

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		// fmt.Println(err)
		panic(err)
	}

	db.Debug().Exec(`
	DO $$ BEGIN
		CREATE TYPE admin.db_type AS ENUM ('PostgreSQL','Oracle','Ignite');
	EXCEPTION
		WHEN duplicate_object THEN null;
	END $$;`)

	md := MonitoringDatabase{}
	db.AutoMigrate(&md)
	md1 := MonitoringDatabase{
		DBType: PostgreSQL,
	}
	db.Debug().Create(&md1)

	fmt.Println(md1)
}
