package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

const (
	versionMajor = 0
	versionMinor = 2
	versionPatch = 0
)

var (
	// Filled out when running go build using the following flag
	// -ldflags "-X ${IMPORT_PATH}/main.buildDate=${BUILD_DATE} -X ${IMPORT_PATH}/main.commit=${COMMIT}"
	buildDate = "unknown_build_date"
	commit    = "unknown_commit"
)

func main() {
	logrus.Info("Starting Slack Food Bot, version:", version())
	c := readConfig()
	checkRequiredConfig(c)
	checkDebug(c)

	slackC := Config{c.slackWebhook}

	fri := Fri{}
	friBot := NewBot(slackC, &fri, c.messageTrigger)

	bf := Bf{}
	bfBot := NewBot(slackC, &bf, c.messageTrigger)

	go func() {
		friBot.Start()
	}()

	go func() {
		bfBot.Start()
	}()

	// Exit when receive SIGINT or SIGTERM signal
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
}

type config struct {
	debug          bool
	slackWebhook   string
	messageTrigger time.Time
}

func readConfig() config {
	c := config{}

	s, b := os.LookupEnv("DEBUG")
	c.debug = b && strings.ToLower(s) == "true"

	s, b = os.LookupEnv("SLACK_WEBHOOK")
	if b {
		c.slackWebhook = s
	}

	s, b = os.LookupEnv("MESSAGE_TRIGGER")
	if b {
		mt, err := time.Parse("15:04", s)
		if err != nil {
			logrus.Error("Failed to parse MESSAGE_TRIGGER, must be in format 15:04")
			os.Exit(1)
		}

		c.messageTrigger = mt
	}

	return c
}

func checkRequiredConfig(c config) {
	if c.slackWebhook == "" {
		logrus.Error("No SLACK_WEBHOOK env")
		os.Exit(1)
	}

	if c.messageTrigger.IsZero() {
		logrus.Error("No MESSAGE_TRIGGER env")
		os.Exit(1)
	}

}

func checkDebug(c config) {
	if c.debug {
		logrus.Info("Enabling debug logging")
		logrus.SetLevel(logrus.DebugLevel)
	}
}

func version() string {
	return fmt.Sprintf("%d.%d.%d-%s-%s", versionMajor, versionMinor, versionPatch, buildDate, commit)
}
