package main

import (
	"github.com/gorilla/mux"
	"net/http"
	"html/template"
	"fmt"
	"os/exec"
	// "io"
	"io/ioutil"
	"strconv"
	"strings"
	"labix.org/v2/mgo/bson"
	"code.google.com/p/gcfg"
)

type Config struct {
    Settings struct {
    	Port string
	    Location string
	    Addr string
    }
}

func main() {
	var cfg Config

	err := gcfg.ReadFileInto(&cfg, "settings.gcfg")
	if err != nil {
		fmt.Println("Test")
		panic(err)
	}

	WEB_PORT := cfg.Settings.Addr + ":" + cfg.Settings.Port

	r := mux.NewRouter()

	// Web Routes
	r.HandleFunc("/input_job", InputJobDisplay).Methods("GET")
	r.HandleFunc("/input_job", InputJobSubmit).Methods("POST")
	r.HandleFunc("/", mainJobPage) // Lists all the jobs
	r.HandleFunc("/jobs/{job_id}/tasks", jobTasks) // Lists all the tasks for a job
	r.HandleFunc("/jobs/{job_id}/results", jobResults) // Lists all the results for a job

	// Serve static files. 
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static",http.FileServer(http.Dir("static/"))))

	// API Routes
	s := r.PathPrefix("/api/jobs").Subrouter()
	s.HandleFunc("/", ListJobs)
	s.HandleFunc("", ListJobs)
	s.HandleFunc("/fetch", fetchJob)
	s.HandleFunc("/{job_id}", ShowJob)
	s.HandleFunc("/{job_id}/results", JobIsDone)
	s.HandleFunc("/{job_id}/fetch", FetchTask)
	s.HandleFunc("/{job_id}/tasks", ListTasks)
	s.HandleFunc("/{job_id}/pause_toggle", PauseToggle)
	s.HandleFunc("/{job_id}/tasks/{task_id}", ShowTask)
	s.HandleFunc("/{job_id}/tasks/{task_id}/results", ReportResults).Methods("POST")
	s.HandleFunc("/{job_id}/tasks/{task_id}/done", FinishTask)

	http.Handle("/", r)
	http.ListenAndServe(WEB_PORT, nil)
}

func InputJobSubmit(w http.ResponseWriter, r *http.Request){
    // name := r.PostFormValue("name")
   	r.ParseMultipartForm(32 << 20)
    file, _, err := r.FormFile("hashfile")
    
    if err != nil {
        fmt.Println(err)
        return
    }
    
	defer file.Close()
    
	output, err := ioutil.ReadAll(file)

	if err != nil {
			fmt.Println(err)
			return
	}

    attack, err := strconv.Atoi(r.FormValue("attack"))

    if err != nil {
    	fmt.Println(err)
    	return
    }

    hashtype, err := strconv.Atoi(r.FormValue("hashtype"))

    if err != nil {
    	fmt.Println(err)
    	return
    }

    if r.FormValue("passcode") != "testcode" {
        fmt.Println("No Auth.")
    }

    maxLimit := runHashcat(attack, hashtype, r.FormValue("mask"))

	createJob(attack, hashtype, output, r.FormValue("mask"), maxLimit, r.FormValue("name"), getConnection())

	fmt.Fprintf(w, "")
}

func InputJobDisplay(w http.ResponseWriter, r *http.Request){
	t := template.New("Job Input")
	t, err := template.ParseFiles("templates/job_input.html")

	if (err != nil) {
		panic(err)
	}

	t.Execute(w, nil)
}

func runHashcat(attackMode int, hashMode int, mask string) (result int) {
	var cfg Config

	err := gcfg.ReadFileInto(&cfg, "settings.gcfg")
	if err != nil {
		fmt.Println("Test")
		panic(err)
	}

	fmt.Println("Starting hashcat task to get mask tasks.")

	command := cfg.Settings.Location

	result = 1

	// fmt.Println(command + " -a " + attackMode + " -m " + hashMode + " input.txt " + mask + " --force " + "--keyspace")

	out, err := exec.Command(command, "-a", strconv.Itoa(attackMode), "-m", strconv.Itoa(hashMode), "input.txt", mask, "--force",  "--keyspace").CombinedOutput()

	if err != nil {
		fmt.Println(err)
	}

	output := string(out)

	result, err = strconv.Atoi(strings.Split(output, "\n")[2])

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Got a length of ", result)

	return
}


// Used only to keep calculations out of the template. 
// Should never been in the db, only calculated on the fly.
type DisplayJob struct {
	Id         int
	Name 		string
	TasksDispatched	int
	Progress	float64
	ResultsFound	int
	Complete	bool
	Paused	bool
}

func mainJobPage(w http.ResponseWriter, r *http.Request){
	t := template.New("Jobs List")
	t, err := template.ParseFiles("templates/index.html")

	if (err != nil) {
		panic(err)
	}

	// Fetch all the jobs
	var results []Job

	session, c := getCollection("tasks")
	defer session.Close()

	err = c.Find(bson.M{}).All(&results)
	if err != nil {
		fmt.Println(err)
	}

	displayResults := make([]DisplayJob, len(results))

	// Create a display array for the webpage
	for i := 0; i < len(results); i++ {
		progress := (float64(results[i].Start) / float64(results[i].Finish)) * 100
		displayResults[i] = DisplayJob {
								results[i].Id,
								results[i].Name,
								len(results[i].Tasks),
								progress,
								//float64(int(progress * 10)) / 10,
								len(results[i].Results),
								results[i].IsDone(),
								results[i].IsPaused(),
							}
	}

	data := struct {
		DisplayJobs []DisplayJob
	}{ displayResults }

	t.Execute(w, data)
	return
}

func jobTasks(w http.ResponseWriter, r *http.Request){
	t := template.New("Task List")
	t, err := template.ParseFiles("templates/tasks.html")

	if (err != nil) {
		panic(err)
	}

	// Fetch the job
	session, _ := getCollection("tasks")
	defer session.Close()

	jobId := jobId(r, "job_id")

	job := loadJob(jobId, session)

	data := struct {
		Tasks []Task
	}{ job.Tasks }

	t.Execute(w, data)
	return
}

func jobResults(w http.ResponseWriter, r *http.Request){
	t := template.New("Results List")
	t, err := template.ParseFiles("templates/results.html")

	if (err != nil) {
		panic(err)
	}

	// Fetch the job
	session, _ := getCollection("tasks")
	defer session.Close()

	jobId := jobId(r, "job_id")

	job := loadJob(jobId, session)

	data := struct {
		Results []Result
	}{ job.Results }

	t.Execute(w, data)
	return
}
