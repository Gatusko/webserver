package internal

import (
	"os"
	"testing"
	"time"
)

func TestNewDB(t *testing.T) {
	_, err := NewDB(".")
	if err != nil {
		t.Fatalf("File need to be created %s", err)
	}
	_, err = os.ReadFile("./myDb.json")
	if err != nil {
		t.Fatalf("File should exist: %s", err)
	}
}

func TestDB_CreateChirp(t *testing.T) {
	err := os.Remove("./myDb.json")
	myDB, err := NewDB(".")
	if err != nil {
		t.Fatalf("File need to be created %s", err)
	}
	_, err = myDB.CreateChirp("Test Test Test")
	t.Log(LoadedDB.Chirps)
	if err != nil {
		t.Fatalf("It should create and save to chirpy %s", err)
	}

}

func TestDB_CreateChirpMultiThread(t *testing.T) {
	myDB, err := NewDB(".")
	if err != nil {
		t.Fatalf("File need to be created %s", err)
	}
	go myDB.CreateChirp("Test Test Test")
	go myDB.CreateChirp("Test Test Test2")
	go myDB.CreateChirp("Test Test Test3")
	go myDB.CreateChirp("Test Test Test")
	go myDB.CreateChirp("Test Test Test")
	go myDB.CreateChirp("Test Test Test")
	go myDB.CreateChirp("Test Test Test")
	go myDB.CreateChirp("Test Test Test")
	time.Sleep(500 * time.Millisecond)
}
