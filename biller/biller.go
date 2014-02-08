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

type playerLookup struct {
	name     string
	billerId uint
}

type zoneAndPid struct {
	zone string
	pid  uint
}

type zonePids struct {
	lock   sync.RWMutex
	lookup map[zoneAndPid]playerLookup
}

var zonePidLookup = zonePids{lookup: map[zoneAndPid]playerLookup{}}

type Player struct {
	name string
	zone string
}

type Biller struct {
	db *sql.DB

	zonesLock sync.RWMutex
	zones     map[string]*zone.Zone

	playersLock sync.RWMutex
	players     map[string]*Player
}

func NewBiller(file string) *Biller {
	biller := Biller{players: map[string]*Player{}}

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

func (b *Biller) PidToBillerId(zone string, pid uint) (uint, bool) {
	zonepid := zoneAndPid{zone: zone, pid: pid}
	zonePidLookup.lock.RLock()
	defer zonePidLookup.lock.RUnlock()

	val, ok := zonePidLookup.lookup[zonepid]
	return val.billerId, ok
}

func (b *Biller) PidToName(zone string, pid uint) (string, bool) {
	zonepid := zoneAndPid{zone: zone, pid: pid}
	zonePidLookup.lock.RLock()
	defer zonePidLookup.lock.RUnlock()

	val, ok := zonePidLookup.lookup[zonepid]
	return val.name, ok
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

func (b *Biller) EnterArena(name string, zone string) bool {
	b.zonesLock.Lock()
	defer b.zonesLock.Unlock()

	_, ok := b.zones[zone]
	if !ok {
		log.Printf("EnterArena: Unknown zone: %s\n", zone)
		return false
	}

	b.playersLock.Lock()
	defer b.playersLock.Unlock()
	player, ok := b.players[name]
	if !ok {
		log.Printf("EnterArena: Unknown player: %s\n", zone)
		return false
	}

	player.zone = zone
	return true
}

func (b *Biller) LeaveArena(name string, zone string) bool {
	b.zonesLock.Lock()
	defer b.zonesLock.Unlock()

	_, ok := b.zones[zone]
	if !ok {
		log.Printf("(Biller::LeaveArena) unknown zone: %s\n", zone)
		return false
	}

	b.playersLock.Lock()
	defer b.playersLock.Unlock()
	player, ok := b.players[name]
	if !ok {
		log.Printf("(Biller::LeaveArena) unknown player: %s\n", name)
		return false
	}

	if player.zone != zone {
		log.Printf("(Biller::LeaveArena) player is not in zone: %s\n", zone)
		return false
	}

	// TODO: Update database with player
	player.zone = ""
	delete(b.players, name)

	return true
}

func (b *Biller) Login(name string, password string) (error, bool) {
	b.playersLock.Lock()
	defer b.playersLock.Unlock()

	if b.loggedIn(name) {
		return fmt.Errorf("'%s' already logged in", name), false
	}

	var pw string
	err := b.db.QueryRow(`SELECT "password" FROM "player" WHERE name = ?`, name).Scan(&pw)
	switch {
	case err == sql.ErrNoRows:
		return fmt.Errorf("'%s' not registered", name), true
	case err != nil:
		log.Printf("(Biller::Login) sql error: %s\n", err)
		return fmt.Errorf("unknown error"), false
	}

	if pw != password {
		return fmt.Errorf("wrong password for player '%s'", name), false
	}

	b.players[name] = &Player{name: name, zone: ""}

	return nil, false
}

func (b *Biller) loggedIn(name string) bool {
	_, ok := b.players[name]
	if !ok {
		return false
	}

	return true
}

func (b *Biller) LoggedIn(name string) bool {
	b.playersLock.RLock()
	defer b.playersLock.RUnlock()

	return b.loggedIn(name)
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
