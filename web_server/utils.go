package main

import (
	"fmt"
	"net/http"
)

func ErrorResponse(resp http.ResponseWriter, errorCode int) {
	ret := fmt.Sprintf("{\"code\" : %d}", errorCode)
	resp.Write([]byte(ret))
}
