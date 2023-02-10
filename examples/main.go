package main

import (
	"log"

	"github.com/bendersilver/sqleasy"
)

func main() {
	db, err := sqleasy.New()
	if err != nil {
		panic(err)
	}

	defer db.Close()
	if err = fn(db); err != nil {
		log.Fatal(err)
	}

}

func fn(db *sqleasy.Conn) error {

	err := db.Exec("create table foo (id integer not null primary key, name text);")
	if err != nil {
		return err
	}

	err = db.Exec("insert into foo values($1, 'www'), ($2, 'ads');", int64(1), int64(34))
	if err != nil {
		return err
	}

	rows, err := db.Query("select * from foo order by id;")
	if err != nil {
		return err
	}
	for rows.Next() {
		v, err := rows.Values()
		if err != nil {
			break
		}
		log.Println(v)
	}
	return rows.Err()

}
