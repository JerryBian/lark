package internal

import (
	"embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Port int `yaml:"port" envconfig:"ENV_SERVER_PORT"`
		GinMode string `yaml:"ginMode" envconfig:"ENV_SERVER_GIN_MODE"`
		SessionSecret string `yaml:"sessionSecret" envconfig:"ENV_SERVER_SESSION_SECRET"`
		Title string `yaml:"title" envconfig:"ENV_SERVER_TITLE"`
		MaxWordsAllowed int `yaml:"maxWordsAllowed" envconfig:"ENV_SERVER_MAX_WORDS_ALLOWED"`
	} `yaml:"server"`

	Database struct {
		Location string `yaml:"location" envconfig:"ENV_DB_LOCATION"`
		ConnStr string
	} `yaml:"database"`

	User struct {
		Name string `yaml:"name" envconfig:"ENV_USER_NAME"`
		Password string `yaml:"password" envconfig:"ENV_USER_PASSWORD"`
	} `yaml:"user"`

	Az struct {
		ConnStr string `yaml:"connStr" envconfig:"ENV_AZ_CONNSTR"`
		BlobContainer string `yaml:"blobContainer" envconfig:"ENV_AZ_BLOB_CONTAINER"`
		Blob string `yaml:"blob" envconfig:"ENV_AZ_BLOB"`
		BackupIntervalInMinutes int `yaml:"backupInterval" envconfig:"ENV_AZ_BACKUP_INTERVAL"`
	} `yaml:"az"`

	Runtime struct {
		LastBackupAt time.Time
		LastModifiedAt time.Time
		F *embed.FS
		AppVer string
		GitHash string
		BuildTime string
		GoVer string
	}
}

func (c *Config) Load() {
	// load yaml
	f, err := c.Runtime.F.Open("internal/config/default.yml")
	if err != nil {
		panic(err)
	}

	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(c)
	if err != nil {
		panic(err)
	}

	// load env
	err = envconfig.Process("", c)
	if err != nil {
		panic(err)
	}

	verify(c)
}

func verify(c *Config){
	dbDir := c.Database.Location
	if len(dbDir) <= 0 {
		panic("Env ENV_DB_LOCATION is missing!")
	}

	dbDir, err := filepath.Abs(dbDir)
	if err != nil {
		panic(err)
	}

	if _, err := os.Stat(dbDir); errors.Is(err, os.ErrNotExist) {
		err = os.MkdirAll(dbDir, os.ModePerm)
		if err != nil {
			panic(err)
		}
	}

	dbFile := filepath.Join(dbDir, "lark.db")
	if err != nil {
		panic(err)
	}
	
	validUserName := c.User.Name
	if len(validUserName) <= 0{
		panic("Env ENV_USER_NAME is missing!")
	}

	validPassword := c.User.Password
	if len(validPassword) <= 0{
		panic("Env ENV_USER_PASSWORD is missing!")
	}

	if c.Az.ConnStr != "" {
		if c.Az.BackupIntervalInMinutes < 1 {
			panic("Env ENV_AZ_BACKUP_INTERVAL must be positive integer.")
		}
	}

	if c.Server.Port > 0 {
		port := strconv.Itoa(c.Server.Port)
		os.Setenv("PORT", port) // This is for gin port
	}

	if c.Server.GinMode != "" {
		os.Setenv("GIN_MODE", c.Server.GinMode)
	}

	if c.Server.SessionSecret == "" {
		c.Server.SessionSecret = time.Now().Format("2006-01-02 15:04:05")
	}

	if c.Server.MaxWordsAllowed <= 0 {
		c.Server.MaxWordsAllowed = 320
	}

	c.Database.ConnStr = fmt.Sprintf("file:%s?cache=shared&mode=rwc&_foreign_keys=on&_journal_mode=WAL", dbFile)
	c.Runtime.GoVer = runtime.Version()
}