package seeders

import (
	"log/slog"

	"github.com/bcc-media/wayfarer/internal/ulid"
)

// SeedChurches creates churches with varied sizes and countries
func (s *Seeder) SeedChurches(stats *Stats) error {
	countries := []string{"Norway", "Sweden", "Denmark", "Finland", "Netherlands", "Germany", "United Kingdom", "United States"}
	categories := []string{"S", "L", "XL"}
	categoryWeights := []int{5, 3, 2} // 50% S, 30% L, 20% XL

	query := `
		INSERT INTO churches (id, name, country, category)
		VALUES ($1, $2, $3, $4)
	`

	for i := 0; i < s.Config.NumChurches; i++ {
		id := ulid.NewChurchID()
		country := countries[s.Fake.IntBetween(0, len(countries)-1)]

		// Weighted random category selection
		categoryIndex := s.Fake.IntBetween(1, 10)
		var category string
		if categoryIndex <= categoryWeights[0] {
			category = categories[0] // S
		} else if categoryIndex <= categoryWeights[0]+categoryWeights[1] {
			category = categories[1] // L
		} else {
			category = categories[2] // XL
		}

		name := s.Fake.Company().Name() + " Church"

		_, err := s.DB.Pool.Exec(s.Ctx, query, id, name, country, category)
		if err != nil {
			return err
		}
		s.Data.ChurchIDs = append(s.Data.ChurchIDs, id)
		stats.Churches++

		if (i+1)%1000 == 0 {
			slog.Info("Churches progress", "created", i+1, "total", s.Config.NumChurches)
		}
	}

	return nil
}
