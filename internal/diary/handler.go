package internal

import (
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"

	C "github.com/JerryBian/lark/internal/config"
	I "github.com/JerryBian/lark/internal/core"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/parser"
)

func init() {
	log.SetOutput(os.Stdout)
}

type Handler struct {
	Conf *C.Config
}

func (h *Handler) Run() {
	r := gin.Default()
	sessionSecret := []byte(h.Conf.Server.SessionSecret)
	r.Use(sessions.Sessions("_lark_", sessions.NewCookieStore(sessionSecret)))

	templ := template.Must(template.New("").ParseFS(h.Conf.Runtime.F, "internal/diary/html/*.html"))
	r.SetHTMLTemplate(templ)

	staticFs, _ := fs.Sub(h.Conf.Runtime.F, "static")
	r.StaticFS("/static", http.FS(staticFs))

	r.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", gin.H{
			"Title": "Login",
			"Config": h.Conf,
		})
	})

	r.POST("/login", h.loginHander)
	r.GET("/logout", logoutHandler)

	authRoute := r.Group("/")
	authRoute.Use(AuthRequired)
	authRoute.GET("/", h.indexHandler)
	authRoute.GET("/add", h.addDiaryGetHandler)
	authRoute.POST("/api/add", h.addWordHandler)
	authRoute.GET("/:year/:month/:day", h.getDiariesHandler)
	authRoute.GET("/edit/:id", h.editDiaryGetHandler)
	authRoute.GET("/revision/:id", h.revisionHandler)

	r.Run()
}

func (h *Handler) indexHandler(c *gin.Context) {
	navs, err := getDiaryNavs(h.Conf)
	if err != nil {
		log.Println(err)
		c.String(http.StatusBadRequest, "Something is wrong.")
	}

	repo := Db{Conf: h.Conf}

	d, err := repo.GetLatestDiaries(30)
	if err != nil {
		log.Println(err)
		c.String(http.StatusBadRequest, "Something is wrong.")
		return
	}

	NormalizeAll(d)
	c.HTML(http.StatusOK, "index.html", gin.H{
		"Navs": navs,
		"Title": "首页",
		"Config": h.Conf,
		"D": d,
	})
}

func (h *Handler) revisionHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id")); if err != nil {
		log.Println(err)
		c.String(http.StatusBadRequest, "Something is wrong.")
		return
	}

	repo := Db{Conf: h.Conf}
	d, err := repo.GetDiaryById(id)
	if err != nil{
		log.Println(err)
		c.String(http.StatusBadRequest, "Something is wrong.")
		return
	}

	if d.Id == 0 {
		log.Printf("diary id %v not found for edit", id)
		c.String(http.StatusBadRequest, "Something is wrong.")
		return
	}
	
	navs, err := getDiaryNavs(h.Conf)
	if err != nil {
		log.Println(err)
		c.String(http.StatusBadRequest, "Something is wrong.")
	}

	Normalize(&d)

	c.HTML(http.StatusOK, "revision.html", gin.H{
		"Navs": navs,
		"Title": fmt.Sprintf("版本：%v", id),
		"Config": h.Conf,
		"D": d,
	})
}

func (h *Handler) editDiaryGetHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id")); if err != nil {
		log.Println(err)
		c.String(http.StatusBadRequest, "Something is wrong.")
		return
	}

	repo := Db{Conf: h.Conf}
	d, err := repo.GetDiaryById(id)
	if err != nil{
		log.Println(err)
		c.String(http.StatusBadRequest, "Something is wrong.")
		return
	}

	if d.Id == 0 {
		log.Printf("diary id %v not found for edit", id)
		c.String(http.StatusBadRequest, "Something is wrong.")
		return
	}

	navs, err := getDiaryNavs(h.Conf)
	if err != nil {
		log.Println(err)
		c.String(http.StatusBadRequest, "Something is wrong.")
		return
	}

	c.HTML(http.StatusOK, "edit.html", gin.H{
		"D": d,
		"Navs": navs,
		"Title": fmt.Sprintf("编辑：%v", d.Id),
		"Config": h.Conf,
	})
}

