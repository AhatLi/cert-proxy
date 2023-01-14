package config

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strconv"

	"gopkg.in/ini.v1"
)

var cfg *Config = nil

func GetConfig() *Config {
	if cfg == nil {
		InitConfig()
	}
	return cfg
}

type Config struct {
	Https_certfile               string
	Https_keyfile                string
	ReadTimeout                  int
	WriteTimeout                 int
	Access_control_allow_origin  string
	Access_control_allow_headers string
	Network_list                 map[string]string
	UseInsecureSkipVerify        bool
}

func (conf *Config) PrintJson() string {
	s, _ := json.Marshal(conf)

	return string(s)
}

func InitConfig() error {
	ctype := os.Getenv("IAM_CONFIG_TYPE")

	if cfg == nil {
		cfg = new(Config)
	}

	var err error
	if ctype == "env" {
		err = cfg.initEnvConfig()
	} else {
		err = cfg.initConf()
	}

	if cfg.UseInsecureSkipVerify {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	return err
}

func (conf *Config) initConf() error {
	cfg, err := ini.Load("config.conf")
	if err != nil {
		return err
	}
	conf.Https_certfile = cfg.Section("certificate").Key("ssl_certfile").String()
	conf.Https_keyfile = cfg.Section("certificate").Key("ssl_keyfile").String()
	if conf.Https_certfile == "" || conf.Https_keyfile == "" {
		return errors.New("Certificate config is required.")
	}

	conf.Network_list = map[string]string{}
	count := 1
	for {
		port := cfg.Section("network").Key("proxy_in_port_" + strconv.Itoa(count)).String()
		url := cfg.Section("network").Key("proxy_out_url_" + strconv.Itoa(count)).String()

		if port == "" || url == "" {
			break
		}

		conf.Network_list[port] = url
		count++
	}

	conf.ReadTimeout, err = cfg.Section("network").Key("read_timeout").Int()
	if err != nil {
		conf.ReadTimeout = 9999
	}
	conf.WriteTimeout, err = cfg.Section("network").Key("write_timeout").Int()
	if err != nil {
		conf.WriteTimeout = 9999
	}
	conf.Access_control_allow_origin = cfg.Section("network").Key("access_control_allow_origin").String()
	if conf.Access_control_allow_origin == "" {
		conf.Access_control_allow_origin = "*"
	}
	conf.Access_control_allow_headers = cfg.Section("network").Key("access_control_allow_headers").String()
	if conf.Access_control_allow_headers == "" {
		conf.Access_control_allow_headers = "*"
	}

	conf.UseInsecureSkipVerify = cfg.Section("network").Key("UseInsecureSkipVerify").MustBool()

	return nil
}

func (conf *Config) initEnvConfig() error {

	conf.Https_certfile = os.Getenv("CERTIFICATE_SSL_CERTFILE")
	conf.Https_keyfile = os.Getenv("CERTIFICATE_SSL_KEYFILE")
	if conf.Https_certfile == "" || conf.Https_keyfile == "" {
		return errors.New("Certificate config is required.")
	}

	count := 1
	for {
		port := os.Getenv("NETWORK_PROXY_IN_PORT_" + strconv.Itoa(count))
		url := os.Getenv("NETWORK_PROXY_IN_URL_" + strconv.Itoa(count))

		if port == "" || url == "" {
			break
		}

		conf.Network_list[port] = url
		count++
	}

	var err error
	conf.ReadTimeout, err = strconv.Atoi(os.Getenv("NETWROK_READ_TIMEOUT"))
	if err != nil {
		conf.ReadTimeout = 9999
	}

	conf.WriteTimeout, err = strconv.Atoi(os.Getenv("NETWROK_WRITE_TIMEOUT"))
	if err != nil {
		conf.WriteTimeout = 9999
	}

	conf.Access_control_allow_origin = os.Getenv("ACCESS_CONRTOL_ALLOW_ORIGIN")
	conf.Access_control_allow_headers = os.Getenv("ACCESS_CONRTOL_ALLOW_HEADERS")

	if conf.Access_control_allow_origin == "" {
		conf.Access_control_allow_origin = "*"
	}

	if conf.Access_control_allow_headers == "" {
		conf.Access_control_allow_headers = "*"
	}

	conf.UseInsecureSkipVerify = false
	if os.Getenv("USE_INSECURE_SKIP_VERIFY") == "true" {
		conf.UseInsecureSkipVerify = true
	}

	return nil
}
