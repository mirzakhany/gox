package probe

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	liveProbe := func() error {
		return nil
	}

	readyProbe := func() error {
		return errors.New("not ready")
	}

	{ // Default probe should return success for both readiness and aliveness
		probeHandler := New(nil)
		readyReq := httptest.NewRequest("GET", "/ready", nil)
		readyW := httptest.NewRecorder()
		probeHandler.ServeHTTP(readyW, readyReq)
		readyRes := readyW.Result()
		require.Equal(t, http.StatusOK, readyRes.StatusCode)

		aliveReq := httptest.NewRequest("GET", "/alive", nil)
		aliveW := httptest.NewRecorder()
		probeHandler.ServeHTTP(aliveW, aliveReq)
		aliveRes := aliveW.Result()
		require.Equal(t, http.StatusOK, aliveRes.StatusCode)
	}

	{
		// ready handler should return error
		probeHandler := New(nil, WithProbe(Readiness, readyProbe), WithProbe(Aliveness, liveProbe))
		readyReq := httptest.NewRequest("GET", "/ready", nil)
		readyW := httptest.NewRecorder()
		probeHandler.ServeHTTP(readyW, readyReq)
		readyRes := readyW.Result()
		require.Equal(t, http.StatusInternalServerError, readyRes.StatusCode)

		// alive probe should return success
		aliveReq := httptest.NewRequest("GET", "/alive", nil)
		aliveW := httptest.NewRecorder()
		probeHandler.ServeHTTP(aliveW, aliveReq)
		aliveRes := aliveW.Result()
		require.Equal(t, http.StatusOK, aliveRes.StatusCode)
	}
}
