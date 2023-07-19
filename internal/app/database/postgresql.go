package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

var db *sql.DB

func InitPostgresql(host, user, password, dbname string, port int) {
	var err error
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+"password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}

	if err := db.Ping(); err != nil {
		panic(err)
	}

	db.SetConnMaxLifetime(time.Hour)
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(200)
	Migration(db)

}

func GetConnectionPool() *sql.DB {
	if db == nil {
		InitPostgresql("localhost", "postgres", "Cuongnguyen2001", "IT4409", 5432)
	}
	return db
}

func Migration(db *sql.DB) {
	db.ExecContext(context.Background(), `CREATE TABLE "blogs" (
		"id" varchar(50) PRIMARY KEY,
		"user_id" varchar(50),
		"title" text NOT NULL,
		"category" text,
		"content" text NOT NULL,
		"picture" text,
		"time_created" timestamp,
		"last_updated" timestamp
	  );`)

	db.ExecContext(context.Background(), `CREATE TABLE "users" (
		"id" varchar(50) PRIMARY KEY,
		"name" varchar,
		"email" varchar UNIQUE NOT NULL,
		"role" varchar,
		"provider" varchar,
		"picture" text,
		"time_created" timestamp,
		"last_updated" timestamp
	  );`)

	db.ExecContext(context.Background(), `CREATE TABLE "comments" (
		"id" varchar(50) PRIMARY KEY,
		"blog_id" varchar(50),
		"user_id" varchar(50),
		"parent_id" varchar(50),
		"content" text NOT NULL,
		"time_created" timestamp,
		"last_updated" timestamp
	  );`)
	db.ExecContext(context.Background(), `CREATE TABLE "permissions" (
		"id" serial  PRIMARY KEY,
		"user_id" varchar(50),
		"resource_id" varchar(50),
		"action" varchar
	  );
	`)

	db.ExecContext(context.Background(), `ALTER TABLE "blogs" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");

	ALTER TABLE "comments" ADD FOREIGN KEY ("blog_id") REFERENCES "blogs" ("id");
	
	ALTER TABLE "comments" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");
	
	ALTER TABLE "comments" ADD FOREIGN KEY ("parent_id") REFERENCES "comments" ("id");
	
	ALTER TABLE "permissions" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");`)
}
