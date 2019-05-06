package cassandra

import (
	"bytes"
	"context"
	"encoding/gob"
	"time"

	"github.com/gocql/gocql"
	"github.com/syaiful6/sersan"
	"github.com/thatique/kuade/pkg/queue"
)

const (
	checkExistense       = `SELECT count(*) FROM sessions WHERE id = ?`
	selectSessionAuthID  = `SELECT auth_id FROM sessions WHERE id = ?`
	selectIDsByAuthID    = `SELECT id FROM sessions_auth_index WHERE auth_id = ?`
	selectSessionByID    = `SELECT auth_id, values, created_at, accessed_at FROM sessions WHERE id = ?`
	updateSession        = `UPDATE sessions SET auth_id = ?, values = ?, created_at = ?, accessed_at = ? WHERE id = ?`
	updateAuthSession    = `INSERT INTO sessions_auth_index(auth_id, id) VALUES(?, ?)`
	deleteSession        = `DELETE FROM sessions WHERE id = ?`
	deleteAuthSession    = `DELETE FROM sessions_auth_index WHERE auth_id = ? AND id = ?`
	deleteAllAuthSession = `DELETE FROM sessions_auth_index WHERE auth_id = ?`
)

type deleteSessionJob struct {
	ID string
}

func (d *deleteSessionJob) Fire(sess *gocql.Session) error {
	query := sess.Query(deleteSession, d.ID)
	defer query.Release()

	return query.Exec()
}

type sessionStore struct {
	session *gocql.Session
	queue   *queue.BoundedQueue
}

func (s *sessionStore) Get(id string) (session *sersan.Session, err error) {
	var (
		authID     string
		values     []byte
		createdAt  time.Time
		accessedAt time.Time
	)
	query := s.session.Query(selectSessionByID, id)
	defer query.Release()

	if err = query.Scan(
		&authID,
		&values,
		&createdAt,
		&accessedAt,
	); err != nil {
		if err == gocql.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}

	session = sersan.NewSession(id, authID, createdAt)
	session.AccessedAt = accessedAt

	dec := gob.NewDecoder(bytes.NewBuffer(values))
	if err = dec.Decode(&session.Values); err != nil {
		return nil, err
	}

	return session, nil
}

func (s *sessionStore) Destroy(id string) (err error) {
	var authID string
	queryAuthID := s.session.Query(selectSessionAuthID, id)
	defer queryAuthID.Release()

	if err = queryAuthID.Scan(&authID); err != nil {
		if err == gocql.ErrNotFound {
			return nil
		}
		return err
	}

	if err = s.deleteSession(id); err != nil && err != gocql.ErrNotFound {
		return err
	}

	if authID != "" {
		if err = s.deleteAuthSession(authID, id); err != nil && err != gocql.ErrNotFound {
			return err
		}
	}

	return nil
}

func (s *sessionStore) DestroyAllOfAuthId(authID string) error {
	var id string
	iter := s.session.Query(selectIDsByAuthID, authID).Iter()

	for iter.Scan(&id) {
		s.queue.Produce(&deleteSessionJob{ID: id})
	}
	if err := iter.Close(); err != nil {
		return err
	}
	// delete this index as well, but execute this without gourotine
	query := s.session.Query(deleteAllAuthSession, authID)
	defer query.Release()

	return query.Exec()
}

func (s *sessionStore) Insert(sess *sersan.Session) (err error) {
	if exist := s.isExists(context.Background(), sess.ID); exist {
		return sersan.SessionAlreadyExists{ID: sess.ID}
	}

	return s.putSession(sess)
}

func (s *sessionStore) Replace(sess *sersan.Session) (err error) {
	if exist := s.isExists(context.Background(), sess.ID); !exist {
		return sersan.SessionDoesNotExist{ID: sess.ID}
	}

	return s.putSession(sess)
}

func (s *sessionStore) putSession(sess *sersan.Session) (err error) {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	if err = enc.Encode(sess.Values); err != nil {
		return err
	}

	query := s.session.Query(updateSession, sess.AuthID, buf.Bytes(), sess.CreatedAt, sess.AccessedAt, sess.ID)
	defer query.Release()
	if err = query.Exec(); err != nil {
		return err
	}

	if sess.AuthID != "" {
		idxSessQuery := s.session.Query(updateAuthSession, sess.AuthID, sess.ID)
		defer idxSessQuery.Release()

		if err = idxSessQuery.Exec(); err != nil {
			return err
		}
	}

	return nil
}

func (s *sessionStore) deleteSession(id string) (err error) {
	deleteQuery := s.session.Query(deleteSession, id)
	defer deleteQuery.Release()

	return deleteQuery.Exec()
}

func (s *sessionStore) deleteAuthSession(authID, id string) (err error) {
	deleteQuery := s.session.Query(deleteAuthSession, authID, id)
	defer deleteQuery.Release()

	return deleteQuery.Exec()
}

func (s *sessionStore) isExists(ctx context.Context, id string) bool {
	var count int
	query := s.session.Query(checkExistense, id).WithContext(ctx)
	if err := query.Scan(&count); err != nil {
		return false
	}

	return count > 0
}
