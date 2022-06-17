package internal

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"

	C "github.com/JerryBian/lark/internal/config"
	I "github.com/JerryBian/lark/internal/core"
)

func init() {
	log.SetOutput(os.Stdout)
}

type Handler struct {
	Conf *C.Config
}

func (h *Handler) Run() {
	r := gin.Default()
	sessionSecret := []byte(time.Now().Local().Format("2006-01-02 15:04:05"))
	r.Use(sessions.Sessions("_lark_", sessions.NewCookieStore(sessionSecret)))

	templ := template.Must(template.New("").ParseFS(h.Conf.Runtime.F, "internal/diary/html/*.html"))
	r.SetHTMLTemplate(templ)

	r.StaticFS("/asset", http.Dir("static"))

	r.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", gin.H{})
	})

	r.POST("/login", h.loginHander)
	r.GET("/logout", logoutHandler)

	authRoute := r.Group("/")
	authRoute.Use(AuthRequired)
	authRoute.GET("/", h.indexHandler)
	authRoute.POST("/api/word/add", h.addWordHandler)

	r.Run() // listen and serve on 0.0.0.0:8080
}

func (h *Handler) indexHandler(c *gin.Context) {
	repo := Db{Conf: h.Conf}

	diaries, err := repo.Dump()
	if err != nil {
		log.Println(err)
		c.String(http.StatusBadRequest, "Something is wrong.")
	}

	c.HTML(http.StatusOK, "index.html", gin.H{
		"Diaries": diaries,
	})
}

func AuthRequired(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get(I.UserKey)
	if user == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	c.Next()
}

func logoutHandler(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get(I.UserKey)

	if user == nil {
		log.Println("Logout error: Invalid session token")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session token"})
		return
	}

	session.Delete(I.UserKey)

	if err := session.Save(); err != nil {
		log.Println("Logout error: Failed to save session")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
		return
	}

	log.Println("Logout successfully.")
	c.Redirect(http.StatusFound, "/login")
}

func (h *Handler) loginHander(c *gin.Context) {
	session := sessions.Default(c)
	userName := c.PostForm("userName")
	password := c.PostForm("password")

	if strings.Trim(userName, " ") == "" || strings.Trim(password, " ") == "" {
		log.Println("Login error: missing username or password.")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid username/password."})
		return
	}

	if userName != h.Conf.User.Name || password != h.Conf.User.Password {
		log.Printf("Login error: invalid username/password(%s/%s).", userName, password)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username/password."})
		return
	}

	session.Options(sessions.Options{MaxAge: 60 * 60 * 24 * 7, HttpOnly: true})
	session.Set(I.UserKey, userName)
	if err := session.Save(); err != nil {
		log.Println("Login error: Failed to save session")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save session."})
		return
	}

	log.Println("Login successfully.")
	c.Redirect(http.StatusFound, "/")
}

func (h *Handler) addWordHandler(c *gin.Context) {
	var word Diary
	var res I.JsonResponse[string]
	if err := c.ShouldBindJSON(&word); err != nil {
		log.Println(err)
		res.Error = err.Error()
		c.JSON(http.StatusBadRequest, res)
		return
	}

	if len(word.Contents) == 0 {
		log.Println("No word content specified, abort.")
		res.Error = "No word content specified, abort."
		c.JSON(http.StatusBadRequest, res)
		return
	}

	now := time.Now().UTC().UnixMicro()
	word.CreatedAt = now
	word.LastModifiedAt = now
	word.Contents[0].CreatedAt = now

	repo := Db{Conf: h.Conf}
	_, err := repo.AddDiary(word)
	if err != nil {
		log.Println(err)
		res.Error = err.Error()
		c.JSON(http.StatusBadRequest, res)
		return
	}

	c.JSON(http.StatusOK, res)
}
