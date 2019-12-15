package bloom

import (
	"github.com/gomodule/redigo/redis"
)

// RedisStorage is a struct representing the Redis backend for the bloom filter.
type RedisStorage struct {
	pool  *redis.Pool
	key   string
	size  uint
	queue []uint
}

// NewRedisStorage creates a Redis backend storage to be used with the bloom filter.
func NewRedisStorage(pool *redis.Pool, key string, size uint) (*RedisStorage, error) {
	var err error

	store := RedisStorage{pool, key, size, make([]uint, 0)}

	conn := store.pool.Get()
	defer conn.Close()
	exists, err := redis.Bool(conn.Do("EXISTS", key))
	if err != nil {
		return &store, err
	}

	if !exists {
		if err := store.init(); err != nil {
			return &store, err
		}
	}

	return &store, nil
}

// init takes care of settings every bit to 0 in the Redis bitset.
func (s *RedisStorage) init() (err error) {
	conn := s.pool.Get()
	defer conn.Close()

	var i uint
	for i = 0; i < s.size; i++ {
		conn.Send("SETBIT", s.key, i, 0)
	}

	err = conn.Flush()

	return
}

// Append appends the bit, which is to be saved, to the queue.
func (s *RedisStorage) Append(bit uint) {
	s.queue = append(s.queue, bit)
}

// Save pushes the bits from the queue to the storage backend, assigning the value 1 in the process.
func (s *RedisStorage) Save() {
	conn := s.pool.Get()
	defer conn.Close()

	for _, bit := range s.queue {
		conn.Send("SETBIT", s.key, bit, 1)
	}

	conn.Flush()
}

// Exists checks if the given bit exists in the Redis backend.
func (s *RedisStorage) Exists(bit uint) (ret bool, err error) {
	conn := s.pool.Get()
	defer conn.Close()

	bitValue, err := redis.Int(conn.Do("GETBIT", s.key, bit))
	if err != nil {
		return
	}
	return bitValue == 1, err
}

// Add add the bits, which are to be saved, to the queue.
func (s *RedisStorage) Add(bits ...uint) {
	s.queue = append(s.queue, bits...)
}

// Exist check if the given bits exists in the Redis backend.
func (s *RedisStorage) Exist(bits ...uint) (ret []bool, err error) {
	conn := s.pool.Get()
	defer conn.Close()

	ret = make([]bool, len(bits))
	for index, bit := range bits {
		bitValue, err := redis.Int(conn.Do("GETBIT", s.key, bit))
		if err != nil {
			return ret, err
		}
		ret[index] = bitValue == 1
	}

	return
}

// Load load all init data and save into the Redis backend.
func (s *RedisStorage) Load(bits ...uint) {
	s.Add(bits...)
	s.Save()
}
