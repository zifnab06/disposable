package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog"
)

type set map[string]any

func updateDomains() (set, error) {
	resp, err := http.Get("https://raw.githubusercontent.com/disposable/disposable-email-domains/master/domains_strict.txt")
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Failed to download file: %s", resp.Status)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	domains := make(set)

	for _, domain := range strings.Split(string(bodyBytes), "\n") {
		domains[domain] = nil
	}

	return domains, nil

}

func main() {
	fmt.Println("Starting update...")
	domains, err := updateDomains()
	if err != nil {
		panic(err)
	}
	fmt.Println("Update finished!")

	updateTimer := time.NewTicker(12 * time.Hour)
	go func() {
		for {
			select {
			case <-updateTimer.C:
				newDomains, err := updateDomains()
				if err != nil {
					fmt.Printf("Failed to update domains: %s", err)
				} else {
					domains = newDomains
				}

			}
		}
	}()
	logger := httplog.NewLogger("disposable", httplog.Options{JSON: true})
	r := chi.NewRouter()
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(httplog.RequestLogger(logger))

	r.Get("/v1/domain/{domain}", func(w http.ResponseWriter, r *http.Request) {
		domain := chi.URLParam(r, "domain")
		w.Header().Add("Content-Type", "text/plain")
		if _, ok := domains[domain]; ok {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	})

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		message := `Disposable Email Checker
This uses https://github.com/disposable/disposable (specifically the strict domain list) to check if a given domain is using a disposable email service.

API:

GET /v1/domain/{domain}
	Returns 200 if this domain _is_ disposable
	Returns 404 if this doain _is not_ disposable
	Note: No text is returned, only the status code. Bytes cost money.
`
		w.Header().Add("Content-Type", "text/plain")
		w.Write([]byte(message))
	})

	http.ListenAndServe(":8080", r)
}
