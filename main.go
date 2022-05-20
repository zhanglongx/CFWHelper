package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
	"gopkg.in/toast.v1"
	"gopkg.in/yaml.v2"
)

const (
	// Notify Const
	APPNAME          = "CFWHelper"
	TITLE_PROXY_MODE = "Clash Global Proxy"
	TITLE_ALLOW_LAN  = "Clash Allow Lan"
	// Later in minute
	LATER = 19
	// Maximum notifications
	MAXNOTIFICATIONS = 3

	// CFW Const
	// Config Url
	URL = "http://127.0.0.1:9090/configs"
	// Back-ground query interval in second
	INTERVAL = 60
)

var errLog *log.Logger

func main() {
	// FIXME: is not safe to open file when fatal
	e, err := os.OpenFile("./CFWHelper.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {
		fmt.Printf("error opening file: %v", err)
		os.Exit(1)
	}

	errLog = log.New(e, "", log.Ldate|log.Ltime)
	errLog.SetOutput(&lumberjack.Logger{
		Filename:   "./CFWHelper.log",
		MaxSize:    1,  // megabytes after which new file is created
		MaxBackups: 3,  // number of backups
		MaxAge:     15, // days
	})

	errLog.Println("started")

	c := &Cfw{
		NotifyGlobal:   NotificationHelper(TITLE_PROXY_MODE),
		NotifyAllowLan: NotificationHelper(TITLE_ALLOW_LAN),

		Url:      URL,
		Interval: time.Duration(INTERVAL),
	}

	c.Listen()
}

func NotificationHelper(title string) func(bool) {
	lastGlobal := time.Time{}
	notifyTimes := 0

	return func(flag bool) {
		if !flag {
			lastGlobal = time.Time{}
			notifyTimes = 0
			return
		}

		// Flag == True
		if notifyTimes >= MAXNOTIFICATIONS {
			return
		}

		if lastGlobal.IsZero() {
			// first time
			lastGlobal = time.Now()
		} else if time.Since(lastGlobal).Minutes() > LATER {
			notification := toast.Notification{
				AppID: APPNAME,
				Title: title,
			}

			err := notification.Push()
			if err != nil {
				errLog.Fatal(err)
			}

			lastGlobal = time.Now()
			notifyTimes++
		}
	}
}

type Cfw struct {
	NotifyGlobal   func(bool)
	NotifyAllowLan func(bool)

	// Config Url
	Url string

	// Query interval in second
	Interval time.Duration
}

func (c *Cfw) Listen() {
	for range time.Tick(time.Second * c.Interval) {
		config, err := c.queryConfig()
		if err != nil {
			errLog.Fatal(err)
		}

		// Global
		if c.NotifyGlobal != nil {
			c.NotifyGlobal(config["mode"] == "global")
		}

		if c.NotifyAllowLan != nil {
			c.NotifyAllowLan(config["allow-lan"] == true)
		}
	}
}

func (c *Cfw) queryConfig() (map[interface{}]interface{}, error) {
	resp, err := http.Get(c.Url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	m := make(map[interface{}]interface{})
	err = yaml.Unmarshal(body, &m)
	if err != nil {
		return nil, err
	}

	return m, nil
}
