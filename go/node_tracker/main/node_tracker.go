package main

import (
	"di_store/node_tracker"
	"di_store/tracing"
	"di_store/util"
	"fmt"
	"net/url"
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	logReportCaller = kingpin.Flag("log-report-caller", "whether the log will include the calling method").Default("false").Bool()
	logLevel        *string
	hostname        = kingpin.Flag("hostname", "hostname of node tracker").Default("").String()
	confPath        = kingpin.Arg("conf_path", "configure file path for node tracker").Required().File()
)

func initLogLevel() {
	levelList := make([]string, len(log.AllLevels))
	for i, level := range log.AllLevels {
		levelName, _ := level.MarshalText()
		levelList[i] = string(levelName)
	}
	logLevel = kingpin.Flag("log-level", "Log level.").Short('l').Default("info").OverrideDefaultFromEnvar("LOG_LEVEL").Enum(levelList...)
}

func main() {
	initLogLevel()
	kingpin.Parse()

	level, _ := log.ParseLevel(*logLevel)
	formatter := &log.TextFormatter{
		FullTimestamp: true,
	}
	log.SetFormatter(formatter)
	log.SetLevel(level)
	log.SetReportCaller(*logReportCaller)

	log.Debug("SetReportCaller: ", *logReportCaller)
	log.Debug("SetLevel: ", *logLevel)

	if tracing.Enabled {
		defer tracing.Close()
	}

	conf, err := util.ReadConfig((*confPath).Name())
	if err != nil {
		log.Fatal(err)
	}

	if *hostname == "" {
		h, err := os.Hostname()
		if err != nil {
			log.Fatal(err)
		}
		hostname = &h
	}
	trackerInfo, err := conf.NodeTracker(*hostname)
	if err != nil {
		log.Fatal(err)
	}

	etcdServerUrls, err := conf.EtcdServerUrls()
	if err != nil {
		log.Fatal(err)
	}

	u, err := url.Parse(etcdServerUrls[0])
	if err != nil {
		log.Fatal(err)
	}
	port, err := strconv.Atoi(u.Port())
	if err != nil {
		log.Fatal(fmt.Errorf("parse url error: %v", err))
	}

	nodeTracker, err := node_tracker.NewNodeTracker(u.Hostname(), port, trackerInfo.RpcHost, trackerInfo.RpcPort)
	if err != nil {
		log.Fatalf("failed to create NodeTracker: %v", err)
	}
	nodeTracker.Serve()
}
