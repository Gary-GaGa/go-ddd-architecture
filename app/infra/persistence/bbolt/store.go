package bbolt

import (
	"encoding/json"
	"errors"
	"time"

	bolt "go.etcd.io/bbolt"

	"go-ddd-architecture/app/domain/gametime"
	"go-ddd-architecture/app/domain/player"
	outPort "go-ddd-architecture/app/usecase/port/out/game"
)

const (
	bucketPlayer     = "player"
	bucketTimestamps = "timestamps"
	keyPlayer        = "player"
	keyTimestamps    = "timestamps"
)

// Store 實作 Game 用例的 Repository，使用 bbolt 做本地單檔儲存。
type Store struct {
	db *bolt.DB
}

// New 開啟或建立資料庫檔案。
func New(path string) (*Store, error) {
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}
	s := &Store{db: db}
	// 建立 buckets
	if err := s.db.Update(func(tx *bolt.Tx) error {
		if _, e := tx.CreateBucketIfNotExists([]byte(bucketPlayer)); e != nil {
			return e
		}
		if _, e := tx.CreateBucketIfNotExists([]byte(bucketTimestamps)); e != nil {
			return e
		}
		return nil
	}); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

// Close 釋放底層資源。
func (s *Store) Close() error { return s.db.Close() }

func (s *Store) Load() (player.Player, gametime.Timestamps, error) {
	var p player.Player
	var ts gametime.Timestamps
	err := s.db.View(func(tx *bolt.Tx) error {
		// player
		bp := tx.Bucket([]byte(bucketPlayer))
		if bp == nil {
			return errors.New("player bucket not found")
		}
		if v := bp.Get([]byte(keyPlayer)); v != nil {
			if e := json.Unmarshal(v, &p); e != nil {
				return e
			}
		}
		// timestamps
		bt := tx.Bucket([]byte(bucketTimestamps))
		if bt == nil {
			return errors.New("timestamps bucket not found")
		}
		if v := bt.Get([]byte(keyTimestamps)); v != nil {
			if e := json.Unmarshal(v, &ts); e != nil {
				return e
			}
		}
		return nil
	})
	if err != nil {
		return p, ts, err
	}
	// 首次啟動：若未存過 timestamps，避免超大 Δt 或負值，將關閉時間設為現在。
	if ts.WallClockAtClose.IsZero() {
		ts.WallClockAtClose = time.Now()
	}
	return p, ts, nil
}

func (s *Store) Save(p player.Player, ts gametime.Timestamps) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		bp := tx.Bucket([]byte(bucketPlayer))
		bt := tx.Bucket([]byte(bucketTimestamps))
		if bp == nil || bt == nil {
			return errors.New("buckets not initialized")
		}
		pb, e := json.Marshal(p)
		if e != nil {
			return e
		}
		if e = bp.Put([]byte(keyPlayer), pb); e != nil {
			return e
		}
		tb, e := json.Marshal(ts)
		if e != nil {
			return e
		}
		if e = bt.Put([]byte(keyTimestamps), tb); e != nil {
			return e
		}
		return nil
	})
}

var _ outPort.Repository = (*Store)(nil)
