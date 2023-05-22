package probe

import (
	"fmt"
	"net/http"
)

type Type int

const (
	Readiness = iota
	Aliveness
)

type Probe struct {
	probe   Type
	handler func() error
}

func WithProbe(probeType Type, handler func() error) Probe {
	return Probe{probe: probeType, handler: handler}
}

func New(router *http.ServeMux, probes ...Probe) http.Handler {
	router.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		if err := checkProbes(probes, Readiness); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(fmt.Sprintf(`{"status": "%s"}`, err)))
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "ready"}`))
	})

	router.HandleFunc("/alive", func(w http.ResponseWriter, r *http.Request) {
		if err := checkProbes(probes, Aliveness); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(fmt.Sprintf(`{"status": "%s"}`, err)))
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "alive"}`))
	})

	return router
}

func checkProbes(probes []Probe, t Type) error {
	for _, c := range probes {
		if c.probe != t {
			continue
		}

		// Run the check and fast fail if failed
		if err := c.handler(); err != nil {
			return err
		}
	}
	return nil
}
