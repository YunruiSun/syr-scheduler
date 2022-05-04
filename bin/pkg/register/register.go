package register

import (
	"github.com/YunruiSun/syr-scheduler/pkg/mybalancedallocation"
	"github.com/spf13/cobra"
	"k8s.io/kubernetes/cmd/kube-scheduler/app"
)

func Register() *cobra.Command {
	return app.NewSchedulerCommand(
		app.WithPlugin(mybalancedallocation.Name, mybalancedallocation.New),
	)
}
