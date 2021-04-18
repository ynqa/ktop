package cmd

import (
	"os"
	"os/signal"
	"regexp"
	"syscall"
	"time"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/metrics/pkg/client/clientset/versioned"

	"github.com/gizak/termui/v3"
	"github.com/spf13/cobra"

	"github.com/ynqa/ktop/pkg/dashboard"
	"github.com/ynqa/ktop/pkg/drawer"
	"github.com/ynqa/ktop/pkg/resources"
)

type ktop struct {
	interval       time.Duration
	nodeQuery      string
	podQuery       string
	containerQuery string
	allNamespaces  bool

	kubeFlags *genericclioptions.ConfigFlags
}

func New() *cobra.Command {
	ktop := &ktop{}
	cmd := &cobra.Command{
		Use:   "ktop",
		Short: "Kubernetes monitoring dashboard on terminal",
		RunE:  ktop.run,
	}
	cmd.Flags().DurationVarP(
		&ktop.interval,
		"interval",
		"i",
		1*time.Second,
		"set interval",
	)
	cmd.Flags().StringVarP(
		&ktop.nodeQuery,
		"node-query",
		"N",
		".*",
		"node query",
	)
	cmd.Flags().StringVarP(
		&ktop.podQuery,
		"pod-query",
		"P",
		".*",
		"pod query",
	)
	cmd.Flags().StringVarP(
		&ktop.containerQuery,
		"container-query",
		"C",
		".*",
		"container query",
	)
	cmd.Flags().BoolVarP(
		&ktop.allNamespaces,
		"all-namespaces",
		"A",
		false,
		"If present, list the requested object(s) across all namespaces.",
	)

	ktop.kubeFlags = genericclioptions.NewConfigFlags(false)
	ktop.kubeFlags.AddFlags(cmd.Flags())

	return cmd
}

func (k *ktop) loop(
	clientset *kubernetes.Clientset,
	metricsclientset *versioned.Clientset,
	podQuery, containerQuery, nodeQuery *regexp.Regexp,
) error {
	// start termui
	if err := termui.Init(); err != nil {
		return err
	}
	defer termui.Close()

	// draw grid
	dashboard := dashboard.New()
	grid := termui.NewGrid()
	grid.Set(
		termui.NewRow(1./2, dashboard.Table()),
		termui.NewRow(1./4, dashboard.CPUGraph().Grid()),
		termui.NewRow(1./4, dashboard.MemoryGraph().Grid()),
	)

	resizing := func() {
		width, height := termui.TerminalDimensions()
		grid.SetRect(0, 1, width, height-1)
	}

	resizing()
	termui.Render(grid)

	errCh := make(chan error)

	tick := time.NewTicker(k.interval)
	recv := make(chan resources.Resources)

	// scheduled to fetch resources from kubernetes metrics server.
	go func() {
		for {
			select {
			case <-tick.C:
				r, err := resources.FetchResources(
					*k.kubeFlags.Namespace,
					clientset,
					metricsclientset,
					podQuery,
					containerQuery,
					nodeQuery,
				)
				if err != nil {
					errCh <- err
					return
				}
				recv <- r
			}
		}
	}()

	event := termui.PollEvents()
	doneCh := make(chan struct{})

	go func() {
		for r := range recv {
			// update table:
			go func(r resources.Resources) {
				var d drawer.TableDrawer
				if r.Len() > 0 {
					d = &drawer.KubeTableDrawer{}
				} else {
					d = &drawer.NopTableDrawer{}
				}
				dashboard.DrawTable(d, r)
				termui.Render(grid)
			}(r)

			// update cpu graph:
			go func(r resources.Resources) {
				dashboard.DrawCPUGraph(r)
				termui.Render(grid)
			}(r)

			// update memory graph:
			go func(r resources.Resources) {
				dashboard.DrawMemoryGraph(r)
				termui.Render(grid)
			}(r)
		}
	}()

	go func() {
		for e := range event {
			switch e.ID {
			case "<Enter>":
				dashboard.Toggle()
			case "<Down>":
				dashboard.ScrollDown()
			case "<Up>":
				dashboard.ScrollUp()
			case "q", "<C-c>":
				doneCh <- struct{}{}
				return
			case "<Resize>":
				resizing()
			}
			termui.Render(grid)
		}
	}()

	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGTERM, os.Interrupt)

	for {
		defer func() {
			close(sig)
			close(recv)
			close(errCh)
			close(doneCh)
		}()
		select {
		case <-sig:
			return nil
		case <-doneCh:
			return nil
		case err := <-errCh:
			if err != nil {
				return err
			}
		}
	}
}

func (k *ktop) run(cmd *cobra.Command, args []string) error {
	if *k.kubeFlags.Namespace == "" {
		*k.kubeFlags.Namespace = "default"
	}
	if k.allNamespaces {
		*k.kubeFlags.Namespace = ""
	}

	// kubernetes clients
	clientset, metricsclientset, err := k.kubeclient()
	if err != nil {
		return err
	}

	// regexp queries
	podQuery, err := regexp.Compile(k.podQuery)
	if err != nil {
		return err
	}
	containerQuery, err := regexp.Compile(k.containerQuery)
	if err != nil {
		return err
	}
	nodeQuery, err := regexp.Compile(k.nodeQuery)
	if err != nil {
		return err
	}

	return k.loop(
		clientset,
		metricsclientset,
		podQuery,
		containerQuery,
		nodeQuery,
	)
}

func (k *ktop) kubeclient() (*kubernetes.Clientset, *versioned.Clientset, error) {
	config, err := k.kubeFlags.ToRESTConfig()
	if err != nil {
		return nil, nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, err
	}
	metricsclientset, err := versioned.NewForConfig(config)
	if err != nil {
		return nil, nil, err
	}
	return clientset, metricsclientset, nil
}
