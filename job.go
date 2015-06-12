package main

import (
	"fmt"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"time"
)

type Job struct {
	Id         int       "_id" // uniq id for task
	AttackMode int       // 0 for dict, or 3 for hashmask.
	HashType   int       // type of hash.
	HashFile   []byte    // the file that contains hashes
	Mask       string    // mask or dict.
	Start      int       // start of number
	Finish     int       // Ending of string
	Tasks      []Task    // subset of tasks
	Results    []Result // sets of results
	Pause 	   bool	     // if the job is paused
	Name	   string    // Name of the job for humans
}


// Save the initial job to the database.
func (j *Job) save(session *mgo.Session) {
	c := session.DB(MONGO_DB).C("tasks")

	err := c.Insert(&j)

	if err != nil {
		panic(err)
	}
}

// Save the updated job to the database.
func (j *Job) update(con *mgo.Session) {
	c := con.DB(MONGO_DB).C("tasks")

	search := bson.M{"_id": j.Id}
	err := c.Update(search, j)

	if err != nil {
		panic(err)
	}

    con.Close()
}

// Checks if a job is done.
func (j *Job) IsDone() bool {
	if j.Start >= j.Finish {
		for _, task := range j.Tasks {
			if task.Done == false {
				return false
			}
		}
		return true
	} else {
		return false
	}
}

// Gets a list of all the tasks todo.
func (j *Job) TasksToDo() bool {
	if j.Start >= j.Finish {
		for _, task := range j.Tasks {
			if task.Done == false && time.Since(task.Timestamp).Minutes() > 15 {
				return true
			}
		}
		return false
	} else {
		return true
	}
}

// Checks if the job is paused.
func (j *Job) IsPaused() bool {
	return j.Pause
}


// HELPER FUNCTIONS

// Loads a job from the database based on job id.
func loadJob(id int, con *mgo.Session) *Job {
	c := con.DB(MONGO_DB).C("tasks")

	job := Job{}
	err := c.Find(bson.M{"_id": id}).One(&job)
	if err != nil {
		fmt.Println(err)
	}

	return &job
}

// Marks a task as finished.
func (j *Job) finishTask(task_id int) {
	j.Tasks[task_id-1].Done = true
	j.update(getConnection())


	fmt.Println("Task has been marked as finished.")
}

// Creates a job and saves it to the database.
func createJob(AttackMode int, HashType int, HashFile []byte, Mask string, max int, name string,  con *mgo.Session) (j Job) {
	c := con.DB(MONGO_DB).C("tasks")

	job := Job{}
	err := c.Find(nil).Sort("-_id").One(&job)
	if err != nil {
		fmt.Println(err)
	}

	id := job.Id + 1

	j = Job{
		Id:         id,
		AttackMode: AttackMode,
		HashType:   HashType,
		HashFile:   HashFile,
		Mask:       Mask,
		Start:      0,
		Finish:     max,	
		Pause: 	    false,
		Name:	    name,
		//Tasks: []JobSub{},
	}

	j.save(con)

    con.Close()

	return
}
