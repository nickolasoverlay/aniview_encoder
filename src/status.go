package src

import (
	"encoding/json"
	"net/http"
)

func QueueStats(w http.ResponseWriter, r *http.Request) {
	length := len(inQueue)
	waiting := inQueue
	isTaskInProcess := curTask.Status == StatusInProcess

	s := Stats{
		Length: length,
		InProcess: InProcess{
			IsRunning: isTaskInProcess,
			Task:      curTask,
		},
		Waiting: waiting,
	}

	if isTaskInProcess {
		s.Length += 1
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s)
}
