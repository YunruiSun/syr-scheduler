package register

import (
	"github.com/spf13/cobra"
	"k8s.io/kubernetes/cmd/kube-scheduler/app"
	"syr-scheduler-test-2/pkg/mybalancedallocation"
)

func Register() *cobra.Command {
	return app.NewSchedulerCommand(
		app.WithPlugin(mybalancedallocation.Name, mybalancedallocation.New),
	)
}
