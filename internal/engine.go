package internal

import (
	C "github.com/JerryBian/lark/internal/config"
	D "github.com/JerryBian/lark/internal/diary"
	"strings"
)

type Engine struct {
	Conf *C.Config
}

func (c *Engine) Run() {
	if strings.EqualFold(c.Conf.Server.Mode, "diary") {
		handler := D.Handler{Conf: c.Conf}
		handler.Run()
	} else if strings.EqualFold(c.Conf.Server.Mode, "book"){

	} else{
		panic("Server mode not supported.")
	}
}