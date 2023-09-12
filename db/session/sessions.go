package sessions

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/nitishm/go-rejson/v4"
	"time"
)

func NewRedisClient(ctx context.Context, cfg ConfigRedis) (*RedisDB, error) {
	redisDB := RedisDB{
		client: redis.NewClient(&redis.Options{
			Addr:     cfg.Addr,
			Password: cfg.Password,
			DB:       cfg.Db,
		}),
		HandlerJSON: rejson.NewReJSONHandler(),
	}
	err := redisDB.client.Ping(ctx).Err()
	redisDB.HandlerJSON.SetGoRedisClient(redisDB.client)
	return &redisDB, err
}

func (r *RedisDB) NewSession(ctx context.Context, sessionUser Session) error {
	_, err := r.HandlerJSON.JSONSet(sessionUser.Username, ".", sessionUser)
	t := time.Minute * 180
	r.client.Expire(ctx, sessionUser.Username, t)
	if err != nil {
		return fmt.Errorf("cant set session:%w", err)
	}
	return nil
}

func (r *RedisDB) ExistSession(ctx context.Context, sessionUser Session) (bool, error) {
	res, err := r.client.Exists(ctx, sessionUser.Username, ".").Result()
	if err != nil {
		return false, fmt.Errorf("cant check exits :%w", err)
	}
	if res == 1 {
		return true, err
	} else {
		return false, err
	}
}

func (r *RedisDB) SaveMess(sessionUser Session) error {
	_, err := r.HandlerJSON.JSONArrAppend(sessionUser.Username, ".Message", sessionUser.Message[0])
	if err != nil {
		return fmt.Errorf("failed to json append: %w", err)
	}
	return err
}

func (r *RedisDB) GetLastMess(sessionUser Session) (string, error) {
	var mess string
	res, err := r.HandlerJSON.JSONArrLen(sessionUser.Username, "Message")
	if err != nil {
		return "", fmt.Errorf("failed to get len array: %w", err)
	}
	lenArr, _ := res.(int)
	lenArr -= 2
	query := fmt.Sprintf("Message[%d]", lenArr)
	res, err = r.HandlerJSON.JSONGet(sessionUser.Username, query)
	if res == nil {
		return "", err
	} else {
		bytesRes := res.([]byte)
		_ = json.Unmarshal(bytesRes, &mess)
		return mess, nil
	}
}

func (r *RedisDB) ExistStartTrain(sessionUser Session) (bool, error) {
	var (
		sess Session
		flag bool
	)
	res, err := r.HandlerJSON.JSONGet(sessionUser.Username, ".")
	bytes, ok := res.([]byte)
	if !ok {
		return false, fmt.Errorf("failed with get session:%w", err)
	}
	err = json.Unmarshal(bytes, &sess)
	if err != nil {
		return false, fmt.Errorf("failed with unmarshal session:%w", err)
	}
	// можно позже оптимизировать
	for _, item := range sess.Message {
		if item == "Начать тренировку" || item == "/startTrain" {
			flag = true
		}
		if item == "Завершить тренировку" || item == "/stopTrain" {
			flag = false
		}
	}
	return flag, nil
}

func (r *RedisDB) SaveTrainSession(TrainSess TrainSess, Username string) error {
	_, err := r.HandlerJSON.JSONSet(Username, ".Train", TrainSess)
	if err != nil {
		return fmt.Errorf("failed to json append train session: %w", err)
	}
	return err
}

func (r *RedisDB) GetTrainSession(Username string) (TrainSess, error) {
	res, err := r.HandlerJSON.JSONGet(Username, ".Train")
	trSession := TrainSess{}
	if res == nil {
		return trSession, err
	} else {
		bytesRes := res.([]byte)
		_ = json.Unmarshal(bytesRes, &trSession)
		return trSession, nil
	}
}

// stutus : 0 - упр не начато, 1 - идет упражнение, 2 - упр завершено
func (r *RedisDB) SetLogExercise(Username string, exerInfo InfoEx) error {
	var setNum, exNum int
	trSession, err := r.GetTrainSession(Username)
	if err != nil {
		return fmt.Errorf("failed with get train session: %w ", err)
	}

	exNum, setNum = getExNumberAndSetNumber(trSession.Exercises)

	path := fmt.Sprintf(".Train.Exercises[%d].InfoEx[%d]", exNum, setNum)
	_, err = r.HandlerJSON.JSONSet(Username, path, exerInfo)
	if err != nil {
		return fmt.Errorf("failed with set exercise: %w ", err)
	}
	return err
}

func (r *RedisDB) NextExercise(Username string) error {
	var exNum int
	trSession, err := r.GetTrainSession(Username)
	if err != nil {
		return fmt.Errorf("failed with get train session: %w ", err)
	}
	for i, item := range trSession.Exercises {
		if item.Status == 0 {
			exNum = i
			err = r.NextWorkoutSet(Username)
			if err != nil {
				return fmt.Errorf("failed with set next set: %w ", err)
			}
			path := fmt.Sprintf(".Train.Exercises[%d].Status", exNum)
			_, err = r.HandlerJSON.JSONSet(Username, path, 1)
			if err != nil {
				return fmt.Errorf("failed with set next exercise: %w ", err)
			}
			if exNum != 0 {
				path = fmt.Sprintf(".Train.Exercises[%d].Status", exNum-1)
				_, err = r.HandlerJSON.JSONSet(Username, path, 2)
				if err != nil {
					return fmt.Errorf("failed with set next exercise: %w ", err)
				}
			}
		}
	}
	return nil
}

func (r *RedisDB) NextWorkoutSet(Username string) error {
	iExer := InfoEx{}
	trSession, err := r.GetTrainSession(Username)
	if err != nil {
		return fmt.Errorf("failed with get train session: %w ", err)
	}
	exNum, _ := getExNumberAndSetNumber(trSession.Exercises)
	path := fmt.Sprintf(".Train.Exercises[%d].InfoEx", exNum)
	_, err = r.HandlerJSON.JSONArrAppend(Username, path, iExer)
	if err != nil {
		return fmt.Errorf("failed with json arr append: %w ", err)
	}
	return nil
}

func (r *RedisDB) StopTrain(Username string) error {
	path := ".Train"
	trSession := TrainSess{}
	_, err := r.HandlerJSON.JSONSet(Username, path, trSession)
	if err != nil {
		return fmt.Errorf("failed with stop train(set): %w ", err)
	}
	return nil
}

func getExNumberAndSetNumber(exercises []Exercises) (int, int) {
	var setNum, exNum int
	for i, item := range exercises {
		if item.Status == 2 {
			continue
		} else {
			setNum, exNum = len(item.InfoEx)-1, i
			break
		}
	}
	return exNum, setNum
}
