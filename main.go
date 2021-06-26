package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/itsubaki/gostream-server/pkg/config"
	"github.com/itsubaki/gostream-server/pkg/gostream"
	"github.com/itsubaki/gostream-server/pkg/plugin"
)

func main() {
	c, err := config.New()
	if err != nil {
		fmt.Printf("new config: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("config=%v\n", c)

	g := gostream.New()
	p := map[string]plugin.Plugin{
		"LogEventPlugin": &plugin.LogEventPlugin{},
	}

	for _, r := range c.Router {
		pp, ok := p[r.Plugin]
		if !ok {
			fmt.Printf("invalid plugin=%v", r.Plugin)
			os.Exit(1)
		}

		if err := pp.Setup(g, r.Path, r.Query); err != nil {
			fmt.Printf("setup failed. path=%v query=%v: %v", r.Path, r.Query, err)
			os.Exit(1)
		}
	}

	s := &http.Server{
		Addr:    c.Port,
		Handler: g.Handler(),
	}

	go func() {
		log.Println("http server listen and serve")
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	ch := make(chan os.Signal, 2)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	<-ch

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		log.Fatalf("http server shutdown: %v\n", err)
	}

	g.Close()
	log.Println("shutdown finished")
}
