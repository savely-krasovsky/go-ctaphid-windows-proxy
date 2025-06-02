package main

import (
	"github.com/savely-krasovsky/go-ctaphid-windows-proxy/internal/config"
	"github.com/savely-krasovsky/go-ctaphid-windows-proxy/internal/domain"
	"log/slog"
	"os"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
)

const svcName = "Ozon Privileged Agent"

type program struct {
	logger   *slog.Logger
	config   *config.Config
	delivery domain.Delivery
}

func (p *program) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown
	changes <- svc.Status{State: svc.StartPending}
	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}

	go func() {
		l, err := p.delivery.Listen()
		if err != nil {
			p.logger.Error("Error while getting a listener!", "err", err)
			os.Exit(1)
		}

		if err := p.delivery.Serve(l); err != nil {
			p.logger.Error("Error while serving delivery!", "err", err)
			os.Exit(1)
		}
	}()

loop:
	for c := range r {
		switch c.Cmd {
		case svc.Interrogate:
			changes <- c.CurrentStatus
		case svc.Stop, svc.Shutdown:
			if err := p.delivery.Shutdown(); err != nil {
				p.logger.Error("Error while shitting down main delivery!", "err", err)
			}
			break loop
		default:
			p.logger.Info("Service shut down!", "cmd", c.Cmd)
		}
	}

	changes <- svc.Status{State: svc.StopPending}
	return

}

func (p *program) run(svcName string, isDebug bool) {
	p.logger.Info("Starting service!", "svc_name", svcName)
	run := svc.Run
	if isDebug {
		run = debug.Run
	}
	if err := run(svcName, p); err != nil {
		p.logger.Error("Error while running service!", "err", err, "svc_name", svcName)
		return
	}
	p.logger.Info("Service successfully shut down!", "svc_name", svcName)
}
