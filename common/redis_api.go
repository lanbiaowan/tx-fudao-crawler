package common

import (
	//"errors"
	"encoding/json"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/lilien1010/tx-fudao-crawler/model"
	"io"
	//"net"
	//"os"
	//"runtime"
	//"strconv"
	//"syscall"
	"time"
)

type handleQueueFunc func(message *string) (code uint32, err error)

type RedisApi struct {
	redisPool   *redis.Pool
	redisServer string
}

func newRedisPoolWithSizeAndPasswd(redisServer string, maxPoolSize int, passwd string) *redis.Pool {
	return &redis.Pool{
		// Maximum number of idle connections in the pool.
		MaxIdle: 10,
		// Maximum number of connections allocated by the pool at a given time.
		// When zero, there is no limit on the number of connections in the pool
		MaxActive: maxPoolSize,
		// Close connections after remaining idle for this duration. If the value
		// is zero, then idle connections are not closed. Applications should set
		// the timeout to a value less than the server's timeout.
		IdleTimeout: 240 * time.Second,
		// If Wait is true and the pool is at the MaxActive limit, then Get() waits
		// for a connection to be returned to the pool before returning.
		Wait: true,
		// Dial is an application supplied function for creating and configuring a
		// connection.
		//
		// The connection returned from Dial must not be in a special state
		// (subscribed to pubsub channel, transaction started, ...).
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", redisServer)
			if err != nil {
				return nil, err
			}
			// 如果密码为空 则使用默认密码
			if passwd != "" {
				if _, err := c.Do("AUTH", passwd); err != nil {
					c.Close()
					return nil, err
				}
			}
			return c, err
		},
		// TestOnBorrow is an optional application supplied function for checking
		// the health of an idle connection before the connection is used again by
		// the application. Argument t is the time that the connection was returned
		// to the pool. If the function returns an error, then the connection is
		// closed.
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}
}

func NewRedisApi(redisServer string, maxPoolSize int, passwd string) *RedisApi {
	pool := newRedisPoolWithSizeAndPasswd(redisServer, maxPoolSize, passwd)
	return &RedisApi{
		redisPool:   pool,
		redisServer: redisServer,
	}
}

func (api *RedisApi) Lpush(key, value string) (listSize int64, err error) {
	redisConn := api.redisPool.Get()
	r, err := redisConn.Do("LPUSH", key, value)
	if err == io.EOF {
		redisConn.Close()
		redisConn = api.redisPool.Get()
		r, err = redisConn.Do("LPUSH", key, value)
	}
	listSize, err = redis.Int64(r, err)
	redisConn.Close()
	return listSize, err
}

func (api *RedisApi) Rpush(key, value string) (listSize int64, err error) {
	redisConn := api.redisPool.Get()
	r, err := redisConn.Do("RPUSH", key, value)
	if err == io.EOF {
		redisConn.Close()
		redisConn = api.redisPool.Get()
		r, err = redisConn.Do("RPUSH", key, value)
	}
	listSize, err = redis.Int64(r, err)
	redisConn.Close()
	return listSize, err
}

///blpop 操作
func (api *RedisApi) BLpop(key string, timeout uint8) (string, error) {
	// 获取一条Redis连接
	redisConn := api.redisPool.Get()
	r, err := redisConn.Do("blpop", key, timeout)
	if err == io.EOF {
		redisConn.Close()
		redisConn = api.redisPool.Get()
		r, err = redisConn.Do("blpop", key, timeout)
	}
	var value []string
	value, err = redis.Strings(r, err)
	redisConn.Close()
	if value == nil {
		return "", err
	}

	return value[len(value)-1], err
}

func (api *RedisApi) StartWorker(
	key string,
	timeout uint8,
	max_go int,
	callback handleQueueFunc) {

	for i := (0); i < max_go; i++ {
		go func() {

			for {
				info, rerr := api.BLpop(key, timeout)
				if rerr != nil {
					fmt.Println("StartLoop() blpop error", rerr, api.redisServer)
					continue
				}
				callback(&info)
			}
		}()

	}
}

func (api *RedisApi) PushTask(key string, task *model.QueueTaskEvent) (listSize int64, err error) {

	body, err := json.Marshal(task)

	if err != nil {
		return 0, err
	}

	listSize, err = api.Rpush(key, string(body))
	return listSize, err

}
