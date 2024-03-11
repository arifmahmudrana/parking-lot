package db

import (
	"context"
)

type ParkingLot struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (d *DB) CreateParkingLot(pl ParkingLot) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt, err := d.dbConn.PrepareContext(ctx, `insert into parking_lots (name) values (?)`)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, pl.Name)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (d *DB) GetParkingLots(page int) ([]ParkingLot, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt, err := d.dbConn.PrepareContext(ctx, `select id, name
																             from parking_lots
																						 LIMIT ?, ?`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, getOffset(page), size)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	parkingLots := make([]ParkingLot, 0, size)
	for rows.Next() {
		var parkingLot ParkingLot
		err := rows.Scan(
			&parkingLot.ID,
			&parkingLot.Name,
		)
		if err != nil {
			return nil, err
		}

		parkingLots = append(parkingLots, parkingLot)
	}

	return parkingLots, nil
}

func (d *DB) GetTotalCountParkingLots() (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `select count(id)
					  from parking_lots
						limit 1`

	row := d.dbConn.QueryRowContext(ctx, query)
	if row == nil {
		return 0, ErrNilQueryRowContext
	}

	err := row.Err()
	if err != nil {
		return 0, err
	}

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

func (d *DB) DoesParkingLotExistByID(parkingLotID int) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt, err := d.dbConn.PrepareContext(ctx, `SELECT EXISTS(
							SELECT id FROM parking_lots WHERE parking_lots.id = ?
						) limit 1`)

	if err != nil {
		return false, err
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, parkingLotID)
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
