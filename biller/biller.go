package biller

import (
	"log"
	"database/sql"
	"os"
	_ "github.com/mattn/go-sqlite3"
)

type Biller struct {
	db *sql.DB
}

func NewBiller(file string) *Biller {
	biller := Biller{}

	if fileExists(file) {
		var err error
		biller.db, err = sql.Open("sqlite3", file)
		if err != nil {
			log.Printf("Could not open database: %s\n", err)
			return nil
		}

		// Load from database
	} else {
		err := biller.createDatabase(file)
		if err != nil {
			log.Printf("Could not create database: %s\n", err)
			return nil
		}
	}
	

	return &biller
}

func (b *Biller) UserExists(name string) bool {
	var billerid uint
	err := b.db.QueryRow(`SELECT "billerid" FROM "player" WHERE name = ?`, name).Scan(&billerid)
	if err != nil {
		return false
	}

	return true
}

func (b *Biller) CreateUser(name string, password string) error {
	// TODO: Check if name and pw contains :

	stmt, err := b.db.Prepare(`INSERT INTO "player" ("name","password") VALUES (?, ?)`)
	if err != nil {
		//log.Printf("CreateUser: Could not prepare statement: %s\n", err)
		return err
	}
	defer stmt.Close()

	res, err := stmt.Exec(name, password)
	if err != nil {
		//log.Printf("CreateUser: Could not execute query: %s\n", err)
		return err
	}

	if n, err := res.RowsAffected(); err != nil || n != 1 {
		//log.Printf("CreateUser: User was not inserted: %s\n", err)
		return err
	}

	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return false
	}

	return true	
}

func (b *Biller) createDatabase(file string) error {
	var err error
	b.db, err = sql.Open("sqlite3", file)
	if err != nil {
		return err
	}

	sql := `
CREATE TABLE "player" ("billerid" INTEGER PRIMARY KEY AUTOINCREMENT  NOT NULL  UNIQUE , "name" VARCHAR NOT NULL  UNIQUE , "password" VARCHAR NOT NULL , "squad" INTEGER, "usage" INTEGER DEFAULT 0, "firstused" DATETIME DEFAULT CURRENT_DATE)
	`

	_, err = b.db.Exec(sql)
	if err != nil {
		return err
	}

	return nil
}
