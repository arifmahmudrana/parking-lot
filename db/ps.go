package db

import "context"

// SELECT * FROM parking_lot.parking_spaces WHERE EXISTS (SELECT * FROM parking_lots where parking_lots.id = 1) and parking_spaces.parking_lots_id = 1;

type ParkingSpace struct {
	ID         int    `json:"id"`
	Status     string `json:"status"`
	SlotNumber int    `json:"slot_number"`
}

func (d *DB) GetParkingSpacesByParkingLot(parkingLotID int) ([]ParkingSpace, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt, err := d.dbConn.PrepareContext(ctx, `SELECT *
	                               						 FROM parking_spaces
																 						 WHERE EXISTS (
																							SELECT * FROM parking_lots where parking_lots.id = ?
																 						 ) and parking_spaces.parking_lots_id = ?
																						 order by created_at asc, parking_spaces.id asc`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, parkingLotID, parkingLotID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var parkingSpaces []ParkingSpace
	slotNumber := 1
	for rows.Next() {
		var ps struct {
			id              int
			createdAt       string
			status          int8
			parking_lots_id int
		}
		err := rows.Scan(
			&ps.id,
			&ps.createdAt,
			&ps.status,
			&ps.parking_lots_id,
		)
		if err != nil {
			return nil, err
		}

		parkingSpaces = append(parkingSpaces, ParkingSpace{
			ID:         ps.id,
			Status:     status(ps.status).value(),
			SlotNumber: slotNumber,
		})

		slotNumber++
	}

	return parkingSpaces, nil
}

func (d *DB) CreateParkingSpaceFromParkingLotID(plID int) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt, err := d.dbConn.PrepareContext(ctx,
		`insert into parking_spaces (parking_lots_id) values (?)`)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, plID)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (d *DB) GetNextParkingSpaceByParkingLot(parkingLotID int) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt, err := d.dbConn.PrepareContext(ctx, `SELECT parking_spaces.id
	                               						 FROM parking_spaces
																 						 WHERE EXISTS (
																							SELECT * FROM parking_lots where parking_lots.id = ?
																 						 ) and parking_lots_id = ?
																						 and status = ?
																						 order by created_at asc, parking_spaces.id asc
																						 limit 1`)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, parkingLotID, parkingLotID, available)
	if row == nil {
		return 0, ErrNilQueryRowContext
	}

	err = row.Err()
	if err != nil {
		return 0, err
	}

	var id int
	if err := row.Scan(&id); err != nil {
		return 0, err
	}

	return id, nil
}

func (d *DB) DoesParkingSpaceExistForMaintananceByParkingLotIDAndID(id, parkingLotID int) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt, err := d.dbConn.PrepareContext(ctx, `SELECT EXISTS(
																							SELECT id
																							FROM parking_spaces
																							WHERE EXISTS (
																								SELECT parking_lots.id
																								FROM parking_lots
																								where parking_lots.id = ?
																							) and parking_lots_id = ?
																							and parking_spaces.id = ?
																							and status != ?
																							limit 1
																						) limit 1`)

	if err != nil {
		return false, err
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, parkingLotID, parkingLotID, id, booked)
	if row == nil {
		return false, ErrNilQueryRowContext
	}

	err = row.Err()
	if err != nil {
		return false, err
	}

	var count int
	if err := row.Scan(&count); err != nil {
		return false, err
	}

	return count > 0, nil
}

func (d *DB) SetParkingSpaceMaintanance(id int, m bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt, err := d.dbConn.PrepareContext(ctx, `UPDATE parking_spaces SET status = ? WHERE (id = ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	status := available
	if m {
		status = maintanance
	}
	_, err = stmt.ExecContext(ctx, status, id)

	return err
}
