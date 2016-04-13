package run

import (
	"errors"
	"net/http"
	"os"
	"strconv"
	"time"

	"gopkg.in/tylerb/graceful.v1"

	"github.com/Sirupsen/logrus"
	"github.com/solher/snakepit/root"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	Port    = "app.port"
	Timeout = "app.timeout"
)

var (
	port    int
	timeout time.Duration
	Logger  = logrus.New()
)

var Builder func(v *viper.Viper, l *logrus.Logger) http.Handler

var Cmd = &cobra.Command{
	Use:   "run",
	Short: "Runs the service",
	RunE: func(cmd *cobra.Command, args []string) error {
		if Builder == nil {
			return errors.New("nil builder func")
		}

		Logger.Info("Building...")
		appHandler := Builder(root.Viper, Logger)

		Logger.Infof("Listening on port %d.", port)
		graceful.Run(":"+strconv.Itoa(port), timeout, appHandler)
		return nil
	},
}

func init() {
	Logger.Formatter = &logrus.TextFormatter{}
	Logger.Out = os.Stdout
	Logger.Level = logrus.DebugLevel

	Cmd.PersistentFlags().IntVarP(&port, "port", "p", 3000, "listening port")
	root.Viper.BindPFlag(Port, Cmd.PersistentFlags().Lookup("port"))

	Cmd.PersistentFlags().DurationVar(&timeout, "timeout", 10*time.Second, "graceful shutdown timeout (0 for infinite)")
	root.Viper.BindPFlag(Timeout, Cmd.PersistentFlags().Lookup("timeout"))
}
