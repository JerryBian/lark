package internal

import (
	C "github.com/JerryBian/lark/internal/config"
	AZ "github.com/JerryBian/lark/internal/plugin/az"
)

type Plugin struct {
	Conf *C.Config
}

func (p *Plugin) Run() {
	aztask := AZ.AzTask { Config: p.Conf }
	aztask.TryRestore()
	go aztask.SpinWorker()
}