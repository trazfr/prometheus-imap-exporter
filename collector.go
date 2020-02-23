package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "imap_messages"
)

var (
	promDescImapMessagesServerOkDesc = prometheus.NewDesc(
		namespace+"_mailbox_ok",
		"1 if the server is OK.",
		[]string{"server", "user"}, nil)
	promDescImapMessagesTotalCount = prometheus.NewDesc(
		namespace+"_total",
		"Number of messages",
		[]string{"server", "user", "mailbox"}, nil)
	promDescImapMessagesUnreadCount = prometheus.NewDesc(
		namespace+"_unread_total",
		"Number of unread messages",
		[]string{"server", "user", "mailbox"}, nil)
)

type imapDialer interface {
	dial() (*client.Client, error)
}

type imapDialerPlainText struct {
	hostport string
	timeout  time.Duration
}

type imapDialerTLS struct {
	imapDialerPlainText
	tlsConfig *tls.Config
}

type imapMetrics struct {
	timeout              time.Duration
	dialer               imapDialer
	host, user, password string
	filter               string
	promCounterOk        prometheus.Counter
}

type Collector struct {
	imapMetrics []*imapMetrics
	promCounter *prometheus.CounterVec
}

func errorToPromResult(err error) float64 {
	if err == nil {
		return 1
	}
	return 0
}

func errorToString(err error) string {
	if err == nil {
		return "ok"
	}
	return "ko"
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	c.promCounter.Describe(ch)
	ch <- promDescImapMessagesServerOkDesc
	ch <- promDescImapMessagesTotalCount
	ch <- promDescImapMessagesUnreadCount
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	errors := make(chan error)
	for _, metric := range c.imapMetrics {
		go func(metric *imapMetrics) {
			err := metric.collect(ch)

			ch <- prometheus.MustNewConstMetric(promDescImapMessagesServerOkDesc, prometheus.GaugeValue,
				errorToPromResult(err),
				metric.host, metric.user)

			res := c.promCounter.WithLabelValues(metric.host, errorToString(err))
			res.Inc()
			res.Collect(ch)

			errors <- err
		}(metric)
	}
	for range c.imapMetrics {
		if err := <-errors; err != nil {
			log.Println(err)
		}
	}
}

func (i *imapMetrics) collect(ch chan<- prometheus.Metric) error {
	c, err := i.dialer.dial()
	if err != nil {
		return fmt.Errorf("Error while dialing %s: %s", i.host, err)
	}
	defer i.disconnect(c)

	if err := c.Login(i.user, i.password); err != nil {
		return fmt.Errorf("Error while logging into %s: %s", i.host, err)
	}

	mailboxes := make(chan *imap.MailboxInfo)
	done := make(chan error)
	go func() {
		done <- c.List("", i.filter, mailboxes)
	}()

	mailboxesList := []*imap.MailboxInfo{}
	for m := range mailboxes {
		mailboxesList = append(mailboxesList, m)
	}
	if err := <-done; err != nil {
		return fmt.Errorf("Error while fetching the mailboxes on %s: %s", i.host, err)
	}

	for _, m := range mailboxesList {
		mbox, err := c.Status(m.Name, []imap.StatusItem{imap.StatusMessages, imap.StatusUnseen})
		if err != nil {
			err = fmt.Errorf("Error while fetching mailbox %s on %s: %s", mbox.Name, i.host, err)
		}
		ch <- prometheus.MustNewConstMetric(promDescImapMessagesTotalCount, prometheus.GaugeValue,
			float64(mbox.Messages),
			i.host, i.user, m.Name)
		ch <- prometheus.MustNewConstMetric(promDescImapMessagesUnreadCount, prometheus.GaugeValue,
			float64(mbox.Unseen),
			i.host, i.user, m.Name)
	}
	return err
}

func (i *imapMetrics) disconnect(client *client.Client) {
	if err := client.Logout(); err != nil {
		log.Println("Could not logout:", err)
	}
	client.Terminate()
}

func (i *imapDialerPlainText) dial() (*client.Client, error) {
	dialer := net.Dialer{
		Timeout: i.timeout,
	}
	return client.DialWithDialer(&dialer, i.hostport)
}

func (i *imapDialerTLS) dial() (*client.Client, error) {
	dialer := net.Dialer{
		Timeout: i.timeout,
	}
	return client.DialWithDialerTLS(&dialer, i.hostport, i.tlsConfig)
}

func NewCollector(config *Config, client *http.Client) Collector {
	collector := Collector{
		promCounter: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "fetch_total",
			Help:      "Number of times the fetch performed.",
		}, []string{"host", "result"}),
	}

	for _, account := range config.Accounts {
		host, _, err := net.SplitHostPort(account.URL.Host)
		if err != nil {
			log.Fatalf("Could not split host/port: %s", err)
		}
		password, _ := account.URL.User.Password()

		var dialer imapDialer
		if account.URL.Scheme == "imaps" {
			dialer = &imapDialerTLS{
				imapDialerPlainText: imapDialerPlainText{
					hostport: account.URL.Host,
					timeout:  config.Timeout,
				},
				tlsConfig: account.TLSConfig,
			}
		} else {
			dialer = &imapDialerPlainText{
				hostport: account.URL.Host,
				timeout:  config.Timeout,
			}
		}

		collector.imapMetrics = append(collector.imapMetrics, &imapMetrics{
			filter:   account.Filter,
			timeout:  config.Timeout,
			dialer:   dialer,
			host:     host,
			user:     account.URL.User.Username(),
			password: password,
		})
	}

	return collector
}
