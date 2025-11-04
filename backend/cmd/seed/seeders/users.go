package seeders

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/bcc-media/wayfarer/internal/ulid"
)

// SeedUsers creates 75 users distributed across churches
func (s *Seeder) SeedUsers(stats *Stats) error {
	numUsers := 75

	query := `
		INSERT INTO users (id, members_id, email, name, gender, birthdate, church_id, avatar_url)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	genders := []string{"MALE", "FEMALE"}

	for i := 0; i < numUsers; i++ {
		// Random gender
		gender := genders[rand.Intn(len(genders))]

		// Generate realistic name based on gender
		var name string
		if gender == "MALE" {
			name = s.Fake.Person().FirstNameMale() + " " + s.Fake.Person().LastName()
		} else {
			name = s.Fake.Person().FirstNameFemale() + " " + s.Fake.Person().LastName()
		}

		// Random age between 13 and 80 - convert to birthdate
		age := 13 + rand.Intn(68)
		now := time.Now()
		birthdate := now.AddDate(-age, -rand.Intn(12), -rand.Intn(28))

		// Random church
		churchID := s.Data.ChurchIDs[rand.Intn(len(s.Data.ChurchIDs))]

		// Generate members ID (fake external ID)
		membersID := fmt.Sprintf("MEM-%d", 10000+i)

		// Generate email
		email := s.Fake.Internet().Email()

		// Avatar URL (use placeholder service)
		avatarURL := fmt.Sprintf("https://i.pravatar.cc/150?img=%d", i+1)

		id := ulid.NewUserID()
		_, err := s.DB.Pool.Exec(s.Ctx, query,
			id,
			membersID,
			email,
			name,
			gender,
			birthdate,
			churchID,
			avatarURL,
		)
		if err != nil {
			return err
		}

		s.Data.UserIDs = append(s.Data.UserIDs, id)
		stats.Users++
	}

	return nil
}
