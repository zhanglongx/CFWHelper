package main

import (
	"errors"
	"flag"
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
	VERSION = "1.0.2"

	// Notify Const
	APPNAME          = "CFWHelper"
	TITLE_PROXY_MODE = "Clash Not Rule"
	TITLE_ALLOW_LAN  = "Clash Allow Lan"
	// Later in minute
	LATER = 19
	// Maximum notifications
	MAXNOTIFICATIONS = 3

	// CFW Const
	// Config Yml
	CONFIGYML = "d:\\Users\\zhlx\\.config\\clash\\config.yaml"
	// Back-ground query interval in second
	INTERVAL = 60
)

var errLog *log.Logger

func main() {
	optVer := flag.Bool("version", false, "print version")

	flag.Parse()

	if *optVer {
		fmt.Printf("CFWHelper %s\n", VERSION)
		os.Exit(0)
	}

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

	errLog.Printf("CFWHelper %s", VERSION)
	errLog.Println("started")

	c := &Cfw{
		NotifyNotRule:  NotificationHelper(TITLE_PROXY_MODE),
		NotifyAllowLan: NotificationHelper(TITLE_ALLOW_LAN),

		Interval: time.Duration(INTERVAL),
	}

	err = c.LoadYML(CONFIGYML)
	if err != nil {
		log.Fatal(err)
	}

	c.Listen()
}

func NotificationHelper(title string) func(bool) {
	lastTime := time.Time{}
	notifyCnt := 0

	return func(flag bool) {
		if !flag {
			lastTime = time.Time{}
			notifyCnt = 0
			return
		}

		// Flag == True
		if notifyCnt >= MAXNOTIFICATIONS {
			return
		}

		if lastTime.IsZero() {
			// first time
			lastTime = time.Now()
		} else if time.Since(lastTime).Minutes() > LATER {
			notification := toast.Notification{
				AppID: APPNAME,
				Title: title,
			}

			err := notification.Push()
			if err != nil {
				errLog.Fatal(err)
			}

			lastTime = time.Now()
			notifyCnt++
		}
	}
}

type Cfw struct {
	NotifyNotRule  func(bool)
	NotifyAllowLan func(bool)

	// Config url
	url    string
	secret string

	// Query interval in second
	Interval time.Duration
}

func (c *Cfw) LoadYML(configYml string) error {
	yml, err := ioutil.ReadFile(configYml)
	if err != nil {
		return err
	}

	m := make(map[interface{}]interface{})
	err = yaml.Unmarshal(yml, m)
	if err != nil {
		return err
	}

	if m["external-controller"] == nil {
		return errors.New("external-controller not exists")
	}

	c.url = fmt.Sprintf("http://%s/configs", m["external-controller"].(string))

	if m["secret"] != nil {
		c.secret = m["secret"].(string)
	}

	return nil
}

func (c *Cfw) Listen() {
	for range time.Tick(time.Second * c.Interval) {
		config, err := c.queryConfig()
		if err != nil {
			errLog.Println(err)

			continue
		}

		// Proxy Mode
		if c.NotifyNotRule != nil {
			c.NotifyNotRule(config["mode"] != "rule")
		}

		// Allow Lan
		if c.NotifyAllowLan != nil {
			c.NotifyAllowLan(config["allow-lan"] == true)
		}
	}
}

func (c *Cfw) queryConfig() (map[interface{}]interface{}, error) {
	req, err := http.NewRequest("GET", c.url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.secret))

	client := &http.Client{}
	resp, err := client.Do(req)

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
