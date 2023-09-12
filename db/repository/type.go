package repository

import "time"

type User struct {
	Id   int    `json:"user_id" db:"user_id"`
	Name string `json:"user_name" db:"user_name"`
	Age  int    `json:"age" db:"age"`
}

type Exercise struct {
	ExerciseId  int    `json:"exercise_id" db:"exercise_id"`
	User_name   string `json:"user_name" db:"user_name"`
	MuscleGroup string `json:"muscle_group" db:"muscle_group"`
	Activity    string `json:"activity" db:"activity"`
}

type TrainDay struct {
	TrainingId int    `json:"training_id" db:"training_id"`
	TypeName   string `json:"name_train_day" db:"name_train_day"`
	User_name  string `json:"user_name" db:"user_name"`
}

type TrainSession struct {
	SessionId      int       `json:"session_id" db:"session_id"`
	UserName       string    `json:"user_name" db:"user_name"`
	Name_train_day string    `json:"name_train_day" db:"name_train_day"`
	Status         bool      `json:"status" db:"status"`
	TrainStartTime time.Time `json:"training_start_time" db:"training_start_time"`
	TrainEndTime   time.Time `json:"training_end_time" db:"training_end_time"`
}

type LogExercise struct {
	LogId        int    `json:"log_id" db:"log_id"`
	SessionId    int    `json:"session_id" db:"session_id"`
	ExerciseName string `json:"exercise_name" db:"exercise_name"`
	NumberSet    int    `json:"number_set" db:"number_set"`
	SerialNumSet int    `json:"serial_number_set" db:"serial_number_set"`
	Weight       int    `json:"weight" db:"weight"`
	Reps         int    `json:"reps" db:"reps"`
}

type Train_ex_day struct {
	Id    int `json:"id" db:"id"`
	Tr_id int `json:"tr_day_id" db:"tr_day_id"`
	Ex_id int `json:"ex_exercise_id" db:"exercise_id"`
}
