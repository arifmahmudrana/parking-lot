package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/arifmahmudrana/parking-lot/db"
	"github.com/go-chi/chi/v5"
)

func (app *application) GetParkingLots(w http.ResponseWriter, r *http.Request) {
	var (
		err  error
		p    = 1
		page = r.URL.Query().Get("page")
	)
	if page != "" {
		p, err = strconv.Atoi(page)
		if err != nil {
			app.errorLog.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	if p < 1 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	count, err := app.dbRepo.GetTotalCountParkingLots()
	if err != nil {
		app.errorLog.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	parkingLots, err := app.dbRepo.GetParkingLots(p)
	if err != nil {
		app.errorLog.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	resultData := struct {
		Data        []db.ParkingLot `json:"data"`
		TotalCount  int             `json:"total_count"`
		CurrentPage int             `json:"current_page"`
	}{
		Data:        parkingLots,
		TotalCount:  count,
		CurrentPage: p,
	}
	encoder := json.NewEncoder(w)
	if err := encoder.Encode(resultData); err != nil {
		app.errorLog.Println(err)
	}
}

func (app *application) CreateParkingLots(w http.ResponseWriter, r *http.Request) {
	var (
		p   db.ParkingLot
		dec = json.NewDecoder(r.Body)
	)
	if err := dec.Decode(&p); err != nil {
		app.errorLog.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	p.Name = strings.TrimSpace(p.Name)
	if p.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id, err := app.dbRepo.CreateParkingLot(p)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	resultData := struct {
		ID int64 `json:"id"`
	}{
		ID: id,
	}
	encoder := json.NewEncoder(w)
	w.WriteHeader(http.StatusCreated)
	if err := encoder.Encode(resultData); err != nil {
		app.errorLog.Println(err)
	}
}

func (app *application) GetParkingSpaces(w http.ResponseWriter, r *http.Request) {
	parkinglotID, err := strconv.Atoi(chi.URLParam(r, "parkinglotID"))
	if err != nil {
		app.errorLog.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	parkingSpaces, err := app.dbRepo.GetParkingSpacesByParkingLot(parkinglotID)
	if err != nil {
		app.errorLog.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	resultData := struct {
		Data []db.ParkingSpace `json:"data"`
	}{
		Data: parkingSpaces,
	}
	encoder := json.NewEncoder(w)
	if err := encoder.Encode(resultData); err != nil {
		app.errorLog.Println(err)
	}
}

func (app *application) CreateParkingSpaces(w http.ResponseWriter, r *http.Request) {
	parkinglotID, err := strconv.Atoi(chi.URLParam(r, "parkinglotID"))
	if err != nil {
		app.errorLog.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	exists, err := app.dbRepo.DoesParkingLotExistByID(parkinglotID)
	if err != nil {
		app.errorLog.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !exists {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id, err := app.dbRepo.CreateParkingSpaceFromParkingLotID(parkinglotID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	resultData := struct {
		ID int64 `json:"id"`
	}{
		ID: id,
	}
	encoder := json.NewEncoder(w)
	w.WriteHeader(http.StatusCreated)
	if err := encoder.Encode(resultData); err != nil {
		app.errorLog.Println(err)
	}
}

func (app *application) ParkParkingSpaces(w http.ResponseWriter, r *http.Request) {
	parkinglotID, err := strconv.Atoi(chi.URLParam(r, "parkinglotID"))
	if err != nil {
		app.errorLog.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	parkingspaceID, err := app.dbRepo.GetNextParkingSpaceByParkingLot(parkinglotID)
	if err != nil || parkinglotID <= 0 {
		app.errorLog.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var (
		b struct {
			UserID int `json:"user_id"`
		}
		dec = json.NewDecoder(r.Body)
	)
	if err := dec.Decode(&b); err != nil {
		app.errorLog.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id, err := app.dbRepo.CreateParkingSpaceReservation(parkingspaceID, b.UserID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	resultData := struct {
		ID int64 `json:"id"`
	}{
		ID: id,
	}
	encoder := json.NewEncoder(w)
	w.WriteHeader(http.StatusCreated)
	if err := encoder.Encode(resultData); err != nil {
		app.errorLog.Println(err)
	}
}

func (app *application) UnParkParkingSpace(w http.ResponseWriter, r *http.Request) {
	parkingSpaceReservationsID, err := strconv.Atoi(chi.URLParam(r, "parkingSpaceReservationsID"))
	if err != nil {
		app.errorLog.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	fee, err := app.dbRepo.UnParkParkingSpaceByID(parkingSpaceReservationsID)
	if err != nil {
		app.errorLog.Println(err)
		st := http.StatusInternalServerError
		if err == db.ErrNilQueryRowContext || err == db.ErrAlreadyUnparked || err == sql.ErrNoRows {
			st = http.StatusBadRequest
		}
		w.WriteHeader(st)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	resultData := struct {
		Fee int `json:"fee"`
	}{
		Fee: fee,
	}
	encoder := json.NewEncoder(w)
	w.WriteHeader(http.StatusCreated)
	if err := encoder.Encode(resultData); err != nil {
		app.errorLog.Println(err)
	}
}

func (app *application) ParkingSpaceMaintanance(w http.ResponseWriter, r *http.Request) {
	parkinglotID, err := strconv.Atoi(chi.URLParam(r, "parkinglotID"))
	if err != nil {
		app.errorLog.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	parkingspaceID, err := strconv.Atoi(chi.URLParam(r, "parkingspaceID"))
	if err != nil {
		app.errorLog.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	exists, err := app.dbRepo.DoesParkingSpaceExistForMaintananceByParkingLotIDAndID(parkingspaceID, parkinglotID)
	if err != nil {
		app.errorLog.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !exists {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var (
		b struct {
			Maintanance bool `json:"maintanance"`
		}
		dec = json.NewDecoder(r.Body)
	)
	if err := dec.Decode(&b); err != nil {
		app.errorLog.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = app.dbRepo.SetParkingSpaceMaintanance(parkingspaceID, b.Maintanance)
	if err != nil {
		app.errorLog.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
