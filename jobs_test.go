package main

import (
	"encoding/json"
	"fmt"
	"labix.org/v2/mgo"
	"testing"
)

func TestJobs(t *testing.T) {
	const MONGO_URL = "localhost"

	// Setup the database
	session, err := mgo.Dial(MONGO_URL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	session.SetMode(mgo.Monotonic, true)
	// Insert data into the database

	job_file := []byte("72db86e4c73b9fabb4810562b236488e")
	job := createJob(3, 0, job_file, "?d?d?d?d", session.Copy())
	jobJson, err := json.Marshal(&job)

	if err != nil {
		t.Errorf("Error creating json for job (%s)", job)
	}

	expectJobString := "{\"Id\":9,\"AttackMode\":3,\"HashType\":0,\"HashFile\":\"NzJkYjg2ZTRjNzNiOWZhYmI0ODEwNTYyYjIzNjQ4OGU=\",\"Mask\":\"?d?d?d?d\",\"Start\":1,\"Finish\":1000}"

	if string(jobJson) != expectJobString {
		t.Errorf("CreateJob (%s) doesn't match expected string (%s).", string(jobJson), expectJobString)
	}

	fmt.Println("Pre-Tast")

	// fmt.Println(job)
	// fmt.Println("Post-Task")

	postJob := loadJob(job.Id, session.Copy())

	// fmt.Println(postJob)
	if &job != postJob {
		t.Errorf("Local Job (%s) and Remote Job (%s) don't match.", &job, postJob)
	}
}
