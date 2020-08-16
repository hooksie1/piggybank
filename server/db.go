package server

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"go.etcd.io/bbolt"
)

type Recorder interface {
	AddRecord() error
	GetRecord() error
	DeleteRecord() error
}

func WriteRecord(r Recorder) error {
	if err := r.AddRecord(); err != nil {
		return err
	}

	return nil
}

func ReadRecord(r Recorder) error {
	if err := r.GetRecord(); err != nil {
		return err
	}

	return nil
}

func RemoveRecord(r Recorder) error {
	if err := r.DeleteRecord(); err != nil {
		return err
	}

	return nil
}

func BackupLocal() error {
	db, err := bbolt.Open(dbName, 0600, nil)
	if err != nil {
		return fmt.Errorf("error opening database: %s", err)
	}

	defer db.Close()

	err = db.View(func(tx *bbolt.Tx) error {
		f, err := os.OpenFile("./backup.db", os.O_RDWR|os.O_CREATE, 0600)
		if err != nil {
			return err
		}
		writer := bufio.NewWriter(f)
		_, err = tx.WriteTo(writer)
		if err != nil {
			return err
		}

		return nil
	})

	return nil
}

func BackupHTTP(w http.ResponseWriter, r *http.Request) {
	db, err := bbolt.Open(dbName, 0600, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("error opening database: %s", err)
	}

	defer db.Close()

	err = db.View(func(tx *bbolt.Tx) error {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", `attachment; filename="piggy-backup.db"`)
		w.Header().Set("Content-Length", strconv.Itoa(int(tx.Size())))
		_, err := tx.WriteTo(w)
		return err
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
