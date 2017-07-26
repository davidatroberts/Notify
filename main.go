package main

import (
	"Notify/Observer"
	"net/http"

	"fmt"

	"github.com/gorilla/mux"
)

func handleRequests(subject *Observer.Subject) {
	router := mux.NewRouter()
	router.HandleFunc(
		"/notify/phone/msg",
		func(w http.ResponseWriter, req *http.Request) {
			msg := Observer.MessageEvent{Message: "received"}
			subject.NotifyObservers("text", msg)
		},
	).Methods("POST")

	http.ListenAndServe(":8080", router)
}

func main() {
	// create the subject with single channel
	subject := Observer.NewSubject()

	// create ghost observer to wait for the message
	ghost := Observer.Observer{
		Chnl: subject.AddObserver("text"),
		Handler: func(event Observer.Event) {
			messageEvent, ok := event.(Observer.MessageEvent)
			if ok {
				fmt.Printf("Ghost got this message: %s\n", messageEvent.Message)
			}
		},
	}
	ghost.Process()

	storm := Observer.Observer{
		Chnl: subject.AddObserver("text"),
		Handler: func(event Observer.Event) {
			messageEvent, ok := event.(Observer.MessageEvent)
			if ok {
				fmt.Printf("Storm got this message: %s\n", messageEvent.Message)
			}
		},
	}
	storm.Process()

	handleRequests(subject)
}
