package rest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"go.uber.org/zap"
)

// RunHttpServer starts a http server on given port. handler will be created when making the http.Server object.
// it will be a blocking call and will do gracefully shutdown the server when given context canceled.
// example:
//
//	gox.RunHttpServer(ctx, func(router chi.Router) http.Handler {
//		return SetupRouter(router).Handler
//	}, WithPort("9090"))
func RunHttpServer(ctx context.Context, createHandler func(router chi.Router) http.Handler, options ...Option) {
	nopLogger := zap.NewNop()
	cfg := &config{
		port:   DefaultPort,
		logger: nopLogger,
	}

	for _, o := range options {
		if err := o(cfg); err != nil {
			log.Fatalf("Run HTTP server failed %e", err)
		}
	}

	apiRouter := chi.NewRouter()
	if len(cfg.middlewares) == 0 {
		// set default middlewares
		apiRouter.Use(DefaultMiddlewares()...)
	}

	if !cfg.setCors {
		apiRouter.Use(cors.New(DefaultCorsOption()).Handler)
	}

	if cfg.logger == nopLogger {
		log.Println("WARN: no logger is set")
	} else {
		apiRouter.Use(RequestLogger(cfg.logger))
	}

	srv := &http.Server{
		Addr:    net.JoinHostPort("", cfg.port),
		Handler: createHandler(apiRouter),
	}

	go func() {
		cfg.logger.Info("Start http server", zap.String("port", cfg.port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			cfg.logger.Fatal("Start HTTP server failed", zap.Error(err))
		}
	}()

	<-ctx.Done()
	cfg.logger.Info("Http Server received a shutdown signal", zap.Int("gracefulShutdownSec", DefaultGracefulShutdownSec))

	shutdownCtx, cancel := context.WithTimeout(context.Background(), DefaultGracefulShutdownSec*time.Second)
	defer func() {
		cancel()
	}()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		cfg.logger.Fatal("http server shutdown failed", zap.Error(err))
	}
	cfg.logger.Info("Http Server exited properly")
}

func WriteJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		_, _ = fmt.Fprintln(w, err)
	}
}

// Message defines a general model for server messages
type Message struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func WriteMessage(w http.ResponseWriter, code string, message string) {
	WriteJSON(w, http.StatusOK, Message{
		Code:    code,
		Message: message,
	})
}

func WriteError(w http.ResponseWriter, code int, message string) {
	WriteJSON(w, code, Message{
		Code:    errCodeFromHttp(code),
		Message: message,
	})
}

func ReadJSON(r *http.Request, target interface{}) (int, error) {
	dec := json.NewDecoder(r.Body)

	err := dec.Decode(&target)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {

		case errors.As(err, &syntaxError):
			return http.StatusBadRequest, fmt.Errorf("request body contains badly-formed JSON (at position %d)", syntaxError.Offset)

		// In some circumstances Decode() may also return an
		// io.ErrUnexpectedEOF error for syntax errors in the JSON. There
		// is an open issue regarding this at
		// https://github.com/golang/go/issues/25956.
		case errors.Is(err, io.ErrUnexpectedEOF):
			return http.StatusBadRequest, fmt.Errorf("request body contains badly-formed JSON")

		// Catch any type errors, like trying to assign a string in the
		// JSON request body to a int field in our Person struct. We can
		// interpolate the relevant field name and position into the error
		// message to make it easier for the client to fix.
		case errors.As(err, &unmarshalTypeError):
			return http.StatusBadRequest, fmt.Errorf("request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)

		// Catch the error caused by extra unexpected fields in the request
		// body. We extract the field name from the error message and
		// interpolate it in our custom error message. There is an open
		// issue at https://github.com/golang/go/issues/29035 regarding
		// turning this into a sentinel error.
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return http.StatusBadRequest, fmt.Errorf("request body contains unknown field %s", fieldName)

		// An io.EOF error is returned by Decode() if the request body is
		// empty.
		case errors.Is(err, io.EOF):
			return http.StatusBadRequest, fmt.Errorf("request body must not be empty")

		// Catch the error caused by the request body being too large. Again
		// there is an open issue regarding turning this into a sentinel
		// error at https://github.com/golang/go/issues/30715.
		case err.Error() == "http: request body too large":
			return http.StatusBadRequest, fmt.Errorf(err.Error())

		default:
			return http.StatusBadRequest, fmt.Errorf(http.StatusText(http.StatusInternalServerError))
		}
	}

	// Call decode again, using a pointer to an empty anonymous struct as
	// the destination. If the request body only contained a single JSON
	// object this will return an io.EOF error. So if we get anything else,
	// we know that there is additional data in the request body.
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return http.StatusBadRequest, fmt.Errorf("request body must only contain a single JSON object")
	}

	return http.StatusOK, nil
}

func DefaultBadRequestHandler(w http.ResponseWriter, _ *http.Request, err error) {
	WriteError(w, http.StatusBadRequest, err.Error())
}

func errCodeFromHttp(code int) string {
	codeMap := map[int]string{
		http.StatusBadRequest:          "ErrBadRequest",
		http.StatusInternalServerError: "ErrInternalServer",
		http.StatusUnauthorized:        "ErrUnauthorized",
		http.StatusConflict:            "ErrAlreadyExist",
		http.StatusForbidden:           "ErrForbidden",
	}

	if c, ok := codeMap[code]; ok {
		return c
	}

	return "ErrInternalServer"
}

func RequestLogger(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			method := r.Method
			query := r.URL.RawQuery
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			t0 := time.Now()
			next.ServeHTTP(ww, r)
			latency := time.Since(t0)

			logFunc := logger.Info
			if ww.Status() >= http.StatusInternalServerError {
				logFunc = logger.Error
			}

			logFunc(fmt.Sprintf("request handled: %s %s", method, path),
				zap.Int("code", ww.Status()),
				zap.String("query", query),
				zap.Duration("latency", latency))
		})
	}
}

func DefaultCorsOption() cors.Options {
	return cors.Options{
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}
}

func DefaultMiddlewares() []func(next http.Handler) http.Handler {
	return []func(next http.Handler) http.Handler{
		middleware.Timeout(60 * time.Second),
		middleware.RequestID,
		middleware.RealIP,
		middleware.Recoverer,
		middleware.SetHeader("X-Content-Type-Options", "nosniff"),
		middleware.SetHeader("X-Frame-Options", "deny"),
		middleware.NoCache,
	}
}
