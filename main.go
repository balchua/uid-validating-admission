package main

import (
	"os"

	"github.com/balchua/uid-validating-webhook/server"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

type Config struct {
	ListenOn string `default:"0.0.0.0:8080"`
	TlsCert  string `default:"/etc/webhook/certs/cert.pem"`
	TlsKey   string `default:"/etc/webhook/certs/key.pem"`
	Debug    bool   `default:"true"`
}

func init() {
	// Log as JSON instead of the default ASCII formatter.
	logrus.SetFormatter(&logrus.JSONFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	logrus.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	logrus.SetLevel(logrus.DebugLevel)
}
func main() {
	config := &Config{}
	envconfig.Process("", config)

	if config.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	logrus.Infoln(config)
	runAsAC := server.RunAsUserAdmission{}
	s := server.GetAdmissionServerNoSSL(&runAsAC, config.ListenOn)
	//s := server.GetAdmissionValidationServer(&runAsAC, config.TlsCert, config.TlsKey, config.ListenOn)
	//s.ListenAndServeTLS("", "")
	s.ListenAndServe()
}
