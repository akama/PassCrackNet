package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"net/http"
	"strconv"
)

const (
	// Database constants
	MONGO_URL = "localhost"
	MONGO_DB  = "jobs_dev"
)

func ListJobs(w http.ResponseWriter, r *http.Request) {
	var results []Job

	session, c := getCollection("tasks")
	defer session.Close()

	err := c.Find(bson.M{}).All(&results)
	if err != nil {
		fmt.Println(err)
	}

	jsonResults, err := json.Marshal(results)
	if err != nil {
		fmt.Println(err)
	}

	jsonWriter(w, jsonResults)

}

/*
	job_id is passed in though the url using mux
*/
func ShowJob(w http.ResponseWriter, r *http.Request) {
	var result Job

	session, c := getCollection("tasks")
	defer session.Close()

	jobId := jobId(r, "job_id")

	fmt.Println("Fetching ", jobId)

	err := c.Find(bson.M{"_id": jobId}).One(&result)
	if err != nil {
		fmt.Println(err)
	}

	jsonResult, err := json.Marshal(result)
	if err != nil {
		fmt.Println(err)
	}

	jsonWriter(w, jsonResult)
}

func fetchJob(w http.ResponseWriter, r *http.Request) {
	var results []Job

	session, c := getCollection("tasks")
	defer session.Close()

	iter := c.Find(nil).Iter()
	err := iter.All(&results)
	if err != nil {
		panic(err)
	}

	for _, job := range results {
		if job.TasksToDo() && !job.IsPaused() {
			jsonResult, err := json.Marshal(job)
			if err != nil {
				fmt.Println(err)
				break
			}
			jsonWriter(w, jsonResult)
		}
	}

	jsonResult, err := json.Marshal(Job{})
	if err != nil {
		panic(err)
	}

	jsonWriter(w, jsonResult)
}

/*
	job_id is passed in though the url using mux
*/
func JobIsDone(w http.ResponseWriter, r *http.Request) {
	var result Job
	sendJson := make(map[string]interface{})

	session, c := getCollection("tasks")
	defer session.Close()

	jobId := jobId(r, "job_id")

	err := c.Find(bson.M{"_id": jobId}).One(&result)
	if err != nil {
		fmt.Println(err)
	}

	sendJson["Id"] = result.Id
	sendJson["Done"] = result.IsDone()

	jsonResult, err := json.Marshal(sendJson)
	if err != nil {
		fmt.Println(err)
	}

	jsonWriter(w, jsonResult)
}

func PauseToggle(w http.ResponseWriter, r *http.Request) {
	jobId := jobId(r, "job_id")

	session, _ := getCollection("tasks")
	defer session.Close()


	job := loadJob(jobId, session)


	job.Pause = !job.Pause

	job.update(session)

	if job.Pause == true {
		fmt.Fprintf(w, "Job %d has been paused.", job.Id)
	} else {
		fmt.Fprintf(w, "Job %d has been unpaused.", job.Id)
	}
}

func jobId(r *http.Request, field string) (jobId int) {
	vars := mux.Vars(r)
	jobId, err := strconv.Atoi(vars[field])
	if err != nil {
		fmt.Println(err)
	}

	return
}

func jsonWriter(w http.ResponseWriter, json []byte) {
	w.Header().Set("Content-Length", strconv.Itoa(len(json)))
	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}


func getCollection(name string) (*mgo.Session, *mgo.Collection) {
	session := getConnection()
	c := session.DB(MONGO_DB).C(name)

	return session, c
}

func getConnection() *mgo.Session {
	// Setup the database
	session, err := mgo.Dial(MONGO_URL)
	if err != nil {
		panic(err)
	}

	session.SetMode(mgo.Monotonic, true)

	return session
}
