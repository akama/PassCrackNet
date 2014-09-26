package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

func ListTasks(w http.ResponseWriter, r *http.Request) {
	session, _ := getCollection("tasks")
	defer session.Close()

	jobId := jobId(r, "job_id")

	job := loadJob(jobId, session)

	jsonResults, err := json.Marshal(job.Tasks)
	if err != nil {
		fmt.Println(err)
	}

	jsonWriter(w, jsonResults)
}

func ShowTask(w http.ResponseWriter, r *http.Request) {
	session, _ := getCollection("tasks")
	defer session.Close()

	IdJob := jobId(r, "job_id")
	taskId := jobId(r, "task_id")

	job := loadJob(IdJob, session)

	if len(job.Tasks) > 0 {
		for _, task := range job.Tasks {
			if task.Id == taskId {
				jsonResults, err := json.Marshal(task)
				if err != nil {
					fmt.Println(err)
				}
				jsonWriter(w, jsonResults)
			}
		}
	}
}

func TestPost(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()            // Parses the request body
	x := r.Form.Get("hello") // x will be "" if parameter is not set
	fmt.Println(x)

}

func FetchTask(w http.ResponseWriter, r *http.Request) {
	var taskRate int
	m := jsonProssesor(r)

	taskRate, err := strconv.Atoi(m["task_rate"].(string))
	if err != nil {
		errorWriter(w, err)
		return
	}

	job_id := jobId(r, "job_id")
	session := getConnection()
	job := loadJob(job_id, session)
	t := job.createTask(session, taskRate)
	jsonResult, err := json.Marshal(t)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("Sending out task %d\n", t.Id)


	jsonWriter(w, jsonResult)
}

func ReportResults(w http.ResponseWriter, r *http.Request) {

	m := jsonProssesor(r)

	fmt.Println(m)

	session := getConnection()

	job := loadJob(jobId(r, "job_id"), session)

	hash := m["Hash"].(string)
	salt := m["Salt"].(string)
	password := m["Password"].(string) 
	job.createResult(session, hash, salt, password)

	new_job := loadJob(jobId(r, "job_id"), session)

	fmt.Println(new_job)
}

func FinishTask(w http.ResponseWriter, r *http.Request) {
	session := getConnection()

	j := loadJob(jobId(r, "job_id"), session)

	task_id := jobId(r, "task_id")

	fmt.Printf("Job %d and task %d was being reported as finished.", j.Id, task_id)

	j.finishTask(task_id)
}

func jsonProssesor(r *http.Request) map[string]interface{} {
	var jsonResults map[string]interface{}

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
	}

	err = json.Unmarshal(b, &jsonResults)
	if err != nil {
		fmt.Println(err)
	}

	return jsonResults
}

func errorWriter(w http.ResponseWriter, err error) {
	fmt.Println(err)
	fmt.Fprintf(w, err.Error())
}