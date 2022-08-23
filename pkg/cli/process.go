package cli

import (
	"context"
	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"log"
	"os"
	"os/signal"
	"podchaosmonkey/pkg/chaosmonkey"
	"podchaosmonkey/pkg/kubeclient"
	"syscall"
)

type config struct {
	KubeConfig  string `envconfig:"KUBE_CONFIG" default:""`
	Labels      string `envconfig:"LABELS" default:""`
	Namespace   string `envconfig:"NAME_SPACE" default:"default"`
	Schedule    string `envconfig:"SCHEDULE" default:"10s"`
	SelfPodName string `envconfig:"MY_POD_NAME" default:""`
}

func (o *config) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.Labels, "label-selector", "l", o.Labels, "Label selector with `kubectl label ...` syntax")
	fs.StringVarP(&o.KubeConfig, "kube-config", "k", o.KubeConfig, "path to kube config file")
	fs.StringVarP(&o.Namespace, "namespace", "n", o.Namespace, "namespace to monitor and delete pods")
	fs.StringVarP(&o.Schedule, "schedule", "s", o.Schedule, "schedule to run the process, in duration string format e.g 10s, 1h")
}
func NewProcessCMD() (*cobra.Command, error) {
	runtimeConfig := &config{}
	err := envconfig.Process("", runtimeConfig)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	cmd := &cobra.Command{
		Use: "run",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := kubeclient.Create(runtimeConfig.KubeConfig)
			if err != nil {
				return err
			}

			wl, err := chaosmonkey.NewWorkload(runtimeConfig.Namespace, runtimeConfig.Schedule, runtimeConfig.SelfPodName, chaosmonkey.SelectConfig{
				Labels: runtimeConfig.Labels,
			}, client)
			if err != nil {
				return err
			}

			errChan := wl.Start(ctx)
			go func() {
				for {
					select {
					case err = <-errChan:
						log.Println(err)
					}
				}
			}()

			done := make(chan bool, 1)

			SetupCloseHandler(cancel, done)
			return nil
		},
	}
	runtimeConfig.AddFlags(cmd.PersistentFlags())
	log.Printf("running with config %+v\n", *runtimeConfig)
	return cmd, nil

}

func SetupCloseHandler(cancelFunc func(), done chan bool) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cancelFunc()
		log.Println("SIGTERM received")
		done <- true
	}()
	<-done
	log.Println("exiting")
}
