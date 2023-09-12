package events

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"telegram_bot/db/repository"
	sessions "telegram_bot/db/session"
	"telegram_bot/exchagerate"
	"time"
)

func (p *EventProcess) doCmd(event Event) error {

	event.Text = strings.TrimSpace(event.Text)
	log.Printf("got new command %s from %s", event.Text, event.Username)

	// для теста контекст находится тут
	ctx := context.Background()

	// Сохранение сообщения в редис в виде сессии
	existSession, err := p.redis.ExistSession(ctx, convEvent(event))
	if err != nil {
		return fmt.Errorf("problem with check session:%w", err)
	}
	if !existSession {
		if err = p.redis.NewSession(ctx, convEvent(event)); err != nil {
			return fmt.Errorf("problem with save new session:%w", err)
		}
	} else {
		err = p.redis.SaveMess(convEvent(event))
		if err != nil {
			return fmt.Errorf("failed with save message: %w", err)
		}
	}
	problemExist, err := p.checkBeforeDoCmd(ctx, event)
	if err != nil {
		return fmt.Errorf("problem with check before cmd:%w", err)
	}
	if problemExist {
		return err
	}
	switch event.Text {
	case start:
		err = p.sendKeyboard(ctx, event)
		if err != nil {
			return fmt.Errorf("failed with send keyboard: %w", err)
		}
		return p.tg.SendMessage(event.ChatId, "Привет!, это бот Габидена:)")
	case exchangeRateK:
		err = p.sendKeyboard(ctx, event)
		if err != nil {
			return fmt.Errorf("failed with send keyboard: %w", err)
		}
	case getFiat, exchangerateFiatK:
		fiat := exchagerate.ParseFiat()
		resF := fmt.Sprintf("курс $ - %s \nкурс € - %s\nкурс ₽ - %s", fiat["USD"], fiat["EUR"], fiat["RUB"])
		return p.tg.SendMessage(event.ChatId, resF)
	case getCrypto, exchangerateCryptoK:
		crypto := exchagerate.ParseCrypto()
		resC := fmt.Sprintf("курс ₿ - %s \nкурс ETH - %s", crypto["Bitcoin"], crypto["Ethereum"])
		return p.tg.SendMessage(event.ChatId, resC)
	case createUser, createUserK:
		return p.createUser(ctx, event)
	case createExercise, createExercisesK:
		err = p.createExercise(ctx, event)
		if err != nil {
			return fmt.Errorf("failed with create exercises: %w", err)
		}
	case createTrainday, createTraindayK:
		err = p.createTrainDay(ctx, event)
		if err != nil {
			return fmt.Errorf("failed with create train day: %w", err)
		}
	case startTrain, startTrainK:
		if err = p.startTrain(ctx, event); err != nil {
			return fmt.Errorf("failed with start train: %w", err)
		}
	case stopTrain, stopTrainK:
		if err = p.writeLogTrain(ctx, event); err != nil {
			return fmt.Errorf("failed with write log exercise: %w", err)
		}
		if err = p.redis.StopTrain(event.Username); err != nil {
			return fmt.Errorf("cant stop train on redis: %w", err)
		}
		if err = p.stopTrain(ctx, event); err != nil {
			return fmt.Errorf("failed with stop train: %w", err)
		}
	case trainK:
		err = p.sendKeyboard(ctx, event)
		if err != nil {
			return fmt.Errorf("failed with send keyboard: %w", err)
		}
	case nextExercise:
		err = p.sendKeyboard(ctx, event)
		if err != nil {
			return fmt.Errorf("failed with send keyboard: %w", err)
		}
		err = p.redis.NextExercise(event.Username)
		if err != nil {
			return fmt.Errorf("failed with next exercise: %w", err)
		}
	case help:
		return p.tg.SendMessage(event.ChatId, "У меня 2 функции\n"+
			"1) /exchangerate_fiat - курс фиатных валют\n2)/exchangerate_crypto - курс криптовалют")
	default:
		return p.DefaultCmd(ctx, event)
	}
	return nil
}

