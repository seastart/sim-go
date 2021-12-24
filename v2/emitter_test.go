package emitter

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

//key
const key = "I8Hzqy5xCJ3ohoHFVCOVWApg-1vgiyK-"

// appid:channel
const appid = "sturgm9rwhabef_h5gu5_ldgvm_5gsdn"
const powerch = ""

// 订阅topic时是否根据业务uid互斥
const mutex = true

// 业务端用户信息 appid:channel|uid
// 改为sdk自动拼
// var uname = appid + ":" + powerch + "|" + "50j8d";
const uname = "50j8d"

// 房间topic appid:channel/topic
// 改为sdk自动拼
// var topic =  appid + ":" + powerch + "/live/nyjq1";
const topic = "live/nyjq1"
const host = "tcps://im.open.seastart.cn:8883"

func TestEndToEnd(t *testing.T) {
	clientA(t)
	clientB(t)

	// stop after 10 seconds
	time.Sleep(1 * time.Second)
}

func clientA(t *testing.T) {

	// Create the client and connect to the broker
	c, _ := Connect(host, func(_ *Client, msg Message) {
		fmt.Printf("[emitter] -> [A] received: '%s' topic: '%s'\n", msg.Payload(), msg.Topic())
	}, WithAppid(appid), WithPowerch(powerch), WithUsername(uname))

	// Subscribe to demo channel
	fmt.Println("[emitter] <- [A] subscribing to 'demo/...'")
	err := c.Subscribe(key, "sdk-integration-test/", nil)
	assert.NoError(t, err)
}

func clientB(t *testing.T) {

	// Create the client and connect to the broker
	c, _ := Connect(host, func(_ *Client, msg Message) {
		fmt.Printf("[emitter] -> [B] received: '%s' topic: '%s'\n", msg.Payload(), msg.Topic())
	}, WithAppid(appid), WithPowerch(powerch), WithUsername(uname))

	c.OnPresence(func(_ *Client, ev PresenceEvent) {
		fmt.Printf("[emitter] -> [B] presence event: %d subscriber(s) at topic: '%s'\n", len(ev.Who), ev.Channel)
	})

	fmt.Println("[emitter] <- [B] querying own name")
	id := c.ID()
	assert.NotEmpty(t, id)

	// Subscribe to demo channel
	c.Subscribe(key, "sdk-integration-test/", func(_ *Client, msg Message) {
		fmt.Printf("[emitter] -> [B] received on specific handler: '%s' topic: '%s'\n", msg.Payload(), msg.Topic())
	})

	// Ask for presence
	fmt.Println("[emitter] <- [B] asking for presence on 'sdk-integration-test/'")
	err := c.Presence(key, "sdk-integration-test/", true, false)
	assert.NoError(t, err)

	// Publish to the channel
	fmt.Println("[emitter] <- [B] publishing to 'sdk-integration-test/'")
	err = c.Publish(key, "sdk-integration-test/", "hello")
	assert.NoError(t, err)
}

func TestFormatTopic(t *testing.T) {
	const key = "pGrtRRL6RrjAdExSArkMzBZOoWr2eB3L" // everything on sdk-integration-test/

	// Create the client and connect to the broker
	c, _ := Connect("", func(_ *Client, msg Message) {
		fmt.Printf("[emitter] -> [B] received: '%s' topic: '%s'\n", msg.Payload(), msg.Topic())
	})

	tests := []struct {
		key     string
		channel string
		options []Option
		result  string
	}{
		{channel: "a/b/c", result: "a/b/c/"},
		{key: "key", channel: "channel", result: "key/channel/"},
		{key: "key", channel: "a/b/c", result: "key/a/b/c/"},
		{key: "key", channel: "channel", options: []Option{WithMutex()}, result: "key/channel/?mutex=1"},
		{key: "key", channel: "a/b/c", options: []Option{WithoutEcho()}, result: "key/a/b/c/?me=0"},
		{key: "key", channel: "a/b/c", options: []Option{WithoutEcho(), WithAtLeastOnce(), WithLast(100)}, result: "key/a/b/c/?me=0&last=100"},
		{key: "key", channel: "a/b/c", options: []Option{WithAtLeastOnce(), WithoutEcho(), WithLast(100)}, result: "key/a/b/c/?me=0&last=100"},
		{key: "key", channel: "a/b/c", options: []Option{WithoutEcho(), WithLast(100), WithAtLeastOnce()}, result: "key/a/b/c/?me=0&last=100"},
	}

	for _, tc := range tests {
		topic := c.formatTopic(tc.key, tc.channel, tc.options)
		assert.Equal(t, tc.result, topic)
	}
}

func TestGetHeader(t *testing.T) {
	tests := []struct {
		options []Option
		qos     byte
		retain  bool
	}{

		{options: []Option{WithoutEcho()}, qos: 0, retain: false},
		{options: []Option{WithoutEcho(), WithAtLeastOnce(), WithLast(100)}, qos: 1, retain: false},
		{options: []Option{WithAtLeastOnce(), WithoutEcho(), WithLast(100)}, qos: 1, retain: false},
		{options: []Option{WithoutEcho(), WithLast(100), WithAtLeastOnce()}, qos: 1, retain: false},
		{options: []Option{WithoutEcho(), WithRetain(), WithAtMostOnce()}, qos: 0, retain: true},
	}

	for _, tc := range tests {
		qos, retain := getHeader(tc.options)
		assert.Equal(t, tc.qos, qos)
		assert.Equal(t, tc.retain, retain)
	}
}

func TestFormatShare(t *testing.T) {
	topic := formatShare("/key/", "share1", "/a/b/c/", []Option{WithoutEcho()})
	assert.Equal(t, "key/$share/share1/a/b/c/?me=0", topic)
}

func TestPresence(t *testing.T) {
	c := NewClient()

	var events []PresenceEvent
	c.OnPresence(func(_ *Client, ev PresenceEvent) {
		events = append(events, ev)
	})

	c.onMessage(nil, &message{
		topic:   "emitter/presence/",
		payload: ` {"time":1589626821,"event":"status","channel":"retain-demo/","who":[{"id":"B"}, {"id":"C"}]}`,
	})

	c.onMessage(nil, &message{
		topic:   "emitter/presence/",
		payload: ` {"time":1589626821,"event":"subscribe","channel":"retain-demo/","who":{"id":"A"}}`,
	})

	assert.Equal(t, 2, len(events))
}
