package main

import (
	"fmt"
	"log"
	"net/http"
)

func homePage(response http.ResponseWriter, request *http.Request) {
	fmt.Fprintf(response, "<h1>Главная страница сайта</h1>")
}

func aboutPage(response http.ResponseWriter, request *http.Request) {
	fmt.Fprintf(response, "<h1>О нас</h1>")
}

func pingHandler(response http.ResponseWriter, request *http.Request) {
	if request.Method == http.MethodGet {
		response.WriteHeader(http.StatusOK)
		fmt.Fprintf(response, "<h1>200 pong!</h1>")
	} else {
		http.Error(response, "Данный метод не разрешен", http.StatusMethodNotAllowed)
	}
}

func main() {
	http.HandleFunc("/", homePage)
	http.HandleFunc("/about", aboutPage)
	http.HandleFunc("/ping", pingHandler)

	fmt.Println("Сервер работает на порту 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
