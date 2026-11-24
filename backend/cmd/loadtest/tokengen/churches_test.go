package main

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func fakeMint(userID string) (string, error) { return "token-" + userID, nil }

func TestBuildChurchSessions_GroupsPlayersUnderTheirAdmin(t *testing.T) {
	admins := []churchMember{{ChurchID: "CH2", UserID: "A2"}, {ChurchID: "CH1", UserID: "A1"}}
	players := []churchMember{
		{ChurchID: "CH1", UserID: "P1"},
		{ChurchID: "CH2", UserID: "P2"},
		{ChurchID: "CH1", UserID: "P3"},
	}

	sessions, err := buildChurchSessions(admins, players, fakeMint, map[string]string{"P3": "auth0-P3"})
	require.NoError(t, err)

	assert.Equal(t, []ChurchSession{
		{
			ChurchID: "CH1",
			Admin:    UserToken{UserID: "A1", Token: "token-A1"},
			Players: []ChurchPlayer{
				{UserID: "P1", Token: "token-P1"},
				{UserID: "P3", Token: "token-P3", Auth0Token: "auth0-P3"},
			},
		},
		{
			ChurchID: "CH2",
			Admin:    UserToken{UserID: "A2", Token: "token-A2"},
			Players:  []ChurchPlayer{{UserID: "P2", Token: "token-P2"}},
		},
	}, sessions)
}

func TestBuildChurchSessions_RejectsABrokenFixture(t *testing.T) {
	tests := []struct {
		name    string
		admins  []churchMember
		players []churchMember
		want    string
	}{
		{
			name:    "players without an admin",
			admins:  []churchMember{{ChurchID: "CH1", UserID: "A1"}},
			players: []churchMember{{ChurchID: "CH1", UserID: "P1"}, {ChurchID: "CH2", UserID: "P2"}},
			want:    "church CH2 has players but no admin",
		},
		{
			name:    "admin without players",
			admins:  []churchMember{{ChurchID: "CH1", UserID: "A1"}, {ChurchID: "CH2", UserID: "A2"}},
			players: []churchMember{{ChurchID: "CH1", UserID: "P1"}},
			want:    "church CH2 has an admin but no players",
		},
		{
			name:    "two admins for one church",
			admins:  []churchMember{{ChurchID: "CH1", UserID: "A1"}, {ChurchID: "CH1", UserID: "A2"}},
			players: []churchMember{{ChurchID: "CH1", UserID: "P1"}},
			want:    "church CH1 has more than one admin",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := buildChurchSessions(tt.admins, tt.players, fakeMint, nil)
			assert.EqualError(t, err, tt.want)
		})
	}
}

func TestBuildChurchSessions_MintFailure(t *testing.T) {
	failing := func(string) (string, error) { return "", errors.New("boom") }
	_, err := buildChurchSessions(
		[]churchMember{{ChurchID: "CH1", UserID: "A1"}},
		[]churchMember{{ChurchID: "CH1", UserID: "P1"}},
		failing, nil)
	assert.ErrorContains(t, err, "boom")
}
