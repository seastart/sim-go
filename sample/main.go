package main

import (
	"fmt"
	"log"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	emitter "github.com/seastart/sim-go/v2"
)

//key
const key = "ddjHNjFMK97u25aXKkC9alPOinMKdZ0z"

// appid:channel
const appid = "stnv3uld-t40uu9pzw7ucozmcepolisu"
const powerch = ""

// 订阅topic时是否根据业务uid互斥
const mutex = false

// 业务端用户信息 appid:channel|uid
// 改为sdk自动拼
// var uname = appid + ":" + powerch + "|" + "50j8d";
const uname = "50j8d"

// 房间topic appid:channel/topic
// 改为sdk自动拼
// var topic =  appid + ":" + powerch + "/live/nyjq1";
const topic = "live/nyjq1"
const host = "tcps://im.open.seastart.cn:8883"

const debug = true

func main() {
	if debug {
		mqtt.ERROR = log.New(os.Stdout, "[ERROR] ", 0)
		mqtt.CRITICAL = log.New(os.Stdout, "[CRIT] ", 0)
		mqtt.WARN = log.New(os.Stdout, "[WARN]  ", 0)
		mqtt.DEBUG = log.New(os.Stdout, "[DEBUG] ", 0)
	}

	clientA()
	clientB()

	// stop after 10 seconds
	time.Sleep(100 * time.Second)
}

func clientA() {
	// 可以设置全局默认的msg handler
	c, err := emitter.Connect(host, func(_ *emitter.Client, msg emitter.Message) {
		fmt.Printf("[emitter] -> [A] received: '%s' topic: '%s'\n", msg.Payload(), msg.Topic())
	}, emitter.WithAppid(appid), emitter.WithPowerch(powerch), emitter.WithUsername(uname))

	if err != nil {
		fmt.Printf("A connect error %v\n", err)
	}
	c.OnError(func(c *emitter.Client, e emitter.Error) {
		fmt.Printf("emitter error %v\n", e)
	})

	// Subscribe to topic
	// 也可以设置单个topic的msg handler
	if mutex {
		c.Subscribe(key, topic, nil, emitter.WithMutex())
	} else {
		c.Subscribe(key, topic, nil, emitter.WithMutex())
	}
}

func clientB() {
	// Create the client and connect to the broker
	c, err := emitter.Connect(host, func(_ *emitter.Client, msg emitter.Message) {
		fmt.Printf("[emitter] -> [B] received: '%s' topic: '%s'\n", msg.Payload(), msg.Topic())
	}, emitter.WithAppid(appid), emitter.WithPowerch(powerch), emitter.WithUsername("shu2"))
	if err != nil {
		fmt.Printf("B connect error %v\n", err)
	}
	// Set the presence handler
	c.OnPresence(func(_ *emitter.Client, ev emitter.PresenceEvent) {
		fmt.Printf("[emitter] -> [B] presence event: %d subscriber(s) at topic: '%s'\n", len(ev.Who), ev.Channel)
	})

	id := c.ID()
	fmt.Println("[emitter] -> [B] my name is " + id)

	// Subscribe to topic
	c.Subscribe(key, topic, func(_ *emitter.Client, msg emitter.Message) {
		fmt.Printf("[emitter] -> [B] received on specific handler: '%s' topic: '%s'\n", msg.Payload(), msg.Topic())
	})

	// Ask for presence
	c.Presence(key, topic, true, false)

	// Publish to the channel
	c.Publish(key, topic, "hello")
}
