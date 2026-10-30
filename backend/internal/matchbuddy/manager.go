package matchbuddy

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/inox/inox/backend/internal/ws"
)

func mustMarshal(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		return json.RawMessage("{}")
	}
	return json.RawMessage(b)
}

// Manager orchestrates RoomEngine instances for every room in the application.
type Manager struct {
	mu          sync.RWMutex
	engines     map[string]*RoomEngine
	commentator Commentator
	broadcast   func(*ws.Event)
}

// NewManager creates a match buddy manager.
func NewManager(commentator Commentator, broadcast func(*ws.Event)) *Manager {
	if commentator == nil {
		commentator = NewTemplateCommentator()
	}
	return &Manager{
		engines:     make(map[string]*RoomEngine),
		commentator: commentator,
		broadcast:   broadcast,
	}
}

// GetOrCreateEngine returns the RoomEngine for a given room, initializing it if absent.
func (m *Manager) GetOrCreateEngine(roomID string) *RoomEngine {
	m.mu.Lock()
	defer m.mu.Unlock()

	engine, exists := m.engines[roomID]
	if !exists {
		engine = NewRoomEngine(roomID, m.commentator, m.broadcast)
		m.engines[roomID] = engine
	}
	return engine
}

// StopRoom stops and removes a room's engine when the room closes or is emptied.
func (m *Manager) StopRoom(roomID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if engine, exists := m.engines[roomID]; exists {
		engine.Stop()
		delete(m.engines, roomID)
		slog.Info("stopped match buddy engine for room", "room_id", roomID)
	}
}

// GetInitialState returns the initial MATCH_STATE event for a room.
func (m *Manager) GetInitialState(roomID string) *ws.Event {
	engine := m.GetOrCreateEngine(roomID)
	snap := engine.Snapshot()
	return &ws.Event{
		Type:      ws.EventType("MATCH_STATE"),
		RoomID:    roomID,
		Timestamp: time.Now().UnixMilli(),
		Payload:   mustMarshal(snap),
	}
}

// HandleEvent processes inbound client MATCH_* events from the Hub.
func (m *Manager) HandleEvent(event *ws.Event) {
	engine := m.GetOrCreateEngine(event.RoomID)

	switch event.Type {
	case ws.EventType("MATCH_CONTROL"):
		var p struct {
			Action string   `json:"action"`
			Speed  *float64 `json:"speed,omitempty"`
		}
		if err := json.Unmarshal(event.Payload, &p); err != nil {
			slog.Warn("malformed MATCH_CONTROL payload", "room_id", event.RoomID, "err", err)
			return
		}

		switch p.Action {
		case "start":
			if err := engine.Start(); err != nil {
				m.sendError(event.RoomID, event.SenderID, err.Error())
			}
		case "pause":
			engine.Pause()
		case "resume":
			engine.Resume()
		case "set_speed":
			if p.Speed != nil {
				engine.SetSpeed(*p.Speed)
			}
		case "skip_over":
			engine.SkipToNextOver()
		case "reset":
			engine.Reset()
		case "try_solo":
			engine.ToggleSolo()
		}

	case ws.EventType("MATCH_TEAM_SET"):
		var p struct {
			PlayerIDs     []string `json:"player_ids"`
			CaptainID     string   `json:"captain_id"`
			ViceCaptainID string   `json:"vice_captain_id"`
			PresetID      string   `json:"preset_id"`
		}
		if err := json.Unmarshal(event.Payload, &p); err != nil {
			slog.Warn("malformed MATCH_TEAM_SET payload", "room_id", event.RoomID, "err", err)
			return
		}

		err := engine.SetTeam(event.SenderID, event.SenderName, p.PlayerIDs, p.CaptainID, p.ViceCaptainID, p.PresetID)
		if err != nil {
			m.sendError(event.RoomID, event.SenderID, err.Error())
		}

	case ws.EventType("MATCH_WHATIF"):
		var p struct {
			Scenario string `json:"scenario"`
		}
		if err := json.Unmarshal(event.Payload, &p); err != nil {
			return
		}
		proj, err := engine.WhatIf(p.Scenario)
		if err != nil {
			m.sendError(event.RoomID, event.SenderID, err.Error())
			return
		}

		card := FeedCard{
			ID:        fmt.Sprintf("feed-whatif-%d", time.Now().UnixMilli()),
			Kind:      FeedKindWhatIf,
			Headline:  fmt.Sprintf("🔮 What If: %s next ball?", p.Scenario),
			BodyText:  proj.Explanation,
			Timestamp: time.Now().UnixMilli(),
		}
		engine.broadcastFeed(card)

	case ws.EventType("MATCH_ASK"):
		var p struct {
			Question string `json:"question"`
		}
		if err := json.Unmarshal(event.Payload, &p); err != nil {
			return
		}
		answer, err := engine.Ask(p.Question)
		if err != nil {
			answer = "I'm keeping my eye on the pitch right now! Check the scoreboard above for live stats."
		}

		card := FeedCard{
			ID:        fmt.Sprintf("feed-ans-%d", time.Now().UnixMilli()),
			Kind:      FeedKindAnswer,
			Headline:  fmt.Sprintf("💬 Q: %s", p.Question),
			BodyText:  answer,
			Timestamp: time.Now().UnixMilli(),
		}
		engine.broadcastFeed(card)

	case ws.EventType("MATCH_CATCHUP"):
		catchUp, err := engine.CatchUp()
		if err != nil {
			catchUp = "Match Buddy is running live! Check the leaderboard for current fantasy totals."
		}

		card := FeedCard{
			ID:        fmt.Sprintf("feed-catchup-%d", time.Now().UnixMilli()),
			Kind:      FeedKindCatchUp,
			Headline:  "⚡ Quick Catch-Up",
			BodyText:  catchUp,
			Timestamp: time.Now().UnixMilli(),
		}
		engine.broadcastFeed(card)
	}
}

func (m *Manager) sendError(roomID, targetID, errMsg string) {
	evt := &ws.Event{
		Type:      ws.EventError,
		RoomID:    roomID,
		TargetID:  targetID,
		Timestamp: time.Now().UnixMilli(),
		Payload:   mustMarshal(map[string]string{"message": errMsg}),
	}
	m.broadcast(evt)
}
