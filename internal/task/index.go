package task

import (
	"github.com/robfig/cron/v3"
)

var c *cron.Cron

func init() {
	c = cron.New(cron.WithSeconds())
}

func Start() {
	c.Start()
}

func Stop() {
	c.Stop()
}
