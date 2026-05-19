package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"seat-booking-backend/internal/model"
	"seat-booking-backend/internal/store"
)

type API struct {
	store *store.Store
}

func New(repo *store.Store) *API {
	return &API{store: repo}
}

func (a *API) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", a.handleHealth)
	mux.HandleFunc("GET /api/seats", a.handleListSeats)
	mux.HandleFunc("GET /api/reservations", a.handleListReservations)
	mux.HandleFunc("POST /api/reservations", a.handleCreateReservation)
	mux.HandleFunc("POST /api/reservations/{id}/cancel", a.handleCancelReservation)

	return withCORS(mux)
}

func (a *API) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, model.APIResponse{
		Success: true,
		Message: "ok",
		Data: map[string]string{
			"service": "seat-booking-backend",
			"time":    time.Now().Format(time.RFC3339),
		},
	})
}

func (a *API) handleListSeats(w http.ResponseWriter, r *http.Request) {
	start, end, err := a.readWindowOrNow(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	seats, err := a.store.ListSeats(r.Context(), start, end)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, model.APIResponse{
		Success: true,
		Message: "ok",
		Data: map[string]any{
			"query_window": map[string]string{
				"start_time": start.Format(time.RFC3339),
				"end_time":   end.Format(time.RFC3339),
			},
			"items": seats,
		},
	})
}

func (a *API) handleCreateReservation(w http.ResponseWriter, r *http.Request) {
	var req model.CreateReservationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON request body")
		return
	}

	reservation, err := a.store.CreateReservation(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrSeatNotFound):
			writeError(w, http.StatusNotFound, "seat not found")
		case errors.Is(err, store.ErrSeatNotBookable):
			writeError(w, http.StatusConflict, "seat is fixed or inactive and cannot be booked")
		case errors.Is(err, store.ErrReservationConflict):
			writeError(w, http.StatusConflict, "seat has already been booked in the selected time range")
		default:
			writeError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	writeJSON(w, http.StatusCreated, model.APIResponse{
		Success: true,
		Message: "reservation created",
		Data:    reservation,
	})
}

func (a *API) handleCancelReservation(w http.ResponseWriter, r *http.Request) {
	rawID := r.PathValue("id")
	id, err := store.ParseReservationID(rawID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	reservation, err := a.store.CancelReservation(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrReservationNotFound):
			writeError(w, http.StatusNotFound, "reservation not found")
		case errors.Is(err, store.ErrReservationNotActive):
			writeError(w, http.StatusConflict, "reservation is not active")
		default:
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	writeJSON(w, http.StatusOK, model.APIResponse{
		Success: true,
		Message: "reservation cancelled",
		Data:    reservation,
	})
}

func (a *API) handleListReservations(w http.ResponseWriter, r *http.Request) {
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	if status == "" {
		status = "all"
	}
	if status != "all" && status != "active" && status != "cancelled" {
		writeError(w, http.StatusBadRequest, "status must be all, active, or cancelled")
		return
	}

	query := store.ReservationQuery{
		Status:   status,
		SeatCode: strings.TrimSpace(r.URL.Query().Get("seat_code")),
	}

	startRaw := strings.TrimSpace(r.URL.Query().Get("start_time"))
	endRaw := strings.TrimSpace(r.URL.Query().Get("end_time"))
	if (startRaw == "") != (endRaw == "") {
		writeError(w, http.StatusBadRequest, "start_time and end_time must be provided together")
		return
	}
	if startRaw != "" {
		start, err := a.store.ParseInputTime(startRaw)
		if err != nil {
			writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid start_time: %v", err))
			return
		}
		end, err := a.store.ParseInputTime(endRaw)
		if err != nil {
			writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid end_time: %v", err))
			return
		}
		if !start.Before(end) {
			writeError(w, http.StatusBadRequest, "end_time must be later than start_time")
			return
		}
		query.StartTime = &start
		query.EndTime = &end
	}

	items, err := a.store.ListReservations(r.Context(), query)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, model.APIResponse{
		Success: true,
		Message: "ok",
		Data: map[string]any{
			"items": items,
		},
	})
}

func (a *API) readWindowOrNow(r *http.Request) (time.Time, time.Time, error) {
	startRaw := strings.TrimSpace(r.URL.Query().Get("start_time"))
	endRaw := strings.TrimSpace(r.URL.Query().Get("end_time"))

	if startRaw == "" && endRaw == "" {
		now := time.Now()
		return now, now.Add(time.Second), nil
	}
	if (startRaw == "") != (endRaw == "") {
		return time.Time{}, time.Time{}, fmt.Errorf("start_time and end_time must be provided together")
	}

	start, err := a.store.ParseInputTime(startRaw)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid start_time: %w", err)
	}
	end, err := a.store.ParseInputTime(endRaw)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid end_time: %w", err)
	}
	if !start.Before(end) {
		return time.Time{}, time.Time{}, fmt.Errorf("end_time must be later than start_time")
	}
	return start, end, nil
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, model.APIResponse{
		Success: false,
		Message: message,
	})
}

func writeJSON(w http.ResponseWriter, status int, payload model.APIResponse) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
