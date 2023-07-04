package internal

import (
	"encoding/json"
	"fmt"
	"github.com/Gatusko/webserver/structs"
	"log"
	"os"
	"sync"
)

type DB struct {
	path string
	mux  *sync.RWMutex
}

type DBStrcutre struct {
	Chirps map[int]structs.Chirpy `json:"chirps"`
}

func (dbStruc *DBStrcutre) NewMemory() {
	log.Printf("Creating new Memory")
	dbStruc.Chirps = make(map[int]structs.Chirpy)
}

const DBName = "myDb.json"

var LoadedDB DBStrcutre

func NewDB(path string) (*DB, error) {
	database := DB{}
	database.path = path
	LoadedDB.NewMemory()
	err := database.ensureDB()
	if err != nil {
		return nil, fmt.Errorf("We can't create the Databse : %s", err)
	}
	database.mux = &sync.RWMutex{}

	if err != nil {
		return &database, fmt.Errorf("We have an issue loading the memory %s", err)
	}
	return &database, nil
}

func (db *DB) CreateChirp(body string) (structs.Chirpy, error) {
	db.mux.Lock()
	defer db.mux.Unlock()
	newChirpy := structs.Chirpy{}
	newChirpy.Body = body
	newChirpy.Id = len(LoadedDB.Chirps) + 1
	LoadedDB.Chirps[len(LoadedDB.Chirps)+1] = newChirpy
	err := db.writeDB(LoadedDB)
	if err != nil {
		return newChirpy, fmt.Errorf("Issue creating chyrpy %s", err)
	}
	return newChirpy, nil
}

func (db *DB) GetChirps() ([]structs.Chirpy, error) {
	var allChirps []structs.Chirpy
	for _, value := range LoadedDB.Chirps {
		allChirps = append(allChirps, value)
	}
	return allChirps, nil
}

func (db *DB) writeDB(dbStruc DBStrcutre) error {
	dat, _ := json.Marshal(dbStruc)
	log.Printf("Printing to DB: %s", dat)
	err := os.WriteFile(db.path+"/"+DBName, (dat), 0666)
	if err != nil {
		return fmt.Errorf("Issue writing to db %s", err)
	}
	return nil
}

func (db *DB) loadDB(dbStruc DBStrcutre) (DBStrcutre, error) {
	finalPath := db.path + "/" + DBName
	data, err := os.ReadFile(finalPath)
	if err != nil {
		return DBStrcutre{}, fmt.Errorf("We got an error loading to the memory: %s", err)
	}
	err = json.Unmarshal(data, &dbStruc)
	if err != nil {
		return DBStrcutre{}, fmt.Errorf("We got an error loading to the memory: %s", err)
	}
	return dbStruc, nil
}

func (db *DB) ensureDB() error {
	finalPath := db.path + "/" + DBName
	_, err := os.ReadFile(finalPath)
	if err != nil {
		log.Printf("File doesn't exist, creating it")
		err = os.WriteFile(finalPath, []byte(""), 0666)
		if err != nil {
			return fmt.Errorf("File failed to create it %s", err)
		}
		return nil
	}
	log.Printf("I am here")
	LoadedDB, err = db.loadDB(LoadedDB)
	return nil
}
