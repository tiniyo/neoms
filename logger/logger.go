package logger

import (
	"fmt"
	"github.com/evalphobia/logrus_sentry"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
	logrus_syslog "github.com/sirupsen/logrus/hooks/syslog"
	"log/syslog"
	"os"
	"github.com/tiniyo/neoms/config"
	"strconv"
)

var logLevel = map[string]logrus.Level{
	"debug": logrus.DebugLevel,
	"info":  logrus.InfoLevel,
	"warn":  logrus.WarnLevel,
}

var facilityLevel = map[string]syslog.Priority{
	"local0": syslog.LOG_LOCAL0,
	"local1": syslog.LOG_LOCAL1,
	"local2": syslog.LOG_LOCAL2,
	"local3": syslog.LOG_LOCAL3,
}

var Logger *logrus.Logger

func InitLogger() {
	var err error
	Logger, err = NewLogger(config.Config.Logging.Level, config.Config.Logging.Facility, config.Config.Logging.Tag,
		config.Config.Logging.Sentry, config.Config.Logging.Syslog)
	Logger.SetFormatter(&logrus.JSONFormatter{})
	Logger.SetReportCaller(true)
	if err != nil {
		return
	}
}

func GuardCritical(msg string, err error) {
	if err != nil {
		fmt.Printf("CRITICAL: %s: %v\n", msg, err)
		os.Exit(-1)
	}
}

func NewLogger(level, facility, tag string, sentry string, syslogAddr string) (*logrus.Logger, error) {
	l := logrus.New()

	fmt.Println("Log leven is ", level)
	ll, ok := logLevel[level]
	if !ok {
		fmt.Println("Unsupported loglevel, falling back to debug!")
		ll = logLevel["debug"]
	}
	l.Level = ll

	if sentry != "" {
		hostname, err := os.Hostname()
		GuardCritical("determining hostname failed", err)

		tags := map[string]string{
			"tag":      tag,
			"hostname": hostname,
		}

		sentryLevels := []logrus.Level{
			logrus.PanicLevel,
			logrus.FatalLevel,
			logrus.ErrorLevel,
		}
		sentHook, err := logrus_sentry.NewWithTagsSentryHook(sentry, tags, sentryLevels)
		GuardCritical("configuring sentry failed", err)

		l.Hooks.Add(sentHook)
	}

	if syslogAddr != "" {
		lf, ok := facilityLevel[facility]
		if !ok {
			fmt.Println("Unsupported log facility, falling back to local0")
			lf = facilityLevel["local0"]
		}
		sysHook, err := logrus_syslog.NewSyslogHook("udp", syslogAddr, lf, tag)
		if err != nil {
			return l, err
		}
		l.Hooks.Add(sysHook)
	}
	return l, nil
}

func BuildLogEntry(l *logrus.Entry, in map[string]string) *logrus.Entry {
	for k, v := range in {
		l = l.WithField(k, v)
	}
	return l
}
func UuidLog(logLevel, uuid, message string) {
	if logLevel == "Err" {
		Logger.WithField("uuid", uuid).Error(message)
	} else if logLevel == "Info" {
		Logger.WithField("uuid", uuid).Info(message)
	} else {
		Logger.WithField("uuid", uuid).Debug(message)
	}
}

func UuidInboundLog(logLevel, uuid, message string) {
	if unQuoteMsg, err := strconv.Unquote(message); err == nil {
		message = unQuoteMsg
	}
	if logLevel == "Err" {
		Logger.WithField("uuid", uuid).WithField("direction", "inbound").Error(message)
	} else if logLevel == "Info" {
		Logger.WithField("uuid", uuid).WithField("direction", "inbound").Info(message)
	} else {
		Logger.WithField("uuid", uuid).WithField("direction", "inbound").Debug(message)
	}
}

func UuidHttpLog(uuid string, resp *resty.Response) {
	if resp != nil {
		ti := resp.Request.TraceInfo()
		Logger.WithField("uuid", uuid).WithField("Status", resp.Status()).
			WithField("  DNSLookup     :", ti.DNSLookup).
			WithField("  ConnTime      :", ti.ConnTime).
			WithField("  TCPConnTime   :", ti.TCPConnTime).
			WithField("  TLSHandshake  :", ti.TLSHandshake).
			WithField("  ServerTime    :", ti.ServerTime).
			WithField("  ResponseTime  :", ti.ResponseTime).
			WithField("  TotalTime     :", ti.TotalTime).
			WithField("  IsConnReused  :", ti.IsConnReused).
			WithField("  IsConnWasIdle :", ti.IsConnWasIdle).
			WithField("  ConnIdleTime  :", ti.ConnIdleTime).
			WithField("  RequestAttempt:", ti.RequestAttempt).
			//WithField("  RemoteAddr    :", ti.RemoteAddr.String()).
			Info("Http Response Received")
	}
}
