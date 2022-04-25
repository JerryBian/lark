package main

import (
	"database/sql"
	"embed"
	"errors"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

//go:embed img/* style.css html/* bootstrap.min.js bootstrap.min.css
var f embed.FS
var dbLocation string

func main() {
	log.SetOutput(os.Stdout)

	var setupResult = setup()
	if setupResult == false {
		os.Exit(1)
	}

	r := gin.Default()

	templ := template.Must(template.New("").ParseFS(f, "html/*"))
	r.SetHTMLTemplate(templ)

	r.StaticFS("/asset", http.FS(f))

	r.GET("/", func(c *gin.Context) {
		words, err := GetAll()
		if err != nil {
			log.Fatal(err)
			c.String(http.StatusBadRequest, "Something is wrong.")
		}

		c.HTML(http.StatusOK, "index.html", gin.H{
			"Words": words,
		})
	})

	r.POST("/api/word/add", func(c *gin.Context) {
		var word Word
		var res JsonResponse[string]
		if err := c.ShouldBindJSON(&word); err != nil {
			log.Fatal(err)
			res.Error = err.Error()
			c.JSON(http.StatusBadRequest, res)
			return
		}

		now := time.Now()
		word.Created_At = now.Format("2006-01-02 15:04:05")
		_, err := AddWord(word)
		if err != nil {
			log.Fatal(err)
			res.Error = err.Error()
			c.JSON(http.StatusBadRequest, res)
			return
		}

		c.JSON(http.StatusOK, res)
	})

	r.Run() // listen and serve on 0.0.0.0:8080
}

func setup() bool {
	dbDir := os.Getenv("DB_LOCATION")
	log.Println(dbDir)
	if len(dbDir) <= 0 {
		log.Println("Env DB_LOCATION is missing!")
		return false
	}

	if _, err := os.Stat(dbDir); errors.Is(err, os.ErrNotExist) {
		os.Mkdir(dbDir, os.ModePerm)
	}

	dbLocation = filepath.Join(dbDir, "lark.db")
	os.OpenFile(dbLocation, os.O_RDONLY|os.O_CREATE, os.ModePerm)

	return setupDb()
}

func setupDb() bool {
	database, _ := sql.Open("sqlite3", dbLocation)
	statement, _ := database.Prepare(`
		CREATE TABLE IF NOT EXISTS word(
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			content TEXT NOT NULL,
			created_at TEXT NOT NULL
		)
	`)
	statement.Exec()

	return true
}

func GetAll() ([]Word, error) {
	database, _ := sql.Open("sqlite3", dbLocation)
	rows, err := database.Query("SELECT * FROM word ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var res []Word
	for rows.Next() {
		var word Word
		if err := rows.Scan(&word.Id, &word.Content, &word.Created_At); err != nil {
			return nil, err
		}

		res = append(res, word)
	}

	return res, nil
}

func AddWord(word Word) (*Word, error) {
	log.Println("add new word ...")
	database, _ := sql.Open("sqlite3", dbLocation)
	res, err := database.Exec("INSERT INTO word(content, created_at) VALUES(?,?)", word.Content, word.Created_At)
	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	word.Id = id
	log.Println("finished add new word")
	return &word, nil
}

type Word struct {
	Id         int64
	Content    string
	Created_At string
}

type JsonResponse[T any] struct {
	Error string
	Data  T
}
