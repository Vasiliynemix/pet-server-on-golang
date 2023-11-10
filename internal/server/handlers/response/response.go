package response

import "PetProjectGo/internal/config"

type Response struct {
	Status string      `json:"status"`
	Error  interface{} `json:"error,omitempty"`
}

func OK() Response {
	return Response{
		Status: config.StatusOK,
	}
}

func Error(err interface{}) Response {
	return Response{
		Status: config.StatusError,
		Error:  err,
	}
}
