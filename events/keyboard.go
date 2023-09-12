package events

import (
	"context"
	"fmt"
)

func (p *EventProcess) sendKeyboard(ctx context.Context, ev Event) error {
	switch ev.Text {
	case start:
		board := [][]string{{exchangeRateK, trainK}, {createUserK}}
		if p.tg.SendKeyboard(ev.ChatId, board) != nil {
			return fmt.Errorf("failed with send keyboard")
		}
	case exchangeRateK:
		board := [][]string{{exchangerateFiatK, exchangerateCryptoK}}
		if err := p.tg.SendKeyboard(ev.ChatId, board); err != nil {
			return fmt.Errorf("failed with send keyboard: %w", err)
		}
	case trainK:
		board := [][]string{{createExercisesK, createTraindayK}, {startTrainK, stopTrainK}}
		if err := p.tg.SendKeyboard(ev.ChatId, board); err != nil {
			return fmt.Errorf("failed with send keyboard: %w", err)
		}
	case createTraindayK:
		//user := repository.User{Name: ev.Username}
		//exercises, err := p.storage.GetExercises(ctx, user)
		//if err != nil {
		//	return fmt.Errorf("failed with get all exercises: %w", err)
		//}
		//text := fmt.Sprintf("Выберите упражнения для тренировки, напишите ",)
		//p.tg.SendMessage(ev.ChatId, text)
		// код для отправки упражнении в виде встроенной клавиатуры
		//
		//board := make([][]string, 0, len(exercises)/2+1)
		//row := make([]string, 0, 2)
		//for ind, ex := range exercises {
		//	rawString := ex.Activity + "-" + strconv.Itoa(ex.ExerciseId)
		//	row = append(row, rawString)
		//	if ind%2 != 0 || len(exercises)-1 == ind {
		//		board = append(board, row)
		//		row = make([]string, 0, 2)
		//	}
		//}
		//if p.tg.SendKeyboard(ev.ChatId, board) != nil {
		//	return fmt.Errorf("failed with send keyboard")
		//}
	case trainingProcess, startTrain, startTrainK, nextExercise:
		board := [][]string{{nextExercise, stopTrainK}}
		if err := p.tg.SendKeyboard(ev.ChatId, board); err != nil {
			return fmt.Errorf("failed with send keyboard: %w", err)
		}
	}
	return nil
}
