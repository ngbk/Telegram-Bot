package repository

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
)

type ConfigDB struct {
	Host     string
	Port     string
	Username string
	Password string
	DBName   string
	SSLMode  string
}

type Storage interface {
	CreateUser(context.Context, User) error
	CreateExercise(context.Context, Exercise) error
	CreateTrainDay(context.Context, TrainDay) (int, error)
	CreateRelationTrEx(context.Context, Train_ex_day) error
	StartTrain(context.Context, TrainSession) (int, error)
	StopTrain(context.Context, TrainSession) error
	WriteToLogsExercise(context.Context, LogExercise) error
	ExistUser(context.Context, string) (bool, error)
	GetExercises(context.Context, User) ([]Exercise, error)
	GetTrainDays(context.Context, User) ([]TrainDay, error)
	GetTrainDay(context.Context, User, int) (TrainDay, error)
	GetExercisesFromTrain(context.Context, string) ([]string, error)

	//ShowRecords(context.Context)
}

type Postgres struct {
	db *sqlx.DB
}

func New(cfg ConfigDB) (*Postgres, error) {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode)
	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("can't open database:%w", err)
	}
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("can't connect database: %w", err)
	}
	return &Postgres{db: db}, nil
}

func (p *Postgres) CreateUser(ctx context.Context, user User) error {
	query := "INSERT INTO Users (Name, Age) VALUES($1, $2)"
	_, err := p.db.ExecContext(ctx, query, user.Name, user.Age)
	if err != nil {
		return fmt.Errorf("can't create use sq: %w", err)
	}

	return nil
}

func (p *Postgres) CreateExercise(ctx context.Context, ex Exercise) error {
	query := `INSERT INTO Exercises (User_Name, Muscle_Group, Activity) VALUES ($1, $2, $3) `
	_, err := p.db.ExecContext(ctx, query, ex.User_name, ex.MuscleGroup, ex.Activity)
	if err != nil {
		return fmt.Errorf("can't create exercise: %w", err)
	}
	return nil
}

func (p *Postgres) CreateTrainDay(ctx context.Context, day TrainDay) (int, error) {
	var id int
	query := `INSERT INTO Train_Day (Name_Train_Day, User_Name) VALUES ($1, $2) RETURNING Training_Id`
	err := p.db.GetContext(ctx, &id, query, day.TypeName, day.User_name)
	if err != nil {
		return 0, fmt.Errorf("can't create training day: %w", err)
	}
	return id, nil
}

func (p *Postgres) CreateRelationTrEx(ctx context.Context, relTrEx Train_ex_day) error {
	query := `INSERT INTO Train_Day_Exercise (Tr_Day_Id, Exercise_Id) VALUES ($1, $2) `
	_, err := p.db.ExecContext(ctx, query, relTrEx.Tr_id, relTrEx.Ex_id)
	if err != nil {
		return fmt.Errorf("can't create relation training day with exercises: %w", err)
	}
	return nil
}

//func (p *Postgres) getRelationTrEx(ctx context.Context, relTrEx Train_ex_day) error {
//
//}

func (p *Postgres) GetExercises(ctx context.Context, user User) ([]Exercise, error) {
	var exercises []Exercise
	query := `SELECT * FROM Exercises WHERE User_Name = $1`
	err := p.db.SelectContext(ctx, &exercises, query, user.Name)
	if err != nil {
		return nil, fmt.Errorf("can't get exercises: %w", err)
	}
	return exercises, nil
}

func (p *Postgres) GetExercisesFromTrain(ctx context.Context, trainName string) ([]string, error) {
	var activity []string
	query := `SELECT E.Activity
			FROM Train_Day Td
			JOIN Train_Day_Exercise Ted ON Td.Training_Id = Ted.Tr_Day_Id
			JOIN Exercises E ON Ted.Exercise_Id = E.Exercise_Id
			WHERE Td.Name_Train_Day = $1`
	err := p.db.SelectContext(ctx, &activity, query, trainName)
	if err != nil {
		return nil, fmt.Errorf("can't get exercises: %w", err)
	}
	return activity, nil
}

func (p *Postgres) GetTrainDays(ctx context.Context, user User) ([]TrainDay, error) {
	var trainDays []TrainDay
	query := `SELECT * FROM Train_Day  WHERE User_Name = $1`
	err := p.db.SelectContext(ctx, &trainDays, query, user.Name)
	if err != nil {
		return trainDays, fmt.Errorf("can't get train day: %w", err)
	}
	return trainDays, nil
}

func (p *Postgres) GetTrainDay(ctx context.Context, user User, id int) (TrainDay, error) {
	var trainDay TrainDay
	query := `SELECT * FROM Train_Day  WHERE User_Name = $1 AND Training_Id = $2`
	err := p.db.GetContext(ctx, &trainDay, query, user.Name, id)
	if err != nil {
		return trainDay, fmt.Errorf("can't get train day: %w", err)
	}
	return trainDay, nil
}

func (p *Postgres) StartTrain(ctx context.Context, session TrainSession) (int, error) {
	var id int
	query := `INSERT INTO Sessions (User_Name, Name_Train_Day, Status, Training_Start_Time) VALUES ($1, $2, $3, $4) RETURNING Session_Id`
	err := p.db.GetContext(ctx, &id, query, session.UserName, session.Name_train_day, session.Status, session.TrainStartTime)
	if err != nil {
		return id, fmt.Errorf("can't start train: %w", err)
	}
	return id, nil
}

func (p *Postgres) StopTrain(ctx context.Context, session TrainSession) error {
	var id int
	query := `SELECT Session_Id FROM Sessions WHERE User_Name = $1 AND Status = TRUE`
	err := p.db.GetContext(ctx, &id, query, session.UserName)
	query = `UPDATE Sessions SET Training_End_Time = $1, Status = $2 WHERE User_Name = $3 AND Session_Id = $4`
	_, err = p.db.ExecContext(ctx, query, session.TrainEndTime, session.Status, session.UserName, id)
	if err != nil {
		return fmt.Errorf("can't stop train: %w", err)
	}
	return nil
}

func (p *Postgres) WriteToLogsExercise(ctx context.Context, logExercise LogExercise) error {
	query := `INSERT INTO Log_Exercises (Session_Id, Exercise_Name, Weight, Reps, Serial_Number_Set, Number_Set) 
				VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := p.db.ExecContext(ctx, query, logExercise.SessionId, logExercise.ExerciseName,
		logExercise.Weight, logExercise.Reps, logExercise.SerialNumSet, logExercise.NumberSet)
	if err != nil {
		return fmt.Errorf("can't save logs exercise: %w", err)
	}
	return nil
}

func (p *Postgres) ExistUser(ctx context.Context, name string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM Users WHERE Name = $1`
	err := p.db.QueryRowContext(ctx, query, name).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("it was occured error with scan row: %w", err)
	}
	return count > 0, nil
}
