package dbrp

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/influxdata/httprouter"
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
	ctx := r.Context()
	dbrp := &influxdb.DBRPMapping{}

	if err := json.NewDecoder(r.Body).Decode(dbrp); err != nil {
		h.api.Err(w, &influxdb.Error{
			Code: influxdb.EInvalid,
			Msg:  "invalid json structure",
			Err:  err,
		})
		return
	}

	if err := h.dbrpSvc.Create(ctx, dbrp); err != nil {
		h.api.Err(w, &influxdb.Error{
			Code: influxdb.ErrInvalidID.Code,
			Err:  err,
			Msg:  "impossible to create mapping",
		})
		return
	}
	h.api.Respond(w, http.StatusCreated, dbrp)
}

func (h *DBRPHandler) handleGetDBRPs(w http.ResponseWriter, r *http.Request) {
	filter := influxdb.DBRPMappingFilter{}
	orgID := r.URL.Query().Get("orgID")
	if orgID != "" {
		id, err := influxdb.IDFromString(orgID)
		if err != nil {
			h.api.Err(w, influxdb.ErrInvalidID)
			return
		}
		filter.OrgID = id
	} else {
		h.api.Err(w, influxdb.ErrOrgNotFound)
		return
	}

	dbrps, _, err := h.dbrpSvc.FindMany(r.Context(), filter)
	if err != nil {
		h.api.Err(w, &influxdb.Error{
			Code: influxdb.EInternal,
			Err:  err,
		})
		return
	}

	h.api.Respond(w, http.StatusOK, struct {
		Content []*influxdb.DBRPMapping `json:"content"`
	}{
		Content: dbrps,
	})
}

func (h *DBRPHandler) handleGetDBRP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	params := httprouter.ParamsFromContext(ctx)
	id := params.ByName("id")
	if id == "" {
		h.api.Err(w, &influxdb.Error{
			Code: influxdb.EInvalid,
			Msg:  "url missing id",
		})
		return
	}

	var i influxdb.ID
	if err := i.DecodeFromString(id); err != nil {
		h.api.Err(w, err)
		return
	}

	dbrp, err := h.dbrpSvc.FindByID(ctx, i)
	if err != nil {
		h.api.Err(w, err)
		return
	}
	h.api.Respond(w, http.StatusOK, struct {
		Content *influxdb.DBRPMapping `json:"content"`
	}{
		Content: dbrp,
	})
}

func (h *DBRPHandler) handlePatchDBRP(w http.ResponseWriter, r *http.Request) {
	bodyRequest := &struct {
		Default bool `json:"content"`
	}{}

	ctx := r.Context()
	params := httprouter.ParamsFromContext(ctx)
	id := params.ByName("id")
	if id == "" {
		h.api.Err(w, &influxdb.Error{
			Code: influxdb.EInvalid,
			Msg:  "url missing id",
		})
		return
	}

	var i influxdb.ID
	if err := i.DecodeFromString(id); err != nil {
		h.api.Err(w, err)
		return
	}

	dbrp, err := h.dbrpSvc.FindByID(ctx, i)
	if err != nil {
		h.api.Err(w, err)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(bodyRequest); err != nil {
		h.api.Err(w, &influxdb.Error{
			Code: influxdb.EInvalid,
			Msg:  "invalid json structure",
			Err:  err,
		})
		return
	}

	dbrp.Default = bodyRequest.Default

	if err := h.dbrpSvc.Update(ctx, dbrp); err != nil {
		h.api.Err(w, err)
		return
	}

	h.api.Respond(w, http.StatusOK, struct {
		Content *influxdb.DBRPMapping `json:"content"`
	}{
		Content: dbrp,
	})
}

func (h *DBRPHandler) handleDeleteDBRP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	params := httprouter.ParamsFromContext(ctx)
	id := params.ByName("id")
	if id == "" {
		h.api.Err(w, &influxdb.Error{
			Code: influxdb.EInvalid,
			Msg:  "url missing id",
		})
		return
	}

	var i influxdb.ID
	if err := i.DecodeFromString(id); err != nil {
		h.api.Err(w, err)
		return
	}

	err := h.dbrpSvc.Delete(ctx, i)
	if err != nil {
		h.api.Err(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
