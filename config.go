package postscomments1

type Config struct {
	StorageType string
	Postgres    PostgresConfig
}

type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}
