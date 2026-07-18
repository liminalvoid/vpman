package repository

import (
	"context"
	"fmt"

	"vpnmanager/internal/model"
)

type Server struct {
	conn conn
}

func NewServer(conn conn) *Server { return &Server{conn} }

func (r *Server) Create(ctx context.Context, server *model.Server) error {
	const q = `
		INSERT INTO servers (public_address, internal_address, location, location_country, location_city)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, public_address, internal_address, location, location_country, location_city`

	return r.conn.
		QueryRow(
			ctx,
			q,
			server.PublicAddress,
			server.InternalAddress,
			server.Location,
			server.LocationCountry,
			server.LocationCity,
		).
		Scan(
			&server.ID,
			&server.PublicAddress,
			&server.InternalAddress,
			&server.Location,
			&server.LocationCountry,
			&server.LocationCity,
		)
}

func (r *Server) ListAll(ctx context.Context) ([]model.Server, error) {
	const q = `
		SELECT id, public_address, internal_address, location, location_country, location_city
		FROM servers`

	rows, err := r.conn.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("querying servers: %w", err)
	}
	defer rows.Close()

	var out []model.Server
	for rows.Next() {
		var s model.Server
		if err := rows.Scan(
			&s.ID,
			&s.PublicAddress,
			&s.InternalAddress,
			&s.Location,
			&s.LocationCountry,
			&s.LocationCity,
		); err != nil {
			return nil, fmt.Errorf("scanning server: %w", err)
		}

		out = append(out, s)
	}

	return out, rows.Err()
}
