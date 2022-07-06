package main

import (
	"embed"
	"log"
	"os"

	C "github.com/JerryBian/lark/internal/config"
	P "github.com/JerryBian/lark/internal/plugin"
	I "github.com/JerryBian/lark/internal"
)

//go:embed  internal/startup.sql internal/diary/html/* internal/config/default.yml static/*
var f embed.FS
var AppVer = "1.0"
var GitHash = "1234567"
var BuildTime = "2000-01-01"

func main() {
	log.SetOutput(os.Stdout)

	config := C.Config { }
	config.Runtime.F = &f
	config.Load()
	config.Runtime.AppVer = AppVer
	config.Runtime.GitHash = GitHash
	config.Runtime.BuildTime = BuildTime

	repo := I.Sqlite{Conf: &config}
	repo.Startup()

	plugin := P.Plugin{Conf: &config}
	plugin.Run()
	
	handler := I.Engine{Conf: &config}
	handler.Run()
}
