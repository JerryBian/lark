package main

import (
	"database/sql"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	L "github.com/JerryBian/lark/internal"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed img/* style.css html/* bootstrap.min.js bootstrap.min.css
var f embed.FS
var dbConnStr string
var sessionSecret []byte
var validUserName string
var validPassword string
var stopping bool
var lastSavedAt time.Time
var lastModifiedAt time.Time

const userkey = "user"

func main() {
	log.SetOutput(os.Stdout)

	var setupResult = setup()
	if setupResult == false {
		os.Exit(1)
	}

	r := gin.Default()
	r.Use(sessions.Sessions("_lark_", sessions.NewCookieStore(sessionSecret)))

	templ := template.Must(template.New("").ParseFS(f, "html/*"))
	r.SetHTMLTemplate(templ)

	r.StaticFS("/asset", http.FS(f))

	r.GET("/login", func(c *gin.Context){
		c.HTML(http.StatusOK, "login.html", gin.H{})
	})

	r.POST("/login", loginHander)
	r.GET("/logout", logoutHandler)

	authRoute := r.Group("/")
	authRoute.Use(AuthRequired)
	authRoute.GET("/", indexHandler)
	authRoute.POST("/api/word/add", addWordHandler)

	go uploadData()

	r.Run() // listen and serve on 0.0.0.0:8080
}

func uploadData() {
	now := time.Now()
	lastModifiedAt = now
	lastSavedAt = now
	L.CreateContainerIfNotExists()

	for !stopping {
		time.Sleep(time.Hour)
		if lastSavedAt.After(lastModifiedAt) {
			continue
		}

		words, err := GetAll()
		if err != nil{
			log.Println(err)
		}

		if err != nil {
			log.Println(err)
		} else {
			b, err:=json.Marshal(words)
			if err != nil {
				log.Println(err)
			} else {
				err = L.Save("", b)
				if err != nil {
					log.Println(err)
				} else {
					lastSavedAt = time.Now().Local()
					log.Println("JSON saved successfully.")
				}

				err = L.Save("-" + time.Now().Local().Format("20060102"), b)
				if err != nil {
					log.Println(err)
				} else {
					lastSavedAt = time.Now().Local()
					log.Println("JSON snapshot saved successfully.")
				}
			}
		}
	}
}

func logoutHandler(c *gin.Context){
	session := sessions.Default(c)
	user := session.Get(userkey)

	if user == nil {
		log.Println("Logout error: Invalid session token")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session token"})
		return
	}

	session.Delete(userkey)

	if err := session.Save(); err != nil {
		log.Println("Logout error: Failed to save session")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
		return
	}

	log.Println("Logout successfully.")
	c.Redirect(http.StatusFound, "/login")
}

func loginHander(c *gin.Context){
	session := sessions.Default(c)
	userName := c.PostForm("userName")
	password := c.PostForm("password")

	if strings.Trim(userName, " ") == "" || strings.Trim(password, " ") == "" {
		log.Println("Login error: missing username or password.")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid username/password."})
		return
	}

	if userName != validUserName || password != validPassword {
		log.Printf("Login error: invalid username/password(%s/%s).", userName, password)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username/password."})
		return
	}

	session.Options(sessions.Options{ MaxAge: 60 * 60 * 24 * 7, HttpOnly: true })
	session.Set(userkey, userName)
	if err := session.Save(); err != nil {
		log.Println("Login error: Failed to save session")
		c.JSON(http.StatusInternalServerError, gin.H{ "error": "failed to save session." })
		return
	}

	log.Println("Login successfully.")
	c.Redirect(http.StatusFound, "/")
}

func addWordHandler(c *gin.Context) {
	var word Word
	var res JsonResponse[string]
	if err := c.ShouldBindJSON(&word); err != nil {
		log.Fatal(err)
		res.Error = err.Error()
		c.JSON(http.StatusBadRequest, res)
		return
	}

	now := time.Now().Local()
	word.Created_At = now.Format("2006-01-02 15:04:05")
	_, err := AddWord(word)
	if err != nil {
		log.Fatal(err)
		res.Error = err.Error()
		c.JSON(http.StatusBadRequest, res)
		return
	}

	c.JSON(http.StatusOK, res)
}

func indexHandler(c *gin.Context) {
	words, err := GetAll()
	if err != nil {
		log.Fatal(err)
		c.String(http.StatusBadRequest, "Something is wrong.")
	}

	c.HTML(http.StatusOK, "index.html", gin.H{
		"Words": words,
	})
}

func AuthRequired(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get(userkey)
	if user == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	c.Next()
}

func setup() bool {
	dbDir := os.Getenv("DB_LOCATION")
	log.Println(dbDir)
	if len(dbDir) <= 0 {
		log.Println("Env DB_LOCATION is missing!")
		return false
	}

	dbDir, err := filepath.Abs(dbDir)
	if err != nil {
		log.Println(err)
	}

	if _, err := os.Stat(dbDir); errors.Is(err, os.ErrNotExist) {
		err = os.MkdirAll(dbDir, os.ModePerm)
		if err != nil {
			log.Println(err)
		}
	}

	dbLocation := filepath.Join(dbDir, "lark.db")
	if err != nil {
		log.Println(err)
	}
	
	//os.OpenFile(dbLocation, os.O_RDONLY|os.O_CREATE, os.ModePerm)

	validUserName = os.Getenv("USERNAME")
	if len(validUserName) <= 0{
		log.Println("Env USERNAME is missing!")
		return false
	}

	validPassword = os.Getenv("PASSWORD")
	if len(validUserName) <= 0{
		log.Println("Env PASSWORD is missing!")
		return false
	}

	sessionSecret = []byte(time.Now().Local().Format("2006-01-02 15:04:05"))
	dbConnStr = fmt.Sprintf("file:%s?cache=shared&mode=rwc", dbLocation)
	return setupDb()
}

func setupDb() bool {
	database, err := sql.Open("sqlite3", dbConnStr)
	if err != nil {
		log.Println(err)
	}
	//defer database.Close()

	database.Ping()
	statement, err := database.Prepare(`
		CREATE TABLE IF NOT EXISTS word(
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			content TEXT NOT NULL,
			created_at TEXT NOT NULL
		)
	`)

	if err != nil {
		log.Println(err)
	}

	statement.Exec()

	rows, _ := database.Query(`
		SELECT EXISTS (SELECT 1 FROM word)
	`)

	recordExists := true
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(&recordExists); err != nil {
			log.Println(err)
		}
	}

	// try to load from Azure Blob
	if !recordExists {
		b, err := L.Load()
		if err != nil {
			log.Println(err)
		} else {
			var words []Word
			err=json.Unmarshal(b, &words)
			if err != nil {
				log.Println(err)
			} else{
				for _, item := range words {
					_, err = AddWord(item)
					if err != nil {
						log.Println(err)
					}
				}
			}
		}
	}

	return true
}

func GetAll() ([]Word, error) {
	database, _ := sql.Open("sqlite3", dbConnStr)
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
	database, _ := sql.Open("sqlite3", dbConnStr)
	res, err := database.Exec("INSERT INTO word(content, created_at) VALUES(?,?)", word.Content, word.Created_At)
	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	word.Id = id
	lastModifiedAt = time.Now().Local()
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
