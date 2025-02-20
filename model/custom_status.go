// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package model

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

const (
	UserPropsKeyCustomStatus = "customStatus"

	CustomStatusTextMaxRunes = 100
	MaxRecentCustomStatuses  = 5
	DefaultCustomStatusEmoji = "speech_balloon"
)

var validCustomStatusDuration = map[string]bool{
	"thirty_minutes": true,
	"one_hour":       true,
	"four_hours":     true,
	"today":          true,
	"this_week":      true,
	"date_and_time":  true,
}

type CustomStatus struct {
	Emoji     string    `json:"emoji"`
	Text      string    `json:"text"`
	Duration  string    `json:"duration"`
	ExpiresAt time.Time `json:"expires_at"`
}

func (cs *CustomStatus) PreSave() {
	if cs.Emoji == "" {
		cs.Emoji = DefaultCustomStatusEmoji
	}

	if cs.Duration == "" && !cs.ExpiresAt.Before(time.Now()) {
		cs.Duration = "date_and_time"
	}

	runes := []rune(cs.Text)
	if len(runes) > CustomStatusTextMaxRunes {
		cs.Text = string(runes[:CustomStatusTextMaxRunes])
	}
}

func (cs *CustomStatus) ToJson() string {
	csCopy := *cs
	b, _ := json.Marshal(csCopy)
	return string(b)
}

func (cs *CustomStatus) AreDurationAndExpirationTimeValid() bool {
	if cs.Duration == "" && (cs.ExpiresAt.IsZero() || !cs.ExpiresAt.Before(time.Now())) {
		return true
	}

	if validCustomStatusDuration[cs.Duration] && !cs.ExpiresAt.Before(time.Now()) {
		return true
	}

	return false
}

func CustomStatusFromJson(data io.Reader) *CustomStatus {
	var cs *CustomStatus
	_ = json.NewDecoder(data).Decode(&cs)
	return cs
}

func RuneToHexadecimalString(r rune) string {
	return fmt.Sprintf("%04x", r)
}

type RecentCustomStatuses []CustomStatus

func (rcs RecentCustomStatuses) Contains(cs *CustomStatus) bool {
	var csJSON = cs.ToJson()

	// status is empty
	if cs == nil || csJSON == "" || (cs.Emoji == "" && cs.Text == "") {
		return false
	}

	for _, status := range rcs {
		if status.ToJson() == csJSON {
			return true
		}
	}

	return false
}

func (rcs RecentCustomStatuses) Add(cs *CustomStatus) RecentCustomStatuses {
	newRCS := rcs[:0]

	// if same `text` exists in existing recent custom statuses, modify existing status
	for _, status := range rcs {
		if status.Text != cs.Text {
			newRCS = append(newRCS, status)
		}
	}
	newRCS = append(RecentCustomStatuses{*cs}, newRCS...)
	if len(newRCS) > MaxRecentCustomStatuses {
		newRCS = newRCS[:MaxRecentCustomStatuses]
	}
	return newRCS
}

func (rcs RecentCustomStatuses) Remove(cs *CustomStatus) RecentCustomStatuses {
	var csJSON = cs.ToJson()
	if csJSON == "" || (cs.Emoji == "" && cs.Text == "") {
		return rcs
	}

	newRCS := rcs[:0]
	for _, status := range rcs {
		if status.ToJson() != csJSON {
			newRCS = append(newRCS, status)
		}
	}

	return newRCS
}

func (rcs RecentCustomStatuses) ToJson() string {
	b, _ := json.Marshal(rcs)
	return string(b)
}

func RecentCustomStatusesFromJson(data io.Reader) RecentCustomStatuses {
	var rcs RecentCustomStatuses
	_ = json.NewDecoder(data).Decode(&rcs)
	return rcs
}
