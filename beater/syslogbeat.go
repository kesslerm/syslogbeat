package beater

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/libbeat/publisher/bc/publisher"

	"github.com/digitalocean/captainslog"

	"github.com/kesslerm/syslogbeat/config"
)

type Syslogbeat struct {
	done   chan struct{}
	config config.Config
	client publisher.Client
	server *Server
}

// Creates beater
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	config := config.DefaultConfig
	if err := cfg.Unpack(&config); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}

	location, err := time.LoadLocation(config.Location)
	if err != nil {
		return nil, err
	}

	bt := &Syslogbeat{
		done:   make(chan struct{}),
		config: config,
		server: NewServer(
			OptionLocation(location),
			OptionUDPAddr(config.UDPAddr),
		),
	}
	return bt, nil
}

func (bt *Syslogbeat) Run(b *beat.Beat) error {
	logp.Info("syslogbeat is running! Hit CTRL-C to stop it.")

	bt.server.Start()

	bt.client = b.Publisher.Connect()
	for {
        var m captainslog.SyslogMsg
		select {
		case <-bt.done:
			return nil
        case m = <-bt.server.q:
		}

		event := common.MapStr{
			"@timestamp": time.Now(),
			"type":       b.Info.Beat,
			"logsource": m.Host,
			"message": strings.TrimSpace(m.Content),
			"priority": m.Pri.Priority,
			"facility": m.Pri.Facility,
			"facility_label": m.Pri.Facility.String(),
			"severity": m.Pri.Severity,
			"severity_label": m.Pri.Severity.String(),
			"time": m.Time,
			"program": m.Tag.Program,
		}

		if pid, _ := strconv.Atoi(m.Tag.Pid); pid > 0 {
			event.Put("pid", pid)
		}

		if m.IsCee {
			event.Put("cee", true)
		}

		if m.IsJSON {
			event.Put("json", m.JSONValues)
		}

		bt.client.PublishEvent(event)
		logp.Info("Event sent")
	}
}

func (bt *Syslogbeat) Stop() {
    bt.server.Stop()
	bt.client.Close()
	close(bt.done)
}