func (p *EventProcess) DefaultCmd(ctx context.Context, ev Event) error {
	lastMessage, err := p.redis.GetLastMess(convEvent(ev))
	//если тренировка началась
	ok, err := p.redis.ExistStartTrain(convEvent(ev))
	if tr, _ := p.redis.GetTrainSession(ev.Username); ok && tr.Exercises != nil {
		lastMessage = trainingProcess
	}
	if err != nil {
		return fmt.Errorf("poblem with get last mess: %w", err)
	}
	switch lastMessage {
	case createUser, createUserK:
		err = p.createUserHandle(ctx, ev)
		if err != nil {
			return fmt.Errorf("failed with create user handle: %w", err)
		}
	case createExercise, createExercisesK:
		err = p.createExerciseHandle(ctx, ev)
		if err != nil {
			return fmt.Errorf("failed with create exercieses handle: %w", err)
		}
	case createTrainday, createTraindayK:
		err = p.createTraindayHandle(ctx, ev)
		if err != nil {
			return fmt.Errorf("failed with create train day handle: %w", err)
		}
	case startTrain, startTrainK:
		err = p.startTrainHandle(ctx, ev)
		if err != nil {
			return fmt.Errorf("failed with  start train handle: %w", err)
		}
	case trainingProcess:
		err = p.sendKeyboard(ctx, ev)
		if err != nil {
			return fmt.Errorf("failed with send keyboard: %w", err)
		}
		err = p.trainingProcessHandle(ev)
		if err != nil {
			return fmt.Errorf("failed with train process: %w", err)
		}
	default:
		err = p.tg.SendMessage(ev.ChatId, "Такой команды нет")
		if err != nil {
			return fmt.Errorf("failed with train process: %w", err)
		}

	}
	return nil
}

func (p *EventProcess) createTrainDay(ctx context.Context, ev Event) error {
	err := p.redis.SaveMess(convEvent(ev))
	if err != nil {
		return fmt.Errorf("failed with save message: %w", err)
	}
	user := repository.User{Name: ev.Username}
	text := "Напишите через запятую название тренировочного дня и числа которые указаны рядом с упражнением\n"
	exercises, err := p.storage.GetExercises(ctx, user)
	if err != nil {
		return fmt.Errorf("failed with get all exercises: %w", err)
	}
	for _, e := range exercises {
		text += fmt.Sprintf("%s - %v\n", e.Activity, e.ExerciseId)
	}
	err = p.tg.SendMessage(ev.ChatId, text)
	if err != nil {
		return fmt.Errorf("failed with send message: %w", err)
	}

	return err
}

func (p *EventProcess) createExercise(ctx context.Context, ev Event) error {
	existUser, err := p.storage.ExistUser(ctx, ev.Username)
	if err != nil {
		return fmt.Errorf("can't check user: %w", err)
	}
	if !existUser {
		err = p.tg.SendMessage(ev.ChatId, "Зарегистрируйтесь для начало")
		if err != nil {
			return fmt.Errorf("failed with send message: %w", err)
		}
		return nil
	}
	err = p.tg.SendMessage(ev.ChatId, "Напишите через запятую название упражнения и группу мышц")
	if err != nil {
		return fmt.Errorf("failed with send message: %w", err)
	}
	return nil
}

func (p *EventProcess) trainProcess(ev Event) error {
	rawText := strings.Fields(ev.Text)
	if len(rawText) != 2 {
		return p.tg.SendMessage(ev.ChatId, "отправьте только вес и кол-во повторений")

	}
	weight, err := strconv.Atoi(rawText[0])
	reps, err1 := strconv.Atoi(rawText[1])
	if err != nil && err1 != nil {
		return p.tg.SendMessage(ev.ChatId, "некорректные данные, введите только 2 числа")
	}
	err = p.redis.SetLogExercise(ev.Username, sessions.InfoEx{
		Weight: weight,
		Reps:   reps,
	})
	if err != nil {
		return fmt.Errorf("failed with set log exercise: %w", err)
	}

	return nil
}

func (p *EventProcess) createUser(ctx context.Context, ev Event) error {
	existUser, err := p.storage.ExistUser(ctx, ev.Username)
	if err != nil {
		return fmt.Errorf("can't check user: %w", err)
	}
	if existUser {
		err = p.tg.SendMessage(ev.ChatId, "Пользователь уже существует")
		if err != nil {
			return fmt.Errorf("failed with send message: %w", err)
		}
	} else {
		err = p.tg.SendMessage(ev.ChatId, "Сколько вам лет?")
		if err != nil {
			return fmt.Errorf("failed with send message: %w", err)
		}
	}
	err = p.redis.SaveMess(convEvent(ev))
	if err != nil {
		return fmt.Errorf("failed with send message: %w", err)
	}
	return nil
}

