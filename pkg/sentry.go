package pkg

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/getsentry/sentry-go"
)

func SetupSentry() {
	sigc := make(chan os.Signal, 2)
	signal.Notify(sigc, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	err := sentry.Init(sentry.ClientOptions{
		Debug:            true,
		Dsn:              SentryDSN,
		AttachStacktrace: true,
		Release:          Commit,
	})
	if err != nil {
		panic(fmt.Errorf("failed to setup sentry: %w", err))
	}

	go func() {
		<-sigc
		sentry.Flush(2 * time.Second)
	}()
}
