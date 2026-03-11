package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	kratos "github.com/go-kratos/kratos/v2"
	klog "github.com/go-kratos/kratos/v2/log"
	khttp "github.com/go-kratos/kratos/v2/transport/http"

	"myblog/internal/conf"
	"myblog/internal/data"
	"myblog/internal/handler"
	"myblog/internal/service"
)

func main() {
	cfg, err := conf.Load("config.yaml")
	if err != nil {
		panic(err)
	}

	db, err := data.OpenDB(cfg)
	if err != nil {
		panic(err)
	}

	if err := data.MigrateAndSeed(context.Background(), db, cfg); err != nil {
		panic(err)
	}

	repo := data.NewRepository(db)
	svc := service.NewAppService(cfg, repo)
	httpHandler := handler.NewHTTPHandler(cfg, svc)

	httpSrv := khttp.NewServer(
		khttp.Address(cfg.Server.HTTP.Addr),
		khttp.Timeout(cfg.HTTPTimeout()),
	)
	httpSrv.HandlePrefix("/", httpHandler.Routes())

	logger := klog.NewStdLogger(os.Stdout)
	app := kratos.New(
		kratos.Name("myblog"),
		kratos.Version("1.0.0"),
		kratos.Logger(logger),
		kratos.Server(httpSrv),
	)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		_ = app.Stop()
	}()

	if err := app.Run(); err != nil {
		panic(err)
	}
}