func (p *EventProcess) startTrain(ctx context.Context, ev Event) error {
	trainDays, err := p.storage.GetTrainDays(ctx, repository.User{Name: ev.Username})
	var text strings.Builder
	for _, item := range trainDays {
		text.WriteString(item.TypeName + "-" + strconv.Itoa(item.TrainingId) + "\n")
	}
	textSend := "Выберите тренировочный день, написав номер тренировочного дня:\n" + text.String()
	err = p.tg.SendMessage(ev.ChatId, textSend)
	if err != nil {
		return fmt.Errorf("can't start train : %w", err)
	}
	return nil
}

func (p *EventProcess) stopTrain(ctx context.Context, ev Event) error {
	session := repository.TrainSession{
		UserName:     ev.Username,
		Status:       false,
		TrainEndTime: time.Now(),
	}
	if err := p.storage.StopTrain(ctx, session); err != nil {
		return fmt.Errorf("can't stop train: %w", err)
	}
	textSend := "Тренировка завершилась"
	err := p.tg.SendMessage(ev.ChatId, textSend)
	if err != nil {
		return fmt.Errorf("can't stop train : %w", err)
	}
	return nil
}

func (p *EventProcess) writeLogTrain(ctx context.Context, ev Event) error {
	trainSess, err := p.redis.GetTrainSession(ev.Username)
	for _, item := range trainSess.Exercises {
		for i, info := range item.InfoEx {
			l := repository.LogExercise{
				SessionId:    trainSess.Sess_id,
				ExerciseName: item.NameExercise,
				NumberSet:    len(item.InfoEx),
				SerialNumSet: i + 1,
				Weight:       info.Weight,
				Reps:         info.Reps,
			}
			err = p.storage.WriteToLogsExercise(ctx, l)
			if err != nil {
				return fmt.Errorf("failed with write log exercises: %w", err)
			}
		}
	}
	return nil
}

func (p *EventProcess) checkBeforeDoCmd(ctx context.Context, ev Event) (bool, error) {
	switch ev.Text {
	case createUser, createUserK:
		exist, err := p.storage.ExistUser(ctx, ev.Username)
		if err != nil {
			return false, fmt.Errorf("failed with check before do cmd createUser: %w", err)
		}
		if exist == true {
			return true, p.tg.SendMessage(ev.ChatId, "Пользователь уже существует")
		}
	case createExercise, createExercisesK:
		exist, err := p.storage.ExistUser(ctx, ev.Username)
		if err != nil {
			return false, fmt.Errorf("failed with check before do cmd createExercise: %w", err)
		}
		if exist == false {
			return true, p.tg.SendMessage(ev.ChatId, "Создайте пользователя")
		}
	case createTrainday, createTraindayK:
		exs, err := p.storage.GetExercises(ctx, repository.User{Name: ev.Username})
		if err != nil {
			return false, fmt.Errorf("failed with check before do cmd createTrainday: %w", err)
		}
		if exs == nil {
			return true, p.tg.SendMessage(ev.ChatId, "У вас нет упражнений")
		}
	case nextExercise:
		trSess, err := p.redis.GetTrainSession(ev.Username)
		if err != nil {
			return false, fmt.Errorf("failed with check before do cmd nextExercise: %w", err)
		}
		if trSess.Exercises[len(trSess.Exercises)-1].Status == 1 {
			return true, p.tg.SendMessage(ev.ChatId, "Упражнений больше нет")
		}
	case startTrain, startTrainK:
		trDays, err := p.storage.GetTrainDays(ctx, repository.User{Name: ev.Username})
		if err != nil {
			return false, fmt.Errorf("failed with check before do cmd startTrain: %w", err)
		}
		if trDays == nil {
			return true, p.tg.SendMessage(ev.ChatId, "Создайте сначала тренировочный день")
		}
	case stopTrain, stopTrainK:
		existStart, err := p.redis.ExistStartTrain(sessions.Session{Username: ev.Username})
		if err != nil {
			return false, fmt.Errorf("failed with check before do cmd stopTrain: %w", err)
		}
		if existStart {
			return false, p.tg.SendMessage(ev.ChatId, "У вас нет активной тренировки")
		}
	}
	return false, nil
}
