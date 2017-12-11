package main

import (
	"github.com/0xAX/notificator"
	"log"
	"path/filepath"
)

type Notification struct {
	notifier *notificator.Notificator
}

func NewNotification() Notification {
	path, _ := filepath.Abs("./")

	return Notification{notificator.New(notificator.Options{
		DefaultIcon: path + "/timeular.png",
		AppName:     "ZEI",
	})}
}

func (n *Notification) Notify(title, message string) {
	log.Println(message)
	n.notifier.Push(title, message, "", notificator.UR_NORMAL)
}
