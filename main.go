package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type imapMetrics struct {
	timeout                    time.Duration
	tls                        bool
	tlsConfig                  *tls.Config
	host, port, user, password string
	filter                     string
}

type context struct {
	listen      string
	imapMetrics []*imapMetrics
}

const (
	namespace = "imap_messages"
)

var (
	imapMessagesServerOk = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "mailbox_ok",
		Help:      "1 if the server is OK",
	}, []string{"server", "user"})
	imapMessagesTotalCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "total_count",
		Help:      "Number of messages",
	}, []string{"server", "user", "mailbox"})
	imapMessagesUnreadCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "unread_count",
		Help:      "Number of unread messages",
	}, []string{"server", "user", "mailbox"})
)

func errorToPromResult(err error) float64 {
	if err == nil {
		return 1
	}
	return 0
}

func (c *context) Describe(ch chan<- *prometheus.Desc) {
	imapMessagesServerOk.Describe(ch)
	imapMessagesTotalCount.Describe(ch)
	imapMessagesUnreadCount.Describe(ch)
}

func (c *context) Collect(ch chan<- prometheus.Metric) {
	errors := make(chan error)
	for _, metric := range c.imapMetrics {
		go func(metric *imapMetrics) {
			err := metric.collect(ch)
			result := imapMessagesServerOk.WithLabelValues(metric.host, metric.user)
			result.Set(errorToPromResult(err))
			result.Collect(ch)
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
	c, err := i.dial()
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
		total := imapMessagesTotalCount.WithLabelValues(i.host, i.user, m.Name)
		total.Set(float64(mbox.Messages))
		total.Collect(ch)
		unread := imapMessagesUnreadCount.WithLabelValues(i.host, i.user, m.Name)
		unread.Set(float64(mbox.Unseen))
		unread.Collect(ch)
	}
	return err
}

func (i *imapMetrics) disconnect(client *client.Client) {
	if err := client.Logout(); err != nil {
		log.Println("Could not logout: ", err)
	}
}

func (i *imapMetrics) dial() (*client.Client, error) {
	dialer := &net.Dialer{
		Timeout: i.timeout,
	}
	hostport := fmt.Sprintf("%s:%s", i.host, i.port)
	if i.tls {
		return client.DialWithDialerTLS(dialer, hostport, i.tlsConfig)
	}
	return client.DialWithDialer(dialer, hostport)
}

func getContext(filename string, client *http.Client) context {
	jsonConfig := NewConfig(filename)

	context := context{
		listen: jsonConfig.Listen,
	}

	for _, account := range jsonConfig.Accounts {
		metric := &imapMetrics{
			filter:  account.Filter,
			timeout: time.Duration(jsonConfig.TimeoutMs) * time.Millisecond,
		}
		if metric.filter == "" {
			metric.filter = "*"
		}
		parsedURL, err := url.Parse(account.URL)
		if err != nil {
			log.Fatalf("Cannot parse URL %s: %s", account.URL, err)
		}

		if parsedURL.Opaque != "" || parsedURL.Path != "" || parsedURL.RawQuery != "" || parsedURL.Fragment != "" {
			log.Fatalf("Wrong URL: %s", account.URL)
		}

		if parsedURL.Scheme == "imap" {
			metric.tls = false
		} else if parsedURL.Scheme == "imaps" {
			metric.tls = true
		} else {
			log.Fatalf("Unknown scheme: %s", parsedURL.Scheme)
		}

		hostport := strings.SplitN(parsedURL.Host, ":", 2)
		metric.host = hostport[0]
		if len(hostport) == 2 {
			metric.port = hostport[1]
		} else {
			if metric.tls {
				metric.port += "993"
			} else {
				metric.port += "143"
			}
		}

		if parsedURL.User == nil {
			log.Fatalln("No user/password")
		}
		metric.user = parsedURL.User.Username()
		metric.password, _ = parsedURL.User.Password()

		if account.SkipTLSValidation {
			metric.tlsConfig = &tls.Config{InsecureSkipVerify: true}
		} else if account.Pem != "" {
			roots := x509.NewCertPool()
			ok := roots.AppendCertsFromPEM([]byte(account.Pem))
			if !ok {
				log.Fatalf("failed to parse root certificate %s", account.Pem)
			}
			metric.tlsConfig = &tls.Config{RootCAs: roots}
		}

		context.imapMetrics = append(context.imapMetrics, metric)
	}

	return context
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Usage", os.Args[0], "<config_file>")
		os.Exit(1)
	}

	client := &http.Client{
		Timeout: time.Second * 10,
	}
	context := getContext(os.Args[1], client)

	prometheus.MustRegister(&context)
	http.Handle("/metrics", promhttp.Handler())
	log.Println(http.ListenAndServe(context.listen, nil))
}
