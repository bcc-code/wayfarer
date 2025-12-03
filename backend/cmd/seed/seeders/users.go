package seeders

import (
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/bcc-media/wayfarer/internal/ulid"
	"github.com/jackc/pgx/v5"
)

// SeedUsers creates users distributed across churches
func (s *Seeder) SeedUsers(stats *Stats) error {
	genders := []string{"MALE", "FEMALE"}
	batchSize := 1000

	for batch := 0; batch < s.Config.NumUsers/batchSize; batch++ {
		batchRows := [][]interface{}{}

		for i := 0; i < batchSize; i++ {
			userIndex := batch*batchSize + i

			// Random gender
			gender := genders[rand.Intn(len(genders))]

			// Generate realistic name based on gender
			var firstName string
			var lastName string
			if gender == "MALE" {
				firstName = s.Fake.Person().FirstNameMale()
			} else {
				firstName = s.Fake.Person().FirstNameFemale()
			}
			lastName = s.Fake.Person().LastName()
			displayName := firstName + " " + lastName

			// Random age between 13 and 80 - convert to birthdate
			age := 13 + rand.Intn(68)
			now := time.Now()
			birthdate := now.AddDate(-age, -rand.Intn(12), -rand.Intn(28))

			// Random church
			churchID := s.Data.ChurchIDs[rand.Intn(len(s.Data.ChurchIDs))]

			// Generate members ID (fake external ID) - start from high number to avoid conflicts
			membersID := fmt.Sprintf("MEM-%d", 1000000+userIndex)

			// Generate email
			email := s.Fake.Internet().Email()

			// Avatar URL (cycle through pravatar images)
			avatarURL := fmt.Sprintf("https://i.pravatar.cc/150?img=%d", (userIndex%70)+1)

			id := ulid.NewUserID()
			batchRows = append(batchRows, []interface{}{id, membersID, email, displayName, firstName, lastName, nil, displayName, gender, birthdate, churchID, avatarURL})
			s.Data.UserIDs = append(s.Data.UserIDs, id)
		}

		// Batch insert
		_, err := s.DB.Pool.CopyFrom(
			s.Ctx,
			pgx.Identifier{"users"},
			[]string{"id", "members_id", "email", "name", "first_name", "last_name", "middle_name", "display_name", "gender", "birthdate", "church_id", "avatar_url"},
			pgx.CopyFromRows(batchRows),
		)
		if err != nil {
			return err
		}

		stats.Users += batchSize
		slog.Info("Users progress", "created", stats.Users, "total", s.Config.NumUsers)
	}

	// Handle remaining users if not exact multiple of batch size
	remaining := s.Config.NumUsers % batchSize
	if remaining > 0 {
		batchRows := [][]interface{}{}

		for i := 0; i < remaining; i++ {
			userIndex := (s.Config.NumUsers/batchSize)*batchSize + i

			gender := genders[rand.Intn(len(genders))]
			var firstName string
			var lastName string
			if gender == "MALE" {
				firstName = s.Fake.Person().FirstNameMale()
			} else {
				firstName = s.Fake.Person().FirstNameFemale()
			}
			lastName = s.Fake.Person().LastName()
			displayName := firstName + " " + lastName

			age := 13 + rand.Intn(68)
			now := time.Now()
			birthdate := now.AddDate(-age, -rand.Intn(12), -rand.Intn(28))
			churchID := s.Data.ChurchIDs[rand.Intn(len(s.Data.ChurchIDs))]
			membersID := fmt.Sprintf("MEM-%d", 1000000+userIndex)
			email := s.Fake.Internet().Email()
			avatarURL := fmt.Sprintf("https://i.pravatar.cc/150?img=%d", (userIndex%70)+1)

			id := ulid.NewUserID()
			batchRows = append(batchRows, []interface{}{id, membersID, email, displayName, firstName, lastName, nil, displayName, gender, birthdate, churchID, avatarURL})
			s.Data.UserIDs = append(s.Data.UserIDs, id)
		}

		_, err := s.DB.Pool.CopyFrom(
			s.Ctx,
			pgx.Identifier{"users"},
			[]string{"id", "members_id", "email", "name", "first_name", "last_name", "middle_name", "display_name", "gender", "birthdate", "church_id", "avatar_url"},
			pgx.CopyFromRows(batchRows),
		)
		if err != nil {
			return err
		}

		stats.Users += remaining
		slog.Info("Users completed", "total", stats.Users)
	}

	return nil
}
