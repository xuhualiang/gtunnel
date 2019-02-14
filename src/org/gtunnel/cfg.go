package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"
	"time"
	"strconv"
)

var (
	lineExp = regexp.MustCompile(`([^ ]*)[ ]+=[ ]+([^ ]+)`)
	wireExp = regexp.MustCompile(`(([ps]{2})/)?([a-zA-Z0-9]+:[0-9]*)/([a-zA-Z0-9]+:[0-9]*)`)
)

type Endpoint struct {
	Addr   *net.TCPAddr
	SSL    bool
	Verify bool
}

func (ep Endpoint) String() string {
	if ep.SSL {
		return fmt.Sprintf("%s", ep.Addr)
	} else {
		return fmt.Sprintf("%s", ep.Addr)
	}
}

type Cfg struct {
	Name    string
	Accept  Endpoint
	Connect Endpoint

	TimeoutIdle time.Duration

	Cert string
	Key  string
}

func (cfg Cfg) String() string {
	return fmt.Sprintf("%s %s/%s", cfg.Name, cfg.Accept, cfg.Connect)
}

func (cfg *Cfg) Valid() bool {
	return !cfg.Accept.SSL || (cfg.Cert != "" && cfg.Key != "")
}

func (cfg *Cfg) Solve(prot string, src string, dst string) error {
	if strings.Trim(prot, " ") == "" {
		prot = "pp"
	}
	srcEp, err := solveEndpoint(src, prot[0] == 's')
	if err != nil {
		return err
	}
	dstEp, err := solveEndpoint(dst, prot[1] == 's')
	if err != nil {
		return err
	}
	cfg.Accept = *srcEp
	cfg.Connect = *dstEp
	return nil
}

func (cfg *Cfg) Timeout(atime time.Time) bool {
	return cfg.TimeoutIdle != 0 && time.Now().After(atime.Add(cfg.TimeoutIdle))
}

func assert(b bool, msg string) {
	if !b {
		panic(msg)
	}
}

func solveEndpoint(s string, ssl bool) (*Endpoint, error) {
	addr, err := net.ResolveTCPAddr("tcp", s)
	if err != nil {
		return nil, err
	}
	return &Endpoint{Addr: addr, SSL: ssl}, nil
}

type Configuration struct {
	Set []Cfg
}

func (configuration *Configuration) append(cfg Cfg, shared Cfg) {
	if cfg.Accept.SSL {
		if cfg.Cert == "" {
			cfg.Cert = shared.Cert
		}
		if cfg.Key == "" {
			cfg.Key = shared.Key
		}
	}
	if cfg.TimeoutIdle == 0 {
		cfg.TimeoutIdle = shared.TimeoutIdle
	}
	if cfg.Valid() {
		configuration.Set = append(configuration.Set, cfg)
	}
}

func (configuration *Configuration) Load(files []string) error {
	for _, f := range files {
		err := configuration.LoadFile(f)
		if err != nil {
			return err
		}
	}
	return nil
}

func (configuration *Configuration) LoadFile(file string) error {
	cfgFile, err := os.Open(file)
	assert(err == nil, fmt.Sprintf("failed to load %s %s", file, err))
	defer cfgFile.Close()

	shared, cur, inSection := Cfg{}, Cfg{}, false

	for scanner := bufio.NewScanner(cfgFile); scanner.Scan(); {
		plainLine := strings.Trim(scanner.Text(), " ")
		line := strings.Trim(plainLine, " ")

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		} else if strings.HasPrefix(line, "[") {
			if !strings.HasSuffix(line, "]") {
				return errors.New(fmt.Sprintf("bad format: %s", plainLine))
			}
			if inSection {
				configuration.append(cur, shared)
			}

			cur = Cfg{Name: strings.Trim(line, "[] ")}
			inSection = true
		} else {
			r := lineExp.FindStringSubmatch(line)
			if len(r) != 3 {
				continue
			}

			if inSection {
				switch r[1] {
				case "cert":
					cur.Cert = r[2]
				case "key":
					cur.Key = r[2]
				case "connect":
					r = wireExp.FindStringSubmatch(r[2])
					assert(len(r) != 3 || len(r) != 5, "")
					if len(r) == 3 {
						err := cur.Solve("pp", r[1], r[2])
						if err != nil {
							return err
						}
					} else if len(r) == 5 {
						err := cur.Solve(r[2], r[3], r[4])
						if err != nil {
							return err
						}
					}
				case "timeout-idle":
					if v, err := strconv.ParseUint(r[2], 10, 64); err != nil {
						cur.TimeoutIdle = time.Duration(v)
					}

				default:
					return errors.New(fmt.Sprintf("bad format: %s", plainLine))
				}
			} else {
				switch r[1] {
				case "cert":
					shared.Cert = r[2]
				case "key":
					shared.Key = r[2]
				case "timeout-idle":
					if v, err := strconv.ParseUint(r[2], 10, 64); err != nil {
						shared.TimeoutIdle = time.Duration(v)
					}
				default:
					return errors.New(fmt.Sprintf("bad format: %s", plainLine))
				}
			}
		}
	}

	if inSection {
		configuration.append(cur, shared)
	}
	return nil
}
