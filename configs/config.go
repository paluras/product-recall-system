package configs

import (
	"flag"
	"fmt"
)

type Config struct {
	DBUser     string
	DBPassword string
	DBHost     string
	DBPort     string
	DBName     string
}

func ParseFlags() *Config {
	conf := &Config{}

	flag.StringVar(&conf.DBUser, "dbuser", "", "Database user")
	flag.StringVar(&conf.DBPassword, "dbpass", "", "Database password")
	flag.StringVar(&conf.DBHost, "dbhost", "localhost", "Database host")
	flag.StringVar(&conf.DBPort, "dbport", "3307", "Database port")
	flag.StringVar(&conf.DBName, "dbname", "scraper_db", "Database name")

	flag.Parse()
	return conf
}

func (c *Config) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
}
