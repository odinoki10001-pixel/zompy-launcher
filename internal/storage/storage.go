package storage

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

type Build struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Emoji   string   `json:"emoji"`
	Thumb   string   `json:"thumb"`
	Loader  string   `json:"loader"`
	Version string   `json:"version"`
	Mods    []string `json:"mods"`
	Tag     string   `json:"tag"`
	Desc    string   `json:"desc"`
	Fav     bool     `json:"fav"`
	Ram     int      `json:"ram"`
	Created int64    `json:"created"`
}

type Store struct {
	mu    sync.RWMutex
	file  string
	items []Build
}

func New(dataDir string) *Store {
	s := &Store{file: filepath.Join(dataDir, "builds.json")}
	s.load()
	if len(s.items) == 0 {
		s.items = seed()
		_ = s.flush()
	}
	return s
}

func (s *Store) load() {
	data, err := os.ReadFile(s.file)
	if err != nil {
		return
	}
	_ = json.Unmarshal(data, &s.items)
}

func (s *Store) flush() error {
	if err := os.MkdirAll(filepath.Dir(s.file), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s.items, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.file, data, 0o644)
}

func (s *Store) Builds() []Build {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Build, len(s.items))
	copy(out, s.items)
	return out
}

func (s *Store) GetBuild(id string) (Build, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, b := range s.items {
		if b.ID == id {
			return b, nil
		}
	}
	return Build{}, errors.New("not found")
}

func (s *Store) SaveBuild(b Build) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.items {
		if s.items[i].ID == b.ID {
			s.items[i] = b
			return s.flush()
		}
	}
	s.items = append(s.items, b)
	return s.flush()
}

func (s *Store) DeleteBuild(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.items {
		if s.items[i].ID == id {
			s.items = append(s.items[:i], s.items[i+1:]...)
			return s.flush()
		}
	}
	return nil
}

func seed() []Build {
	return []Build{
		{ID: "1", Name: "Vanilla Plus", Emoji: "⛏️", Thumb: "t1", Loader: "vanilla", Version: "1.21.4", Tag: "Популярное", Desc: "Чистая ваниль с улучшениями", Fav: true, Ram: 4},
		{ID: "2", Name: "Oceancraft", Emoji: "🌊", Thumb: "t2", Loader: "forge", Version: "1.20.1", Desc: "Подводные города и биомы", Ram: 6},
		{ID: "3", Name: "Magic Realms", Emoji: "🔮", Thumb: "t3", Loader: "forge", Version: "1.19.2", Tag: "Хардкор", Desc: "Магия и ритуалы", Ram: 8},
		{ID: "4", Name: "Survival Hardcore", Emoji: "⚔️", Thumb: "t4", Loader: "fabric", Version: "1.21", Tag: "Новое", Desc: "Реалистичное выживание", Fav: true, Ram: 6},
	}
}
