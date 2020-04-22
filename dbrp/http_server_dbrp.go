package dbrp

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/influxdata/influxdb/v2"
	kithttp "github.com/influxdata/influxdb/v2/kit/transport/http"
	"go.uber.org/zap"
)

type DBRPHandler struct {
	chi.Router
	api     *kithttp.API
	log     *zap.Logger
	dbrpSvc influxdb.DBRPMappingServiceV2
}

// NewHTTPAuthHandler constructs a new http server.
func NewHTTPDBRPHandler(log *zap.Logger, dbrpSvc influxdb.DBRPMappingServiceV2) *DBRPHandler {
	h := &DBRPHandler{
		api:     kithttp.NewAPI(kithttp.WithLog(log)),
		log:     log,
		dbrpSvc: dbrpSvc,
	}

	r := chi.NewRouter()
	r.Use(
		middleware.Recoverer,
		middleware.RequestID,
		middleware.RealIP,
	)

	r.Route("/", func(r chi.Router) {
		r.Post("/", h.handlePostDBRP)
		r.Get("/", h.handleGetDBRPs)

		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.handleGetDBRP)
			r.Patch("/", h.handlePatchDBRP)
			r.Delete("/", h.handleDeleteDBRP)
		})
	})

	h.Router = r
	return h
}

func (h *DBRPHandler) handlePostDBRP(w http.ResponseWriter, r *http.Request) {
}

func (h *DBRPHandler) handleGetDBRPs(w http.ResponseWriter, r *http.Request) {
}

func (h *DBRPHandler) handleGetDBRP(w http.ResponseWriter, r *http.Request) {
}

func (h *DBRPHandler) handlePatchDBRP(w http.ResponseWriter, r *http.Request) {
}

func (h *DBRPHandler) handleDeleteDBRP(w http.ResponseWriter, r *http.Request) {
}
