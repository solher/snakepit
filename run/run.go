package run

import (
	"net/http"
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
)

var Builder func(v *viper.Viper, l *logrus.Logger) http.Handler

var Cmd = &cobra.Command{
	Use:   "run",
	Short: "Runs the service",
	RunE: func(cmd *cobra.Command, args []string) error {
		if Builder == nil {
			panic("nil builder func")
		}

		root.Logger.Info("Initializing...")
		appHandler := Builder(root.Viper, root.Logger)

		root.Logger.Infof("Listening on :%d.", port)
		graceful.Run(":"+strconv.Itoa(port), timeout, appHandler)
		return nil
	},
}

func init() {
	Cmd.PersistentFlags().IntVarP(&port, "port", "p", 3000, "listening port")
	root.Viper.BindPFlag(Port, Cmd.PersistentFlags().Lookup("port"))

	Cmd.PersistentFlags().DurationVar(&timeout, "timeout", 10*time.Second, "graceful shutdown timeout (0 for infinite)")
	root.Viper.BindPFlag(Timeout, Cmd.PersistentFlags().Lookup("timeout"))
}