func (h *Handler) getDiariesHandler(c *gin.Context) {
	year, err := strconv.Atoi(c.Param("year")); if err != nil {
		log.Println(err)
		c.String(http.StatusBadRequest, "Something is wrong.")
		return
	}

	month, err := strconv.Atoi(c.Param("month")); if err != nil {
		log.Println(err)
		c.String(http.StatusBadRequest, "Something is wrong.")
		return
	}

	day, err := strconv.Atoi(c.Param("day")); if err != nil {
		log.Println(err)
		c.String(http.StatusBadRequest, "Something is wrong.")
		return
	}

	repo := Db{Conf: h.Conf}

	v, err := repo.GetDiaries(year, month, day)
	if err != nil {
		log.Println(err)
		c.String(http.StatusBadRequest, "Something is wrong.")
		return
	}

	NormalizeAll(v.Diaries)

	navs, err := getDiaryNavs(h.Conf)
	if err != nil {
		log.Println(err)
		c.String(http.StatusBadRequest, "Something is wrong.")
		return
	}

	c.HTML(http.StatusOK, "dayview.html", gin.H{
		"V": v,
		"Navs": navs,
		"Title": fmt.Sprintf("%v年%v月%v日", year, month, day),
		"Config": h.Conf,
		"ActiveDiaryLink": fmt.Sprintf("/%04d/%02d/%02d", year, month, day),
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

func (h *Handler) addDiaryGetHandler(c *gin.Context) {
	navs, err := getDiaryNavs(h.Conf)
	if err != nil {
		log.Println(err)
		c.String(http.StatusBadRequest, "Something is wrong.")
		return
	}

	c.HTML(http.StatusOK, "add.html", gin.H{
		"Navs": navs,
		"Title": "添加日记",
		"Config": h.Conf,
	})
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

	now := time.Now().UnixMicro()
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

func getDiaryNavs(c *C.Config) ([]DiaryNav, error) {
	repo := Db{Conf: c}

	navs, err := repo.GetDiaryNavs()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return navs, nil
}

func NormalizeAll(ds []Diary) {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs

	for i := range ds {
		initCreatedTime := time.UnixMicro(ds[i].CreatedAt)
		ds[i].DayString = initCreatedTime.Format("2006年01月02日")
		ds[i].DayLink = initCreatedTime.Format("/2006/01/02")
		ds[i].Revisions = len(ds[i].Contents)
		ds[i].CreatedAtString = initCreatedTime.Format("2006年01月02日 15时04分")
		ds[i].TimeOnlyStr = initCreatedTime.Format("15时04分05秒")

		for j := range ds[i].Contents {
			parser := parser.NewWithExtensions(extensions)
			h := string(markdown.ToHTML([]byte(ds[i].Contents[j].Content), parser, nil))
			ds[i].Contents[j].HtmlContent = template.HTML(h)

			createdAt := time.UnixMicro(ds[i].Contents[j].CreatedAt)
			ds[i].Contents[j].CreatedAtString = createdAt.Format("2006年01月02日 15时04分")
			ds[i].Contents[j].ContentLen = utf8.RuneCountInString(ds[i].Contents[j].Content)
		}
	}
} 

func Normalize(d *Diary) {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs
	initCreatedTime := time.UnixMicro(d.CreatedAt)
	d.DayString = initCreatedTime.Format("2006年01月02日")
	d.DayLink = initCreatedTime.Format("/2006/01/02")
	d.Revisions = len(d.Contents)
	d.CreatedAtString = initCreatedTime.Format("2006年01月02日 15时04分")
	d.TimeOnlyStr = initCreatedTime.Format("15时04分05秒")

	for j := range d.Contents {
		parser := parser.NewWithExtensions(extensions)
		h := string(markdown.ToHTML([]byte(d.Contents[j].Content), parser, nil))
		d.Contents[j].HtmlContent = template.HTML(h)

		createdAt := time.UnixMicro(d.Contents[j].CreatedAt)
		d.Contents[j].CreatedAtString = createdAt.Format("2006年01月02日 15时04分")
		d.Contents[j].ContentLen = utf8.RuneCountInString(d.Contents[j].Content)
	}
} 