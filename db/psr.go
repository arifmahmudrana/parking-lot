package db

import (
	"context"
	"database/sql"
	"time"
)

func (d *DB) CreateParkingSpaceReservation(parkingspaceID, userID int) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	tx, err := d.dbConn.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `UPDATE parking_spaces SET status = ? WHERE (id = ?)`)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, booked, parkingspaceID)
	if err != nil {
		return 0, err
	}

	stmt, err = tx.PrepareContext(ctx,
		`insert into parking_space_reservations (user_id, start_time, parking_spaces_id) values (?, ?, ?)`)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, userID, time.Now().UTC().Format(dateFormat), parkingspaceID)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}

	return id, nil
}

func (d *DB) UnParkParkingSpaceByID(parkingSpaceReservationsID int) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	// Get the reservation
	stmt, err := d.dbConn.PrepareContext(ctx, `SELECT *
	                               						 FROM parking_space_reservations
																 						 WHERE id = ?
																						 limit 1`)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, parkingSpaceReservationsID)
	if row == nil {
		return 0, ErrNilQueryRowContext
	}

	err = row.Err()
	if err != nil {
		return 0, err
	}

	var psrRow struct {
		id, userID, fee, parkingSpacesID int
		startTime                        string
		endTime                          sql.NullString
	}
	if err := row.Scan(
		&psrRow.id, &psrRow.userID, &psrRow.startTime,
		&psrRow.endTime, &psrRow.fee, &psrRow.parkingSpacesID,
	); err != nil {
		return 0, err
	}

	if psrRow.endTime.Valid {
		return 0, ErrAlreadyUnparked
	}

	// get end time
	endTime := time.Now().UTC()
	startTime, err := time.Parse(dateFormat, psrRow.startTime)
	if err != nil {
		return 0, err
	}

	// calculate fee
	duration := endTime.Sub(startTime)
	h, rem := int(duration/time.Hour), int(duration%time.Hour)
	fee := h * 10
	if rem > 0 {
		fee += 10
	}

	tx, err := d.dbConn.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// update reservation
	stmt, err = tx.PrepareContext(ctx, `UPDATE parking_space_reservations SET end_time = ?, fee = ? WHERE (id = ?)`)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, endTime.Format(dateFormat), fee, parkingSpaceReservationsID)
	if err != nil {
		return 0, err
	}

	// update parking space make it available
	stmt, err = tx.PrepareContext(ctx, `UPDATE parking_spaces SET status = ? WHERE (id = ?)`)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, available, psrRow.parkingSpacesID)
	if err != nil {
		return 0, err
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}

	return fee, nil
}
