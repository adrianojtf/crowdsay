package poll

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
)

var (
	mu  sync.Mutex
	rdb *redis.Client
)

func init() {
	_ = godotenv.Load()
	rdb = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       getEnvInt("REDIS_DB", 0),
	})
}

func getEnvInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		var i int
		_, err := fmt.Sscanf(value, "%d", &i)
		if err == nil {
			return i
		}
	}
	return fallback
}

func CreatePoll(p *Poll) error {
	mu.Lock()
	defer mu.Unlock()
	p.Votes = make(map[string]int)
	p.Results = make(map[string]int)
	p.ID = generateID()
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}
	err = rdb.Set(context.Background(), pollKey(p.ID), data, 0).Err()
	if err != nil {
		return err
	}
	return nil
}

func GetPoll(id int64) (*Poll, error) {
	data, err := rdb.Get(context.Background(), pollKey(id)).Result()
	if err == redis.Nil {
		return nil, errors.New("poll not found")
	} else if err != nil {
		return nil, err
	}
	var poll Poll
	if err := json.Unmarshal([]byte(data), &poll); err != nil {
		return nil, err
	}
	return &poll, nil
}

func HasVotedIP(poll *Poll, ip string) bool {
	voteKey := voteRedisKey(poll.ID, ip)
	res, _ := rdb.Get(context.Background(), voteKey).Result()
	return res == "1"
}

func VotePoll(pollID int64, option, ip string) error {
	mu.Lock()
	defer mu.Unlock()
	poll, err := GetPoll(pollID)
	if err != nil {
		return err
	}
	if poll.Results == nil {
		poll.Results = make(map[string]int)
	}
	if HasVotedIP(poll, ip) {
		return errors.New("already voted from this IP")
	}
	found := false
	for _, opt := range poll.Options {
		if opt == option {
			found = true
			break
		}
	}
	if !found {
		return errors.New("invalid option")
	}
	// Salva voto no Redis
	voteKey := voteRedisKey(pollID, ip)
	rdb.Set(context.Background(), voteKey, "1", 0)
	poll.Results[option]++
	// Atualiza poll no Redis
	data, err := json.Marshal(poll)
	if err != nil {
		return err
	}
	err = rdb.Set(context.Background(), pollKey(pollID), data, 0).Err()
	if err != nil {
		return err
	}
	return nil
}

func GetAllPolls() ([]*Poll, error) {
	keys, err := rdb.Keys(context.Background(), "poll:*").Result()
	if err != nil {
		return nil, err
	}
	allPolls := make([]*Poll, 0, len(keys))
	for _, k := range keys {
		data, err := rdb.Get(context.Background(), k).Result()
		if err != nil {
			continue
		}
		var p Poll
		if err := json.Unmarshal([]byte(data), &p); err == nil {
			allPolls = append(allPolls, &p)
		}
	}
	return allPolls, nil
}

func generateID() int64 {
	// Usa o Redis INCR para garantir unicidade
	id, _ := rdb.Incr(context.Background(), "poll:id:seq").Result()
	return id
}

func pollKey(id int64) string {
	return fmt.Sprintf("poll:%d", id)
}

func voteRedisKey(pollID int64, ip string) string {
	return fmt.Sprintf("vote:%d:%s", pollID, ip)
}
