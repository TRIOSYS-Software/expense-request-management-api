package configs

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func (c *Config) ConnectDB() error {
	dsn := c.DBUser + ":" + c.DBPassword + "@tcp(" + c.DBHost + ":" + c.DBPort + ")/" + c.DBName + "?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	c.DB = db
	return nil
}
