package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (app *application) routes() http.Handler {
	mux := chi.NewRouter()

	mux.Use(
		middleware.Logger,
		middleware.RequestID,
		middleware.Recoverer,
	)

	mux.Get("/api/parking-lots", app.GetParkingLots)
	mux.Post("/api/parking-lots", app.CreateParkingLots)
	mux.Get("/api/parking-lots/{parkinglotID}/parking-spaces", app.GetParkingSpaces)
	mux.Post("/api/parking-lots/{parkinglotID}/parking-spaces", app.CreateParkingSpaces)
	mux.Post("/api/parking-lots/{parkinglotID}/park", app.ParkParkingSpaces)
	mux.Post("/api/parking-reservations/{parkingSpaceReservationsID}/unpark", app.UnParkParkingSpace)
	mux.Post("/api/parking-lots/{parkinglotID}/parking-spaces/{parkingspaceID}/maintanance", app.ParkingSpaceMaintanance)

	// TODO: implement feature The Parking Manager should be able to get the total number of vehicles parked on any day, total parking time and the total fee collected on

	return mux
}
