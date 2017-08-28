package Notification

import (
	"github.com/0xAX/notificator"
	"path/filepath"
	"log"
)

type Desktop struct {
	notifier *notificator.Notificator
}

func NewDesktop() Desktop {
	path, _ := filepath.Abs("./")

	return Desktop{notificator.New(notificator.Options{
		DefaultIcon: path + "/timeular.png",
		AppName:     "ZEI",
	})}
}

func (n *Desktop) Notify(title, message string) {
	log.Println(message)
	n.notifier.Push(title, message, "", notificator.UR_NORMAL)
}
