package sessions

import (
	"github.com/go-redis/redis/v8"
	"github.com/nitishm/go-rejson/v4"
)

type RedisDB struct {
	client      *redis.Client
	HandlerJSON *rejson.Handler
}

type ConfigRedis struct {
	Addr     string
	Password string
	Db       int
}

type Session struct {
	Username string
	ChatId   int
	Message  []string
	//TrainMode bool
	Train TrainSess
}

type TrainSess struct {
	Sess_id    int
	Train_name string
	Exercises  []Exercises
}

// stutus : 0 - упр не начато, 1 - идет упражнение, 2 - упр завершено

type Exercises struct {
	NameExercise string
	Status       int
	InfoEx       []InfoEx
}
type InfoEx struct {
	Weight int
	Reps   int
}
