package mdb

import (
	"database/sql"
	"log"
	"time"

	"github.com/mattn/go-sqlite3"
)

type EmailEntry struct {
	Id          int64
	Email       string
	ConfirmedAt *time.Time
	OptOut      bool
}

func TryCreate(db *sql.DB) {
	_, err := db.Exec(`
	CREATE TABLE emails (
		id          INTEGER PRIMARY KEY,
		email       TEXT UNIQUE,
		confirmedAt INTEGER,
		optOut      INTEGER
	);
	`)
	if err != nil {
		if sqlError, ok := err.(sqlite3.Error); ok {
			if sqlError.Code != 1 {
				log.Fatal(sqlError)
			}
		} else {
			log.Fatal(err)
		}
	}
}

func emailEntryFromRow(row *sql.Rows) (*EmailEntry, error) {
	var id int64
	var email string
	var confirmedAt int64
	var optOut bool

	err := row.Scan(&id, &email, &confirmedAt, &optOut)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	t := time.Unix(confirmedAt, 0)
	return &EmailEntry{Id: id, Email: email, ConfirmedAt: &t, OptOut: optOut}, nil
}

func CreateEmail(db *sql.DB, email string) error {
	_, err := db.Exec(`INSERT INTO emails(email, confirmedAt, optOut) VALUES(?, 0, 0)`, email)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func GetEmail(db *sql.DB, email string) (*EmailEntry, error) {
	rows, err := db.Query(`
	SELECT id, email, confirmedAt, optOut
	FROM emails
	WHERE email = ?
	`, email)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		return emailEntryFromRow(rows)
	}
	return nil, nil
}

func UpdateEmails(db *sql.DB, entry EmailEntry) error {
	t := entry.ConfirmedAt.Unix()
	_, err := db.Exec(`
		INSERT INTO emails(email, confirmedAt, optOut) VALUES(?, ?, ?)
		ON CONFLICT(email) DO UPDATE SET confirmedAt = ?, optOut = ?
	`, entry.Email, t, entry.OptOut, t, entry.OptOut)

	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func DeleteEmail(db *sql.DB, email string) error {
	_, err := db.Exec(`UPDATE emails SET optOut = 1 WHERE email = ?`, email)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

type GetEmailBatchQueryParams struct {
	Page  int
	Count int
}

func GetEmailBatch(db *sql.DB, params GetEmailBatchQueryParams) ([]EmailEntry, error) {
	rows, err := db.Query(`
		SELECT id, email, confirmedAt, optOut FROM emails
		WHERE optOut = 0 ORDER BY id ASC LIMIT ? OFFSET ?
	`, params.Count, (params.Page-1)*params.Count)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	emails := make([]EmailEntry, 0, params.Count)

	for rows.Next() {
		email, err := emailEntryFromRow(rows)
		if err != nil {
			return nil, err
		}
		emails = append(emails, *email)
	}
	return emails, nil
}