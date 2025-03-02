package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/UserNameShouldBeHere/FlussonicTask/internal/tasktracker"
)

type ErrorResponse struct {
	Errors string `json:"errors"`
}

type ResponseData struct {
	Status int
	Data   any
}

func WriteResponse(w http.ResponseWriter, responseData ResponseData) error {
	jsonData, err := json.Marshal(responseData)
	if err != nil {
		return err
	}

	w.WriteHeader(responseData.Status)
	_, err = w.Write(jsonData)
	if err != nil {
		return err
	}

	return nil
}

func ConvertStatusToStr(status tasktracker.Status) string {
	switch status {
	case tasktracker.NotInList:
		return "not in list"
	case tasktracker.Done:
		return "done"
	case tasktracker.Failed:
		return "failed"
	case tasktracker.Canceled:
		return "canceled"
	case tasktracker.Executing:
		return "executing"
	case tasktracker.Waiting:
		return "waiting"
	default:
		return "undefined"
	}
}
