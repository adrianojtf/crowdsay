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
	mu            sync.Mutex
	rdb           *redis.Client
	inMemoryStore *memoryStore
)

type memoryStore struct {
	polls map[int64]*Poll
	votes map[string]bool
	idSeq int64
}

func init() {
	_ = godotenv.Load()
	if os.Getenv("ENV") == "test" {
		inMemoryStore = &memoryStore{
			polls: make(map[int64]*Poll),
			votes: make(map[string]bool),
			idSeq: 0,
		}
	} else {
		rdb = newRedisClient()
	}
}

func newRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
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
	initPollFields(p)
	if inMemoryStore != nil {
		inMemoryStore.idSeq++
		p.ID = inMemoryStore.idSeq
		inMemoryStore.polls[p.ID] = p
		return nil
	}
	p.ID = generateID()
	return savePoll(p)
}

func initPollFields(p *Poll) {
	if p.Votes == nil {
		p.Votes = make(map[string]int)
	}
	if p.Results == nil {
		p.Results = make(map[string]int)
	}
}

func savePoll(p *Poll) error {
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return rdb.Set(context.Background(), pollKey(p.ID), data, 0).Err()
}

func GetPoll(id int64) (*Poll, error) {
	if inMemoryStore != nil {
		p, ok := inMemoryStore.polls[id]
		if !ok {
			return nil, errors.New("poll not found")
		}
		return p, nil
	}
	return loadPoll(id)
}

func loadPoll(id int64) (*Poll, error) {
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
	if inMemoryStore != nil {
		voteKey := voteRedisKey(poll.ID, ip)
		return inMemoryStore.votes[voteKey]
	}
	return hasVotedRedis(poll.ID, ip)
}

func hasVotedRedis(pollID int64, ip string) bool {
	voteKey := voteRedisKey(pollID, ip)
	res, _ := rdb.Get(context.Background(), voteKey).Result()
	return res == "1"
}

func VotePoll(pollID int64, option, ip string) error {
	mu.Lock()
	defer mu.Unlock()
	if inMemoryStore != nil {
		poll, ok := inMemoryStore.polls[pollID]
		if !ok {
			return errors.New("poll not found")
		}
		if poll.Results == nil {
			poll.Results = make(map[string]int)
		}
		voteKey := voteRedisKey(pollID, ip)
		if inMemoryStore.votes[voteKey] {
			return errors.New("already voted from this IP")
		}
		if !isValidOption(poll.Options, option) {
			return errors.New("invalid option")
		}
		inMemoryStore.votes[voteKey] = true
		poll.Results[option]++
		return nil
	}
	poll, err := loadPoll(pollID)
	if err != nil {
		return err
	}
	if poll.Results == nil {
		poll.Results = make(map[string]int)
	}
	if hasVotedRedis(pollID, ip) {
		return errors.New("already voted from this IP")
	}
	if !isValidOption(poll.Options, option) {
		return errors.New("invalid option")
	}
	if err := saveVoteRedis(pollID, ip); err != nil {
		return err
	}
	poll.Results[option]++
	return savePoll(poll)
}

func isValidOption(options []string, option string) bool {
	for _, opt := range options {
		if opt == option {
			return true
		}
	}
	return false
}

func saveVoteRedis(pollID int64, ip string) error {
	voteKey := voteRedisKey(pollID, ip)
	return rdb.Set(context.Background(), voteKey, "1", 0).Err()
}

func GetAllPolls() ([]*Poll, error) {
	if inMemoryStore != nil {
		polls := make([]*Poll, 0, len(inMemoryStore.polls))
		for _, p := range inMemoryStore.polls {
			polls = append(polls, p)
		}
		return polls, nil
	}
	return loadAllPolls()
}

func loadAllPolls() ([]*Poll, error) {
	keys, err := rdb.Keys(context.Background(), "poll:*").Result()
	if err != nil {
		return nil, err
	}
	allPolls := make([]*Poll, 0, len(keys))
	for _, k := range keys {
		p, err := loadPollByKey(k)
		if err == nil {
			allPolls = append(allPolls, p)
		}
	}
	return allPolls, nil
}

func loadPollByKey(key string) (*Poll, error) {
	data, err := rdb.Get(context.Background(), key).Result()
	if err != nil {
		return nil, err
	}
	var p Poll
	if err := json.Unmarshal([]byte(data), &p); err != nil {
		return nil, err
	}
	return &p, nil
}

func generateID() int64 {
	if inMemoryStore != nil {
		inMemoryStore.idSeq++
		return inMemoryStore.idSeq
	}
	id, _ := rdb.Incr(context.Background(), "poll:id:seq").Result()
	return id
}

func pollKey(id int64) string {
	return fmt.Sprintf("poll:%d", id)
}

func voteRedisKey(pollID int64, ip string) string {
	return fmt.Sprintf("vote:%d:%s", pollID, ip)
}
