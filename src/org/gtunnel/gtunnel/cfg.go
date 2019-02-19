package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
	"strconv"
	"org/gtunnel/api"
)

var (
	lineExp   = regexp.MustCompile(`([^ ]*)[ ]+=[ ]+([^ ]+)`)
)

type Cfg struct {
	Name    string
	Accept  *api.Endpoint
	Connect *api.EndpointList

	KV map[string]string
}

func mkCfg(name string) *Cfg {
	return &Cfg {
		Name: name,
		KV:   make(map[string]string),
	}
}

func (cfg Cfg) String() string {
	return fmt.Sprintf("%s %s", cfg.Name, cfg.KV["connect"])
}

func (cfg *Cfg) Get(k string) string {
	v, ok := cfg.KV[k]
	if !ok {
		return ""
	}
	return v
}

func (cfg *Cfg) Valid() bool {
	if cfg.Name == "" {
		return false
	}
	return !cfg.Accept.SSL ||
		(cfg.KV["Cert"] != "" && cfg.KV["Key"] != "")
}

func (cfg *Cfg) Timeout(aTime time.Time) bool {
	s := cfg.KV["TimeoutIdle"]
	if s == "" {
		return false
	}
	d, _ := strconv.ParseInt(s, 10, 64)
	return api.Due(aTime.Add(time.Duration(d)))
}

func (cfg *Cfg) SkipVerify() bool {
	if val, ok := cfg.KV["skip-verify"]; ok {
		return val == "true"
	}
	return false
}

func (cfg *Cfg) AddIfMiss(other *Cfg) {
	if cfg.KV["cert"] == "" {
		cfg.KV["cert"] = other.KV["cert"]
	}
	if cfg.KV["key"] == "" {
		cfg.KV["key"] = other.KV["key"]
	}
	if cfg.KV["timeout-idle"] == "" {
		cfg.KV["timeout-idle"] = other.KV["timeout-idle"]
	}
	if cfg.KV["skip-verify"] == "" {
		cfg.KV["skip-verify"] = other.KV["skip-verify"]
	}
}

type Configuration struct {
	Set []*Cfg
}

func (configuration *Configuration) append(cfg *Cfg) error {
	r := strings.Split(cfg.KV["connect"], "/")
	api.Assert(len(r) == 2 || len(r) == 3, cfg.KV["connect"])

	if len(r) == 2 {
		srcEp, dstEp, err := api.Solve("pp", r[0], r[1])
		if err != nil {
			return err
		}
		cfg.Accept = srcEp
		cfg.Connect = dstEp
	} else if len(r) == 3 {
		srcEp, dstEp, err := api.Solve(r[0], r[1], r[2])
		if err != nil {
			return err
		}
		cfg.Accept = srcEp
		cfg.Connect = dstEp
	}

	configuration.Set = append(configuration.Set, cfg)
	return nil
}

func (configuration *Configuration) Load(files []string) error {
	for _, f := range files {
		fmt.Printf("loading %s\n", f)

		err := configuration.LoadFile(f)
		if err != nil {
			return err
		}
	}
	return nil
}

func (configuration *Configuration) LoadFile(file string) error {
	cfgFile, err := os.Open(file)
	api.Assert(err == nil, fmt.Sprintf("failed to load %s %s", file, err))
	defer cfgFile.Close()

	var shared, cur *Cfg = mkCfg(""), nil

	for scanner := bufio.NewScanner(cfgFile); scanner.Scan(); {
		plainLine := strings.Trim(scanner.Text(), " ")
		line := strings.Trim(plainLine, " ")

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		} else if strings.HasPrefix(line, "[") {
			if !strings.HasSuffix(line, "]") {
				return errors.New(fmt.Sprintf("bad format: %s", plainLine))
			}
			if cur != nil {
				cur.AddIfMiss(shared)
				if err := configuration.append(cur); err != nil {
					return err
				}
			}

			cur = mkCfg(strings.Trim(line, "[] "))
		} else {
			r := lineExp.FindStringSubmatch(line)
			if len(r) != 3 {
				continue
			}

			if cur != nil {
				cur.KV[r[1]] = r[2]
			} else {
				shared.KV[r[1]] = r[2]
			}
		}
	}

	if cur != nil {
		cur.AddIfMiss(shared)
		if err := configuration.append(cur); err != nil {
			return err
		}
	}
	return nil
}
