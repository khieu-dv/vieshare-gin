package db

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	"github.com/go-gorp/gorp"
	_redis "github.com/go-redis/redis/v7"
	_ "github.com/mattn/go-sqlite3" //import sqlite3
)

//DB ...
type DB struct {
	*sql.DB
}

var db *gorp.DbMap

//Init ...
func Init() {
	dbPath := getDBPath()
	
	// Ensure the database directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatal("Failed to create database directory:", err)
	}

	var err error
	db, err = ConnectDB(dbPath)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize database schema
	if err := initSchema(); err != nil {
		log.Fatal("Failed to initialize database schema:", err)
	}
}

//ConnectDB ...
func ConnectDB(dataSourceName string) (*gorp.DbMap, error) {
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.SqliteDialect{}}
	//dbmap.TraceOn("[gorp]", log.New(os.Stdout, "golang-gin:", log.Lmicroseconds)) //Trace database requests
	return dbmap, nil
}

//GetDB ...
func GetDB() *gorp.DbMap {
	return db
}

//RedisClient ...
var RedisClient *_redis.Client

//InitRedis ...
func InitRedis(selectDB ...int) {

	var redisHost = os.Getenv("REDIS_HOST")
	var redisPassword = os.Getenv("REDIS_PASSWORD")

	RedisClient = _redis.NewClient(&_redis.Options{
		Addr:     redisHost,
		Password: redisPassword,
		DB:       selectDB[0],
		// DialTimeout:        10 * time.Second,
		// ReadTimeout:        30 * time.Second,
		// WriteTimeout:       30 * time.Second,
		// PoolSize:           10,
		// PoolTimeout:        30 * time.Second,
		// IdleTimeout:        500 * time.Millisecond,
		// IdleCheckFrequency: 500 * time.Millisecond,
		// TLSConfig: &tls.Config{
		// 	InsecureSkipVerify: true,
		// },
	})

}

//GetRedis ...
func GetRedis() *_redis.Client {
	return RedisClient
}

// getDBPath returns the SQLite database file path
func getDBPath() string {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/app.db"
	}
	return dbPath
}

// initSchema initializes the database schema if it doesn't exist
func initSchema() error {
	// Check if PocketBase tables exist
	var count int
	err := db.Db.QueryRow("SELECT count(*) FROM sqlite_master WHERE type='table' AND name='users'").Scan(&count)
	if err != nil {
		return err
	}

	// If tables don't exist, create them using PocketBase schema
	if count == 0 {
		// Read and execute the PocketBase schema
		schemaPath := "./db/pocketbase_schema.sql"
		schemaBytes, err := os.ReadFile(schemaPath)
		if err != nil {
			return err
		}

		// Execute the schema
		if _, err := db.Db.Exec(string(schemaBytes)); err != nil {
			return err
		}

		log.Println("PocketBase-compatible database schema initialized successfully")
	}

	return nil
}
