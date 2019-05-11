package memory

import (
	"sync"
	"time"

	"github.com/syaiful6/sersan"
)

type sessionStore struct {
	sync.Mutex
	sessions map[string]*sersan.Session
}

func (s *sessionStore) Get(id string) (*sersan.Session, error) {
	s.Lock()
	defer s.Unlock()

	if v, ok := s.sessions[id]; ok {
		if v.IsSessionExpired(604800, 5184000, time.Now().UTC()) {
			delete(s.sessions, id)
			return nil, nil
		}
		return v, nil
	}

	return nil, nil
}

func (s *sessionStore) Destroy(id string) error {
	s.Lock()
	defer s.Unlock()

	if _, ok := s.sessions[id]; ok {
		delete(s.sessions, id)
	}

	return nil
}

func (s *sessionStore) DestroyAllOfAuthId(authId string) error {
	s.Lock()
	defer s.Unlock()
	nmap := make(map[string]*sersan.Session)
	for k, sess := range s.sessions {
		if sess.AuthID != authId {
			nmap[k] = sess
		}
	}
	s.sessions = nmap

	return nil
}

func (s *sessionStore) Insert(sess *sersan.Session) error {
	s.Lock()
	defer s.Unlock()

	if _, ok := s.sessions[sess.ID]; ok {
		return sersan.SessionAlreadyExists{ID: sess.ID}
	}

	s.sessions[sess.ID] = sess
	return nil
}

func (s *sessionStore) Replace(sess *sersan.Session) error {
	s.Lock()
	defer s.Unlock()

	if _, ok := s.sessions[sess.ID]; ok {
		s.sessions[sess.ID] = sess
		return nil
	}

	return sersan.SessionDoesNotExist{ID: sess.ID}
}
