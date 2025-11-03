package seeders

import (
	"github.com/bcc-media/wayfarer/internal/ulid"
)

// SeedChurches creates 8 churches with varied sizes and countries
func (s *Seeder) SeedChurches(stats *Stats) error {

	churches := []struct {
		name     string
		country  string
		category string
	}{
		{"Oslo Community Church", "Norway", "L"},
		{"Bergen Fellowship", "Norway", "S"},
		{"Stavanger Baptist Church", "Norway", "S"},
		{"Stockholm Evangelical Church", "Sweden", "XL"},
		{"Göteborg Christian Center", "Sweden", "L"},
		{"Copenhagen Grace Church", "Denmark", "L"},
		{"Aarhus Family Church", "Denmark", "S"},
		{"Helsinki Gospel Church", "Finland", "XL"},
	}

	query := `
		INSERT INTO churches (id, name, country, category)
		VALUES ($1, $2, $3, $4)
	`

	for _, church := range churches {
		id := ulid.NewChurchID()
		_, err := s.DB.Pool.Exec(s.Ctx, query, id, church.name, church.country, church.category)
		if err != nil {
			return err
		}
		s.Data.ChurchIDs = append(s.Data.ChurchIDs, id)
		stats.Churches++
	}

	return nil
}
