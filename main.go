package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/hashicorp/http-echo/version"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	listenFlag = flag.String("listen", ":5678", "address and port to listen")
	// textFlag    = flag.String("text", "", "text to put on the webpage")
	versionFlag = flag.Bool("version", false, "display version information")
	serverId    = flag.String("id", "1", "Server id")
	delay       = flag.Int("delay", 0, "optional delay to apply to each response")

	// stdoutW and stderrW are for overriding in test.
	stdoutW     = os.Stdout
	stderrW     = os.Stderr
	reqsCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "reqs_count",
		Help: "The total number of received requests",
	},
		[]string{"server_id"})
)

func main() {
	flag.Parse()

	// Asking for the version?
	if *versionFlag {
		fmt.Fprintln(stderrW, version.HumanVersion)
		os.Exit(0)
	}

	// // Validation
	// if *textFlag == "" {
	// 	fmt.Fprintln(stderrW, "Missing -text option!")
	// 	os.Exit(127)
	// }

	args := flag.Args()
	if len(args) > 0 {
		fmt.Fprintln(stderrW, "Too many arguments!")
		os.Exit(127)
	}

	// Flag gets printed as a page
	mux := http.NewServeMux()
	mux.HandleFunc("/", httpLog(stdoutW, withAppHeaders(httpEcho(fmt.Sprintf("Port: %s, id: %s", *listenFlag, *serverId)))))

	// Health endpoint
	mux.HandleFunc("/health", withAppHeaders(httpHealth()))

	// Register metrics
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:    *listenFlag,
		Handler: mux,
	}
	serverCh := make(chan struct{})
	go func() {
		log.Printf("[INFO] server is listening on %s\n", *listenFlag)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("[ERR] server exited with: %s", err)
		}
		close(serverCh)
	}()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)

	// Wait for interrupt
	<-signalCh

	log.Printf("[INFO] received interrupt, shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("[ERR] failed to shutdown server: %s", err)
	}

	// If we got this far, it was an interrupt, so don't exit cleanly
	os.Exit(2)
}

func httpEcho(v string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Duration(*delay) * time.Millisecond)
		reqsCounter.WithLabelValues(*serverId).Inc()
		fmt.Fprintln(w, v)
	}
}

func httpHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"status":"ok"}`)
	}
}
