package main

import (
	"context"
	"fmt"
	"sort"

	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/jackc/pgx/v5"
)

// ChurchSession is one load-test church for k6/betting-multisession.js: the
// church admin who runs its quiz session and the players who get access to it.
type ChurchSession struct {
	ChurchID string         `json:"churchId"`
	Admin    UserToken      `json:"admin"`
	Players  []ChurchPlayer `json:"players"`
}

// ChurchPlayer is a player token, with a simulated Auth0 token when one was
// minted for the user (REALISM auth dance)
type ChurchPlayer struct {
	UserID     string `json:"userId"`
	Token      string `json:"token"`
	Auth0Token string `json:"auth0Token,omitempty"`
}

type churchMember struct {
	ChurchID string
	UserID   string
}

// loadChurchMembers reads the load-test churches made by
// bench/betting_box.sh reset-multi: their CHURCH_ADMINs, and their members in
// the project (the users a church grant gives access to).
func loadChurchMembers(ctx context.Context, db *database.DB, churchPrefix, projectID string) (admins, players []churchMember, err error) {
	admins, err = queryMembers(ctx, db, `
		SELECT church_id, user_id FROM user_roles
		WHERE role = 'CHURCH_ADMIN' AND church_id LIKE @prefix || '%'
		ORDER BY church_id, user_id`, pgx.NamedArgs{"prefix": churchPrefix})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query church admins: %w", err)
	}
	players, err = queryMembers(ctx, db, `
		SELECT u.church_id, u.id FROM users u
		JOIN user_projects up ON up.user_id = u.id AND up.project_id = @project
		WHERE u.church_id LIKE @prefix || '%'
		ORDER BY u.church_id, u.id`, pgx.NamedArgs{"prefix": churchPrefix, "project": projectID})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query church players: %w", err)
	}
	return admins, players, nil
}

func queryMembers(ctx context.Context, db *database.DB, sql string, args pgx.NamedArgs) ([]churchMember, error) {
	rows, err := db.Pool.Query(ctx, sql, args)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var members []churchMember
	for rows.Next() {
		var m churchMember
		if err := rows.Scan(&m.ChurchID, &m.UserID); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, rows.Err()
}

// buildChurchSessions groups players under their church's admin, ordered by
// church ID, minting a token per user. auth0 maps user IDs to simulated Auth0
// tokens (may be nil). A church needs exactly one admin and at least one
// player: anything else means the fixture is broken, and the run would be
// measuring a different scenario.
func buildChurchSessions(admins, players []churchMember, mint func(userID string) (string, error), auth0 map[string]string) ([]ChurchSession, error) {
	byChurch := make(map[string]*ChurchSession, len(admins))
	for _, a := range admins {
		if _, dup := byChurch[a.ChurchID]; dup {
			return nil, fmt.Errorf("church %s has more than one admin", a.ChurchID)
		}
		token, err := mint(a.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to mint admin token for %s: %w", a.UserID, err)
		}
		byChurch[a.ChurchID] = &ChurchSession{ChurchID: a.ChurchID, Admin: UserToken{UserID: a.UserID, Token: token}}
	}
	for _, p := range players {
		church, ok := byChurch[p.ChurchID]
		if !ok {
			return nil, fmt.Errorf("church %s has players but no admin", p.ChurchID)
		}
		token, err := mint(p.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to mint player token for %s: %w", p.UserID, err)
		}
		church.Players = append(church.Players, ChurchPlayer{UserID: p.UserID, Token: token, Auth0Token: auth0[p.UserID]})
	}

	sessions := make([]ChurchSession, 0, len(byChurch))
	for _, c := range byChurch {
		if len(c.Players) == 0 {
			return nil, fmt.Errorf("church %s has an admin but no players", c.ChurchID)
		}
		sessions = append(sessions, *c)
	}
	sort.Slice(sessions, func(i, j int) bool { return sessions[i].ChurchID < sessions[j].ChurchID })
	return sessions, nil
}
