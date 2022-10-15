package config

type Config struct {
	ServerAddress string `env:"RUN_ADDRESS"    envDefault:"localhost:8080"`
	DatabaseDSN   string `env:"DATABASE_URI"      envDefault:"postgres://matt:pvtjoker@localhost:5432/gophermart?sslmode=disable"`
}
