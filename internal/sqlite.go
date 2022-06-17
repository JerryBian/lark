package internal

import (
	"database/sql"
	"log"
	"os"

	C "github.com/JerryBian/lark/internal/config"
	_ "github.com/mattn/go-sqlite3"
)

type Sqlite struct {
	Conf *C.Config
}

func init() {
	log.SetOutput(os.Stdout)
}

func (s *Sqlite) Startup() {
	database, err := sql.Open("sqlite3", s.Conf.Database.ConnStr)
	if err != nil {
		panic(err)
	}
	defer database.Close()

	log.Println("Running startup.sql...")
	sql, err := s.Conf.Runtime.F.ReadFile("internal/startup.sql")
	if err != nil {
		panic(err)
	}

	_, err = database.Exec(string(sql))

	if err != nil {
		panic(err)
	}
	log.Println("Ran startup.sql completed.")
}