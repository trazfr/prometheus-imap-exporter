package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"log"
	"net/url"
	"os"
	"strings"
	"time"
)

type Account struct {
	Filter    string
	TLSConfig *tls.Config
	URL       *url.URL
}

type Config struct {
	Listen   string
	Timeout  time.Duration
	Accounts []Account
}

type internalAccount struct {
	Filter            string `json:"filter"`
	SkipTLSValidation bool   `json:"skip_tls_validation"`
	Pem               string `json:"pem"`
	URL               string `json:"url"`
}

type internalConfig struct {
	Listen         string            `json:"listen"`
	TimeoutSeconds float64           `json:"timeout"`
	Accounts       []internalAccount `json:"accounts"`
}

func NewConfig(filename string) Config {
	fd, err := os.Open(filename)
	if err != nil {
		log.Fatalln(err)
	}
	defer fd.Close()

	internalConfig := internalConfig{
		Listen:         ":9091",
		TimeoutSeconds: 5,
	}
	if err := json.NewDecoder(fd).Decode(&internalConfig); err != nil {
		log.Fatalln(err)
	}

	config := Config{
		Listen:  internalConfig.Listen,
		Timeout: time.Duration(internalConfig.TimeoutSeconds * float64(time.Second)),
	}
	for _, internalAccount := range internalConfig.Accounts {
		config.Accounts = append(config.Accounts, Account{
			Filter: internalAccount.Filter,
		})
		account := &config.Accounts[len(config.Accounts)-1]

		if account.Filter == "" {
			account.Filter = "*"
		}

		parsedURL, err := url.Parse(internalAccount.URL)
		if err != nil {
			log.Fatalf("Cannot parse URL %s: %s", account.URL, err)
		}

		if parsedURL.Scheme != "imap" && parsedURL.Scheme != "imaps" {
			log.Fatalf("Unknown scheme: %s", parsedURL.Scheme)
		}
		if !strings.Contains(parsedURL.Host, ":") {
			if parsedURL.Scheme == "imaps" {
				parsedURL.Host += ":993"
			} else {
				parsedURL.Host += ":143"
			}
		}
		if parsedURL.User == nil {
			log.Fatalln("No user/password")
		}

		if parsedURL.Opaque != "" || parsedURL.Path != "" || parsedURL.RawQuery != "" || parsedURL.Fragment != "" {
			log.Fatalf("Wrong URL: %s", account.URL)
		}

		if internalAccount.SkipTLSValidation || internalAccount.Pem != "" {
			account.TLSConfig = &tls.Config{}
			if internalAccount.SkipTLSValidation {
				account.TLSConfig.InsecureSkipVerify = true
			} else if internalAccount.Pem != "" {
				roots := x509.NewCertPool()
				ok := roots.AppendCertsFromPEM([]byte(internalAccount.Pem))
				if !ok {
					log.Fatalf("failed to parse root certificate %s", internalAccount.Pem)
				}
				account.TLSConfig.RootCAs = roots
			}
		}
		account.URL = parsedURL
	}

	return config
}
