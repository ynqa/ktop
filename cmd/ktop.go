package cmd

import (
	"os"
	"os/signal"
	"regexp"
	"sync"
	"syscall"
	"time"

	"github.com/gizak/termui/v3"
	"github.com/spf13/cobra"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/ynqa/ktop/pkg/ktop"
	"github.com/ynqa/ktop/pkg/kube"
	"github.com/ynqa/ktop/pkg/ui"
)

const (
	logoStr = `
__  __    ______   ______     ______  
/\ \/ /   /\__  _\ /\  __ \   /\  == \ 
\ \  _"-. \/_/\ \/ \ \ \/\ \  \ \  _-/ 
 \ \_\ \_\   \ \_\  \ \_____\  \ \_\   
  \/_/\/_/    \/_/   \/_____/   \/_/   																			
`
	hintStr = `
<q>, <C-c>      Quit
<Up>            Up
<Down>          Down
<Right>, <Left> Switch Table Mode
`
)

type ktopCmd struct {
	k8sFlags       *genericclioptions.ConfigFlags
	interval       time.Duration
	nodeQuery      string
	podQuery       string
	containerQuery string
	renderMutex    sync.RWMutex
}

func newKtopCmd() *cobra.Command {
	ktop := ktopCmd{}
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
	ktop.k8sFlags = genericclioptions.NewConfigFlags()
	ktop.k8sFlags.AddFlags(cmd.Flags())
	if *ktop.k8sFlags.Namespace == "" {
		*ktop.k8sFlags.Namespace = "default"
	}
	return cmd
}

func (k *ktopCmd) render(items ...termui.Drawable) {
	k.renderMutex.Lock()
	defer k.renderMutex.Unlock()
	termui.Render(items...)
}

func (k *ktopCmd) run(cmd *cobra.Command, args []string) error {
	if err := termui.Init(); err != nil {
		return err
	}
	defer termui.Close()

	kubeclients, err := kube.NewKubeClients(k.k8sFlags)
	if err != nil {
		return err
	}

	// define queries
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

	monitor := ktop.NewMonitor(kubeclients, podQuery, containerQuery, nodeQuery)
	logo := ui.NewTextField()
	logo.Text = logoStr
	logo.TextStyle = termui.NewStyle(termui.ColorWhite, termui.ColorClear, termui.ModifierBold)
	hint := ui.NewTextField()
	hint.Text = hintStr
	hint.TextStyle = termui.NewStyle(termui.Color(244), termui.ColorClear)

	grid := termui.NewGrid()
	grid.Set(
		termui.NewRow(1./6,
			termui.NewCol(1./2, logo),
			termui.NewCol(1./2, hint),
		),
		termui.NewRow(3./12, monitor.GetPodTable()),
		termui.NewRow(5./12, monitor.GetLogs()),
		termui.NewRow(2./12,
			termui.NewCol(1./2, monitor.GetCPUGraph()),
			termui.NewCol(1./2, monitor.GetMemGraph()),
		),
	)
	termWidth, termHeight := termui.TerminalDimensions()
	grid.SetRect(0, 0, termWidth, termHeight)

	events := termui.PollEvents()
	tick := time.NewTicker(k.interval)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, os.Interrupt)

	for {
		select {
		case <-sigCh:
			return nil
		case <-tick.C:
			if err := monitor.Update(); err != nil {
				return err
			}
		case e := <-events:
			switch e.ID {
			case "<Down>":
				monitor.ScrollDown()
			case "<Up>":
				monitor.ScrollUp()
			case "<Right>":
				monitor.Rotate()
			case "<Left>":
				monitor.ReverseRotate()
			case "q", "<C-c>":
				return nil
			case "<Resize>":
				termWidth, termHeight := termui.TerminalDimensions()
				grid.SetRect(0, 0, termWidth, termHeight)
			}
		}
		k.render(grid)
	}
}

func Execute() {
	rootCmd := newKtopCmd()
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
