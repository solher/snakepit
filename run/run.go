package run

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/solher/snakepit/root"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/tylerb/graceful.v1"
)

var (
	port    int
	timeout time.Duration
)

var Builder func(v *viper.Viper) http.Handler

var Cmd = &cobra.Command{
	Use:   "run",
	Short: "Runs the service",
	RunE: func(cmd *cobra.Command, args []string) error {
		if Builder == nil {
			panic("nil builder func")
		}

		fmt.Printf("\nListening on :%d\n", port)
		graceful.Run(":"+strconv.Itoa(port), timeout, Builder(root.Viper))

		return nil
	},
}

func init() {
	root.Viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	Cmd.PersistentFlags().IntVarP(&port, "port", "p", 3000, "listening port")
	root.Viper.BindPFlag("port", Cmd.PersistentFlags().Lookup("port"))

	Cmd.PersistentFlags().DurationVar(&timeout, "timeout", 10*time.Second, "graceful shutdown timeout (0 for infinite)")
	root.Viper.BindPFlag("timeout", Cmd.PersistentFlags().Lookup("timeout"))
}
