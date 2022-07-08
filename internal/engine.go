package internal

import (
	C "github.com/JerryBian/lark/internal/config"
	D "github.com/JerryBian/lark/internal/diary"
)

type Engine struct {
	Conf *C.Config
}

func (c *Engine) Run() {
	handler := D.Handler{Conf: c.Conf}
	handler.Run()
}