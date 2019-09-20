package main

import (
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"strings"
)

type VerifyOpt struct {
	DoVerify   bool
	RootCA     *x509.CertPool
	ServerName string
}

/* flag.Value interface */
func (opt *VerifyOpt) Set(s string) error {
	ss := strings.Split(s, ":")
	opt.DoVerify = true

	for _, subOpt := range ss {
		if subOpt == "" {
			continue
		}

		kv := strings.Split(subOpt, "=")
		if len(kv) != 2 {
			continue
		}

		switch strings.ToLower(kv[0]) {
		case "ca", "root", "rootca":
			pemCerts, err := ioutil.ReadFile(kv[1])
			if err != nil {
				return err
			}

			opt.RootCA = x509.NewCertPool()
			opt.RootCA.AppendCertsFromPEM(pemCerts)

		case "servername", "server-name":
			opt.ServerName = kv[1]
		}
	}

	return nil
}

func (opt VerifyOpt) String() string {
	return fmt.Sprintf("%t %p %s\n", opt.DoVerify, opt.RootCA, opt.ServerName)
}