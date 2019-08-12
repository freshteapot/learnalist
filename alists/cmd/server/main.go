package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type key int

const (
	requestIDKey key = 0
)

type siteConfig struct {
	StaticSiteFolder string
}

var config siteConfig

func main() {
	var logger = logrus.New()

	logger.Formatter = new(logrus.JSONFormatter)
	w := logger.Writer()
	defer w.Close()

	staticSiteFolder := flag.String("static", "", "path to static site builder")
	listenOn := flag.String("port", "9091", "port to listen on")
	flag.Parse()

	*staticSiteFolder = strings.TrimRight(*staticSiteFolder, "/")

	if *staticSiteFolder == "" {
		log.Fatal("Will need the path to static site, add -static=XXX")
	}
	config.StaticSiteFolder = *staticSiteFolder
	//fs := http.FileServer(http.Dir(*staticSiteFolder))
	//http.Handle("/", fs)
	http.HandleFunc("/", serveFiles)

	nextRequestID := func() string {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}

	listenAddr := fmt.Sprintf(":%s", *listenOn)
	server := http.Server{
		Addr:         listenAddr,
		Handler:      tracing(nextRequestID)(logging(logger)(http.DefaultServeMux)),
		ErrorLog:     log.New(w, "", 0),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	done := make(chan bool)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	go func() {
		<-quit
		logger.Println("Server is shutting down...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		server.SetKeepAlivesEnabled(false)
		if err := server.Shutdown(ctx); err != nil {
			logger.Fatalf("Could not gracefully shutdown the server: %v\n", err)
		}
		close(done)
	}()

	logger.Info("Server is ready to handle requests at", listenAddr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatalf("Could not listen on %s: %v\n", listenAddr, err)
	}

	<-done
	logger.Info("Server stopped")
}

func logging(logger *logrus.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				requestID, ok := r.Context().Value(requestIDKey).(string)
				if !ok {
					requestID = "unknown"
				}

				logger.WithFields(logrus.Fields{
					"method":     r.Method,
					"path":       r.URL.Path,
					"remoteAddr": r.RemoteAddr,
					"userAgent":  r.UserAgent(),
					"requestID":  requestID,
				}).Info()
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func tracing(nextRequestID func() string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-Id")
			if requestID == "" {
				requestID = nextRequestID()
			}
			ctx := context.WithValue(r.Context(), requestIDKey, requestID)
			w.Header().Set("X-Request-Id", requestID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func serveFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	err := serveAlist(w, r)
	if err == nil {
		return
	}

	err = serveStatic(w, r)
	if err == nil {
		return
	}

	w.WriteHeader(http.StatusNotFound)
	fmt.Fprint(w, "custom 404")
}

func serveAlist(w http.ResponseWriter, r *http.Request) error {
	parts := strings.Split(r.URL.Path, "/")
	suffix := parts[len(parts)-1]
	parts = strings.Split(suffix, ".")
	if len(parts) != 2 {
		return errors.New("List not found")
	}

	uuid := parts[0]
	isA := parts[1]
	// This code should only serve the lists?
	path := fmt.Sprintf("%s/alists/%s.%s", config.StaticSiteFolder, uuid, isA)

	if _, err := os.Stat(path); err == nil {
		// TODO at this point we can do acl look up.
		http.ServeFile(w, r, path)
		return nil
	}
	return errors.New("List not found")
}

func serveStatic(w http.ResponseWriter, r *http.Request) error {
	// path/to/whatever does *not* exist
	path := fmt.Sprintf("%s/%s", config.StaticSiteFolder, r.URL.Path[1:])
	if _, err := os.Stat(path); err == nil {
		// path/to/whatever exists
		http.ServeFile(w, r, path)
		return nil
	}
	return errors.New("File not found")
}
