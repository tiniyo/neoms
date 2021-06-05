package services

import (
	"log"

	nsq "github.com/nsqio/go-nsq"
)

var w *nsq.Producer

func InitializeNsqProducer() {
	config := nsq.NewConfig()
	w, _ = nsq.NewProducer("127.0.0.1:4150", config)
}

func ProduceInfo() {
	err := w.Publish("write_test", []byte("test"))
	if err != nil {
		log.Panic("Could not connect")
	}
}

func ShutdownProducer() {
	w.Stop()
}
