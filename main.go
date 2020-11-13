package main

import (
	"context"
	"flag"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/magicdvd/aguanbot/event"
	"github.com/magicdvd/aguanbot/service"
	logger "github.com/whatisfaker/zaptrace/log"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

var (
	loginDur  = flag.Duration("login_expire", 20*time.Hour, "login expire default 20h")
	postMax   = flag.Uint("post_max", 10, "post max: 10")
	userFiles = flag.String("user_file", "userlist", "userlist per line(phone pwd)")
	duration  = flag.Duration("dur", 1*time.Hour, "one activity last duration")
	loglevel  = flag.String("level", "", "log level(debug,info,warn,error,panic,fatal)")
	logfile   = flag.String("logfile", "", "log file path")
	signals   = make(chan os.Signal, 1)
)

func main() {
	flag.Parse()
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	//log := log.NewFileLogger("log", *loglevel)
	var log *logger.Factory
	if *logfile == "" {
		log = logger.NewStdLogger(*loglevel)
	} else {
		log = logger.NewFileLogger(*logfile, *loglevel)
	}
	b, err := ioutil.ReadFile(*userFiles)
	if err != nil {
		log.Normal().Fatal("read user files error", zap.Error(err))
		return
	}
	users := strings.Split(string(b), "\n")
	if len(users) == 0 || users[0] == "" {
		log.Normal().Fatal("user files format error |phone pwd|")
		return
	}
	err = service.Mgr().Register(
		service.NewAguanAPI(log.With(zap.String("m", "api"))),
		service.NewBotManager(users, *loginDur),
	)
	if err != nil {
		log.Normal().Fatal("register error", zap.Error(err))
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	grp, ctx := errgroup.WithContext(ctx)
	grp.Go(func() error {
		evt := event.NewActivity(int(*postMax), *duration, log.With(zap.String("m", "event")))
		err := evt.Do(ctx)
		if err != nil {
			return err
		}
		ticker := time.NewTicker(*duration)
		defer ticker.Stop()
		log.Normal().Info("server start")
		for {
			select {
			case <-ticker.C:
				err := evt.Do(ctx)
				if err != nil {
					return err
				}
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	})
	grp.Go(func() error {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
		select {
		case <-signals:
			cancel()
		case <-ctx.Done():
			return ctx.Err()
		}
		return nil
	})
	err = grp.Wait()
	if err != nil && err != context.Canceled {
		log.Trace(ctx).Error("fail", zap.Error(err))
	}
}
