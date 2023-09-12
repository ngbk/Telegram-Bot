create table users (
  user_id serial not null unique,
  name varchar(255) not null unique,
  age smallint
);

create table exercises (
  exercise_id serial not null unique,
  muscle_group varchar(255) not null,
  activity varchar(255) not null unique,
  user_name varchar(255) references users (name)
);

create table train_day(
  training_id serial not null unique,
  name_train_day varchar unique,
  user_name varchar(255)
);

create table sessions (
  session_id serial not null unique,
  user_name varchar(255) references users (name),
  name_train_day varchar references train_day (name_train_day),
  status boolean,
  training_start_time timestamp,
  training_end_time timestamp
);

create table log_exercises (
  log_id serial not null unique,
  session_id integer references sessions (session_id),
  exercise_name varchar(255) references exercises (activity),
  number_set smallint default 4 not null,
  serial_number_set integer,
  weight integer,
  reps smallint
);


CREATE TABLE train_day_exercise (
  id serial PRIMARY KEY,
  tr_day_id integer REFERENCES train_day (training_id),
  exercise_id integer REFERENCES exercises (exercise_id)
);



