package main

import (
	"fmt"
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
		fmt.Fprintf(response, "<h1>200 pong!</h1>")
		response.WriteHeader(http.StatusOK)
	} else {
		http.Error(response, "Данный метод не разрешен", http.StatusMethodNotAllowed)
	}
}

func main() {
	http.HandleFunc("/", homePage)
	http.HandleFunc("/about", aboutPage)
	http.HandleFunc("/ping", pingHandler)

	fmt.Println("Сервер работает на порту 8080")
	http.ListenAndServe(":8080", nil)
}
