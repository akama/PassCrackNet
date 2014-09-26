package main

import (
	"labix.org/v2/mgo"
)

type ReportedResults struct {
	job_id	int `json:"job_id"`
	task_id	int `json:"task_id"`
	result []Result `json:"result"`
}

type Result struct {
	Hash string
	Salt string
	Password string
}

func (j *Job) createResult(session *mgo.Session, hash string, salt string, password string) *Result {
	var r Result

	r = Result {
		Hash : hash,
		Salt : salt,
		Password : password,
	}

	present := false

	for _, result := range j.Results {
		if (result.Hash == hash && result.Salt == salt && result.Password == password) {
			present = true
			break
		}
	}

	if (!present) {
		j.Results = append(j.Results, r)

		j.update(session)
	}
	return &r
}
