package main

import (
	"labix.org/v2/mgo"
	"time"
	"fmt"
)

type Task struct {
	Id      int "_id"
	Start   int
	Finish  int
	Done    bool
	Timestamp	time.Time
}


func (t *Task) finish() {
	t.Done = true

}

// TODO: Add a check that rate is valid 
func (j *Job) createTask(session *mgo.Session, rate int) *Task {
	var t Task

	SECONDS := 240

	total := 0
	task_max := 0

	if len(j.Tasks) > 0 {
		for _, task := range j.Tasks {
			if time.Since(task.Timestamp).Minutes() > 15 && task.Done == false {
				// This is an old task, we need to resend it for reprocessing.
				fmt.Println("Found an old task.")
				return &task
			}
			if task.Id >= task_max {
				task_max = task.Id + 1
			}
		}
	} else {
		task_max = 1
	}

	if len(j.Tasks) < task_max {
		jump := j.Start + (SECONDS * rate)

		if jump > j.Finish {
			total = j.Finish
		} else {
			total = jump
		}

		if j.Start < j.Finish {
			t = Task{
				Id:      task_max,
				Start:   j.Start,
				Finish:  total,
				Done:    false,
				Timestamp: time.Now(),
			}

			// fmt.Println(t)

			j.Tasks = append(j.Tasks, t)
			j.Start = total
		} else {
			t = Task{}
		}

		j.update(session)
	} else {
		for _, task := range j.Tasks {
			if task.Id == task_max {
				return &task
			}
		}
	}

	return &t
}
