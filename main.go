package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"gopkg.in/toast.v1"
	"gopkg.in/yaml.v2"
)

const (
	// Notify Const
	APPNAME = "CFWHelper"
	TITLE   = "Clash Global Proxy"
	// Later in minute
	LATER = 19

	// CFW Const
	// Config Url
	URL = "http://127.0.0.1:9090/configs"
	// Back-ground query interval in second
	INTERVAL = 60
)

var (
	// Notify Action
	ACTION = toast.Action{"protocol", "Remind Later", "remindLater"}
)

func main() {
	lastGlobal := time.Time{}

	c := &Cfw{
		NotifyGlobal: func(flag bool) {
			if !flag {
				lastGlobal = time.Time{}
				return
			}

			// Flag == True
			if lastGlobal.IsZero() {
				// first time
				lastGlobal = time.Now()
			} else if time.Since(lastGlobal).Minutes() > LATER {
				notification := toast.Notification{
					AppID: APPNAME,
					Title: TITLE,
					Actions: []toast.Action{
						ACTION,
					},
				}

				err := notification.Push()
				if err != nil {
					log.Fatalln(err)
				}

				lastGlobal = time.Now()
			}
		},

		Url:      URL,
		Interval: time.Duration(INTERVAL),
	}

	c.Listen()
}

type Cfw struct {
	NotifyGlobal func(bool)

	// Config Url
	Url string

	// Query interval in second
	Interval time.Duration
}

func (c *Cfw) Listen() {
	for range time.Tick(time.Second * c.Interval) {
		config, err := c.queryConfig()
		if err != nil {
			log.Fatal(err)
		}

		// Global
		if config["mode"] == "global" {
			if c.NotifyGlobal != nil {
				c.NotifyGlobal(true)
			}
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
