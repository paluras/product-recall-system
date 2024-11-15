package configs

import "flag"

type Config struct {
	DBUser     string
	DBPassword string
	DBHost     string
	DBName     string
}

func ParseFlags() *Config {
	conf := &Config{}

	flag.StringVar(&conf.DBUser, "dbuser", "", "Database user")
	flag.StringVar(&conf.DBPassword, "dbpass", "", "Database password")
	flag.StringVar(&conf.DBHost, "dbhost", "localhost:3306", "Database host")
	flag.StringVar(&conf.DBName, "dbname", "scraper_db", "Database name")

	flag.Parse()
	return conf
}
