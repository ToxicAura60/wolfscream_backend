package scheduler

import (
	"github.com/robfig/cron/v3"
)

var Cron *cron.Cron

func init() {
    Cron = cron.New()
    Cron.Start()
}