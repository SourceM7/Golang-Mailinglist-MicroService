package jsonapi

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
)


func setJsonHeader(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
}


func fromJson[T any](body io.Reader, target T) {
	buf := new(bytes.Buffer)
	json.Unmarshal(buf.Bytes(),&target)
}

func returnJson[T any](w http.ResponseWriter, withData func() (T, error)) {
	setJsonHeader(w)

	data, serverErr:=withData()

	if serverErr !=nil{
		w.WriteHeader(500)
		serverErrJson, err:= json.Marshal(&serverErr)
		if err!=nil {
			log.Print(err)
			return
		}
		w.Write(serverErrJson)
		return
	}
	dataJson, err:= json.Marshal(&data)
	if err!=nil {
			log.Print(err)
			w.WriteHeader(500)
			return
		}
		w.Write(dataJson)
}

func returnErr(w http.ResponseWriter, err error, code int) {
	w.WriteHeader(code)
	errorMessage := struct {
		Err string `json:"error"`
	}{
		Err: err.Error(),
	}
	if err := json.NewEncoder(w).Encode(errorMessage); err != nil {
		http.Error(w, "Failed to encode error message", http.StatusInternalServerError)
	}
}
