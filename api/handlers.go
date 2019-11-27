package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/emanueljoivo/arrebol/storage"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"net/http"
)

const VersionTag = "0.0.1"
const VersionName = "Havana"

type Version struct {
	Tag  string `json:"Tag"`
	Name string `json:"Name"`
}

type QueueResponse struct {
	ID string `json:"ID"`
	Name string `json:"Name"`
	PendingTasks uint `json:"PendingTasks"`
	RunningTasks uint `json:"RunningTasks"`
	Nodes uint `json:"Nodes"`
	Workers uint `json:"Workers"`
}

type JobSpec struct {
	Label string `json:"Label"`
	Tasks []TaskSpec `json:"Tasks"`
}

type TaskSpec struct {
	ID string  `json:"ID"`
	Config map[string]interface{} `json:"Config"`
	Commands []string `json:"Commands"`
}

var (
	ProcReqErr = errors.New("error while trying to process response")
	EncodeResErr = errors.New("error while trying encode response")
)

func (a *API) CreateQueue(w http.ResponseWriter, r *http.Request) {
	var q storage.Queue

	err := json.NewDecoder(r.Body).Decode(&q)

	if err != nil {
		log.Println(ProcReqErr)
	}

	id := primitive.NewObjectID()
	q.ID = id

	_, err = a.storage.SaveQueue(&q)

	if err != nil {
		write(w, http.StatusInternalServerError, notOkResponse(err.Error()))
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = fmt.Fprintf(w, `{"ID": "%s"}`, id.Hex())
	}
}

func (a *API) RetrieveQueue(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	queueId := params["qid"]

	queue, err := a.storage.RetrieveQueue(queueId)

	if err != nil {
		write(w, http.StatusInternalServerError, notOkResponse(err.Error()))
	} else {
		response := responseFromQueue(queue)
		write(w, http.StatusOK, response)
	}
}

func (a *API) RetrieveQueues(w http.ResponseWriter, r *http.Request) {
	var response []*QueueResponse

	queues, err := a.storage.RetrieveQueues()

	if err != nil {
		write(w, http.StatusInternalServerError, notOkResponse(err.Error()))
	} else {
		for i := 0; i < len(queues); i++ {
			curQueue := responseFromQueue(queues[i])
			response = append(response, curQueue)
		}
		write(w, http.StatusOK, response)
	}
}

func (a *API) CreateJob(w http.ResponseWriter, r *http.Request) {
	var jobSpec JobSpec

	err := json.NewDecoder(r.Body).Decode(&jobSpec)

	if err != nil {
		log.Println(ProcReqErr)
	}

	id := primitive.NewObjectID()

	job := jobFromSpec(jobSpec, id)

	_, err = a.storage.SaveJob(job, id)

	if err != nil {
		write(w, http.StatusInternalServerError, notOkResponse(err.Error()))
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = fmt.Fprintf(w, `{"ID": "%s"}`, id.Hex())
	}
}

func (a *API) GetVersion(w http.ResponseWriter, r *http.Request) {
	write(w, http.StatusOK, Version{Tag: VersionTag, Name: VersionName})
}

func responseFromQueue(queue *storage.Queue) *QueueResponse {
	var pendingTasks uint
	var runningTasks uint
	jobs := queue.Jobs

	for i := 0; i < len(jobs); i++ {
		tasks := jobs[i].Tasks
		for j := 0; j< len(tasks); j++ {
			if tasks[j].State == storage.Pending {
				pendingTasks++
			}
			if tasks[j].State == storage.Running {
				runningTasks++
			}
		}
	}

	return &QueueResponse{
		ID: queue.ID.Hex(),
		Name: queue.Name,
		PendingTasks: pendingTasks,
		RunningTasks: runningTasks,
		Nodes: uint(len(queue.Nodes)),
	}
}

func jobFromSpec(jobSpec JobSpec, id primitive.ObjectID) *storage.Job {
	var tasks []storage.Task

	taskSpecs := jobSpec.Tasks

	for i := 0; i< len(taskSpecs); i++ {
		tasks = append(tasks, storage.Task{
			ID:       taskSpecs[i].ID,
			Config:   taskSpecs[i].Config,
			Commands: taskSpecs[i].Commands,
			State:    storage.Pending,
		})
	}

	return &storage.Job{
		ID:    id,
		Label: jobSpec.Label,
		State: storage.Pending,
		Tasks: tasks,
	}
}

func notOkResponse(err string) []byte {
	return []byte(`{"Message": ` + err)
}

func write(w http.ResponseWriter, statusCode int, i interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(i); err != nil {
		log.Println(EncodeResErr)
	}
}