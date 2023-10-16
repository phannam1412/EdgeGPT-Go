package EdgeGPT

import (
	"errors"
	"fmt"
	"github.com/pavel-one/EdgeGPT-Go/config"
	"time"
)

// Storage for GPT sessions. Use ony for servers
type Storage map[string]*GPT

func NewStorage() *Storage {
	return &Storage{}
}

// GetOrSet get current session, or create new
func (s *Storage) GetOrSet(key string) (*GPT, error) {
	var gpt *GPT

	gpt, err := s.Get(key)
	if err == nil {
		return gpt, nil
	}

	conf, err := config.NewGpt()
	if err != nil {
		return nil, fmt.Errorf("didn't create GPTConfig config: %s", err)
	}

	gpt, err = NewGPT(conf)
	if err != nil {
		return nil, fmt.Errorf("didn't init GPTConfig service: %s", err)
	}

	s.Add(gpt, key)

	return gpt, nil
}

// Add new session
func (s *Storage) Add(gpt *GPT, key string) {
	log.Infoln("New store key:", key)
	(*s)[key] = gpt
}

// Get get current session, or error
func (s *Storage) Get(key string) (*GPT, error) {
	v, ok := (*s)[key]
	if !ok {
		return nil, errors.New("not set session")
	}

	if time.Now().After(v.ExpiredAt) {
		if err := s.Remove(key); err != nil {
			return nil, err
		}
		return nil, errors.New("session is expired")
	}

	log.Infoln("Get store:", key)
	return v, nil
}

// Remove session
func (s *Storage) Remove(key string) error {
	log.Infoln("Store remove:", key)

	so := *s
	_, ok := so[key]
	if !ok {
		return errors.New("not set session")
	}

	delete(so, key)

	return nil
}
