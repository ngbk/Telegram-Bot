package events

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"telegram_bot/db/repository"
	sessions "telegram_bot/db/session"
	"time"
)

func (p *EventProcess) createUserHandle(ctx context.Context, ev Event) error {
	var age int
	age, err := strconv.Atoi(ev.Text)
	if err != nil {
		return p.tg.SendMessage(ev.ChatId, "Вы ввели не числовое значение, повторите команду")

	}
	user := repository.User{Name: ev.Username,
		Age: age,
	}
	if err = p.storage.CreateUser(ctx, user); err != nil {
		return fmt.Errorf("problem with create user: %w", err)
	}
	return p.tg.SendMessage(ev.ChatId, "Пользователь создан")
}

func (p *EventProcess) createExerciseHandle(ctx context.Context, ev Event) error {
	rawExercise := strings.Split(ev.Text, ",")
	if len(rawExercise) != 2 {
		return p.tg.SendMessage(ev.ChatId, "Вы ввели данные не корректно, введите через запятую")
	}
	exercise := repository.Exercise{
		User_name:   ev.Username,
		Activity:    strings.TrimSpace(rawExercise[0]),
		MuscleGroup: strings.TrimSpace(rawExercise[1]),
	}
	err := p.storage.CreateExercise(ctx, exercise)
	if err != nil {
		return fmt.Errorf("failed with create exercise: %w", err)
	}
	return p.tg.SendMessage(ev.ChatId, "Упражнение успешно создано")

}

func (p *EventProcess) createTraindayHandle(ctx context.Context, ev Event) error {
	var tr_id, ex_id int
	rawTday := strings.Split(ev.Text, ",")
	tDay := repository.TrainDay{
		TypeName:  strings.TrimSpace(rawTday[0]),
		User_name: ev.Username,
	}
	tr_id, err := p.storage.CreateTrainDay(ctx, tDay)
	if err != nil {
		return fmt.Errorf("failed with create train day: %w", err)
	}
	for _, item := range rawTday[1:] {
		ex_id, err = strconv.Atoi(strings.TrimSpace(item))
		if err != nil {
			return p.tg.SendMessage(ev.ChatId, "Повторите команду, данные были введены некорректно")
		}
		rel := repository.Train_ex_day{
			Tr_id: tr_id,
			Ex_id: ex_id,
		}
		err = p.storage.CreateRelationTrEx(ctx, rel)
		if err != nil {
			return fmt.Errorf("failed with create relation ex and train day:%w", err)
		}
	}
	return p.tg.SendMessage(ev.ChatId, "Тренировочный день создан успешно")
}

func (p *EventProcess) startTrainHandle(ctx context.Context, ev Event) error {
	var tr_id, sess_id int
	tr := repository.TrainDay{}

	tr_id, err := strconv.Atoi(ev.Text)
	if err != nil {
		return p.tg.SendMessage(ev.ChatId, "Введите правильно номер тренировочного дня")
	}
	tr, err = p.storage.GetTrainDay(ctx, repository.User{Name: ev.Username}, tr_id)
	session := repository.TrainSession{
		UserName:       ev.Username,
		Name_train_day: tr.TypeName,
		Status:         true,
		TrainStartTime: time.Now(),
	}
	sess_id, err = p.storage.StartTrain(ctx, session)
	if err != nil {
		return fmt.Errorf("can't start train: %w", err)
	}
	var activities []string

	activities, err = p.storage.GetExercisesFromTrain(ctx, tr.TypeName)
	if err != nil {
		return fmt.Errorf("failed with get exercise from train_day: %w", err)
	}
	exs := make([]sessions.Exercises, len(activities))
	for i := range activities {
		exs[i].NameExercise = activities[i]
		exs[i].InfoEx = []sessions.InfoEx{}
	}
	s := sessions.TrainSess{
		Sess_id:    sess_id,
		Train_name: tr.TypeName,
		Exercises:  exs,
	}
	err = p.redis.SaveTrainSession(s, ev.Username)
	if err != nil {
		return fmt.Errorf("failed with save train session: %w", err)
	}
	text := "Тренировка началась\n\n"
	for _, item := range activities {
		text += fmt.Sprintf("%s\n", item)
	}
	text += "\n" + t
	return p.tg.SendMessage(ev.ChatId, text)
}

func (p *EventProcess) trainingProcessHandle(ev Event) error {
	err := p.redis.NextWorkoutSet(ev.Username)
	if err != nil {
		return fmt.Errorf("failed with save next workout set: %w", err)
	}
	return p.trainProcess(ev)
}
