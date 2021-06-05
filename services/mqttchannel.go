package services

import (
	"fmt"
	"log"
	"net/url"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/hb-go/json"
)

type TiniyoMqttClient struct {
	topic  string
	client mqtt.Client
}

func connect(clientId string, uri *url.URL) mqtt.Client {
	opts := createClientOptions(clientId, uri)
	client := mqtt.NewClient(opts)
	token := client.Connect()
	for !token.WaitTimeout(3 * time.Second) {
	}
	if err := token.Error(); err != nil {
		log.Fatal(err)
	}
	return client
}

func createClientOptions(clientId string, uri *url.URL) *mqtt.ClientOptions {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s", uri.Host))
	opts.SetUsername(uri.User.Username())
	password, _ := uri.User.Password()
	opts.SetPassword(password)
	opts.SetClientID(clientId)
	return opts
}

func listen(uri *url.URL, topic string) {
	client := connect("sub", uri)
	client.Subscribe(topic, 0, func(client mqtt.Client, msg mqtt.Message) {
		fmt.Printf("* [%s] %s\n", msg.Topic(), string(msg.Payload()))
	})
}

type ConferencePacket struct {
	InfoType  string `json:"type"` /* request / notify / response */
	UserName  string `json:"user_name"`
	UserID    string `json:"user_id"`
	EventType string `json:"event_type"`
	EventName string `json:"event_name"`
	EventData string `json:"event_data"`
}

// export CLOUDMQTT_URL=mqtt://<user>:<pass>@<server>.cloudmqtt.com:<port>/<topic>

func (tmc *TiniyoMqttClient) Publish(cp ConferencePacket) {
	b, _ := json.Marshal(cp)
	tmc.client.Publish(tmc.topic, 0, false, b)
}

func (tmc *TiniyoMqttClient) Initialize(topic string) {
	tmc.topic = topic
	uri, err := url.Parse("mqtt://127.0.0.1:1883/" + topic)
	if err != nil {
		log.Fatal(err)
	}
	go listen(uri, topic)
	tmc.client = connect("pub", uri)
}

/*
func main() {
	tmc := new(TiniyoMqttClient)
	tmc.Initialize("timetest")

	timer := time.NewTicker(2 * time.Second)
	cp := ConferencePacket{}
	cp.InfoType = "notify"
	cp.UserName = "shailesh"
	cp.UserID = "part-495e77c5-238e-46b8-9dc6-97202b0bb1fe"
	cp.EventType = "add-member"

	for t := range timer.C {
		cp.EventData = t.String()
		tmc.Publish(cp)
	}
}
*/
