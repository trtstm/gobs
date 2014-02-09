package biller

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"gobs/zone"
	"log"
	"os"
	"sync"
)

type Player struct {
	name     string
	zone     string
	billerId uint
	pid      uint
	inside   bool
}

type Biller struct {
	db *sql.DB

	zonesLock sync.RWMutex
	zones     map[string]*zone.Zone

	playersLock sync.RWMutex
	players     map[uint]*Player
}

func NewBiller(file string) *Biller {
	biller := Biller{players: map[uint]*Player{}, zones: map[string]*zone.Zone{}}

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

func (b *Biller) PidToBillerId(zoneName string, pid uint) (uint, bool) {
	b.zonesLock.RLock()
	defer b.zonesLock.RUnlock()

	zone, ok := b.zones[zoneName]
	if !ok {
		log.Printf("YEEEEEEY")
		return 0, false
	}

	billerId, ok := zone.ToBillerId(pid)
	return billerId, ok
}

func (b *Biller) PidToName(zoneName string, pid uint) (string, bool) {
	b.zonesLock.RLock()
	defer b.zonesLock.RUnlock()

	zone, ok := b.zones[zoneName]
	if !ok {
		return "", false
	}

	billerId, ok := zone.ToBillerId(pid)
	if !ok {
		return "", false
	}

	b.playersLock.RLock()
	defer b.playersLock.RUnlock()
	player, ok := b.players[billerId]
	if !ok {
		return "", false
	}

	return player.name, true
}

func (b *Biller) CreateZone(name string) {
	b.zonesLock.Lock()
	defer b.zonesLock.Unlock()

	_, ok := b.zones[name]
	if ok {
		log.Printf("CreateZone: Zone already exists: %s\n", name)
		return
	}

	tmp := zone.NewZone(name)
	if tmp == nil {
		log.Printf("CreateZone: Out of memory?: %s\n", name)
		return
	}

	b.zones[name] = tmp
}

func (b *Biller) EnterArena(billerId uint) {
	b.playersLock.Lock()
	defer b.playersLock.Unlock()
	player, ok := b.players[billerId]
	if !ok {
		log.Printf("(Biller::EnterArena) unknown player: %d\n", billerId)
		return
	}

	// TODO: Start usage counter
	player.inside = true
}

func (b *Biller) LeaveArena(billerId uint) {
	b.playersLock.Lock()
	defer b.playersLock.Unlock()
	player, ok := b.players[billerId]
	if !ok {
		log.Printf("(Biller::LeaveArena) unknown player: %d\n", billerId)
		return
	}

	b.zonesLock.RLock()
	defer b.zonesLock.RUnlock()

	zone, ok := b.zones[player.zone]
	// Should never happen
	if !ok {
		log.Printf("(Biller::LeaveArena) unknown zone: %s\n", player.zone)
		return
	}

	// TODO: Update database with player
	zone.RemovePlayer(player.pid)
	delete(b.players, billerId)

	return true
}

func (b *Biller) Login(name string, password string, zoneName string, pid uint) (error, bool) {
	b.playersLock.Lock()
	defer b.playersLock.Unlock()

	b.zonesLock.RLock()
	defer b.zonesLock.RUnlock()

	// Should never happen since zone is added when it connects.
	zone, ok := b.zones[zoneName]
	if !ok {
		return fmt.Errorf("zone not found"), false
	}

	var pw string
	var billerId uint
	err := b.db.QueryRow(`SELECT "password", "billerid" FROM "player" WHERE name = ?`, name).Scan(&pw, &billerId)
	switch {
	case err == sql.ErrNoRows:
		return fmt.Errorf("'%s' not registered", name), true
	case err != nil:
		log.Printf("(Biller::Login) sql error: %s\n", err)
		return fmt.Errorf("unknown error"), false
	}

	if b.loggedIn(billerId) {
		return fmt.Errorf("'%s' already logged in", name), false
	}

	if pw != password {
		return fmt.Errorf("wrong password for player '%s'", name), false
	}

	player := Player{name: name, zone: zoneName, billerId: billerId, pid: pid, inside: false}
	b.players[billerId] = &player
	zone.AddPlayer(player.pid, player.billerId)

	return nil, false
}

func (b *Biller) loggedIn(billerId uint) bool {
	_, ok := b.players[billerId]
	if !ok {
		return false
	}

	return true
}

func (b *Biller) LoggedIn(billerId uint) bool {
	b.playersLock.RLock()
	defer b.playersLock.RUnlock()

	return b.loggedIn(billerId)
}

func (b *Biller) UserExists(name string) bool {
	var billerid uint
	err := b.db.QueryRow(`SELECT "billerid" FROM "player" WHERE name = ?`, name).Scan(&billerid)
	switch {
	case err == sql.ErrNoRows:
		return false
	case err != nil:
		log.Printf("(Biller::Login) sql error: %s\n", err)
		return false
	}

	return true
}

func (b *Biller) CreateUser(name string, password string) error {
	// TODO: Check if name and pw contains :

	stmt, err := b.db.Prepare(`INSERT INTO "player" ("name","password") VALUES (?, ?)`)
	if err != nil {
		log.Printf("(Biller::CreateUser) could not prepare statement: %s\n", err)
		return err
	}
	defer stmt.Close()

	res, err := stmt.Exec(name, password)
	if err != nil {
		log.Printf("(Biller::CreateUser) could not execute query: %s\n", err)
		return err
	}

	if n, err := res.RowsAffected(); err != nil || n != 1 {
		log.Printf("(Biller::CreateUser) user was not inserted: %s\n", err)
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
