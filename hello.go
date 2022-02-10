package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"strconv"
	"strings"
	"fmt"
	"bufio"
	"os"
)

type ApiResponseJSON struct {
	UserID int `json:"userId"`
	Id int `json:"id"`
	Title string `json:"title"`
	Body string `json:"body"`
}

type ValidResult struct {
	resp ApiResponseJSON
	valid bool
}

func makeApiRequest(url string) (ApiResponseJSON, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	var decodedResp = ApiResponseJSON{}
	err = json.Unmarshal(body, &decodedResp)

	return decodedResp, err
}

func apiWorker(jobs <-chan int, results chan<- ApiResponseJSON) {
	for {
		jobID, more := <-jobs
		if more {
			result, err := makeApiRequest("https://jsonplaceholder.typicode.com/posts/" + strconv.Itoa(jobID))
			if err != nil {
				log.Fatalln(err)
				return
			} else {
				results<-result
			}
		} else {
			return
		}
	}
}


func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter validity search term: ")
	searchTerm, _ := reader.ReadString('\n')
	searchTerm = strings.Replace(searchTerm, "\n", "", -1)

	const numberOfRoutines = 3
	const numberOfJobs = 100
	jobs := make(chan int)
	results := make(chan ApiResponseJSON)

	// Create workers for making api calls
	var wg sync.WaitGroup
	for i := 0; i < numberOfRoutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			apiWorker(jobs, results)
		}()
	}

	// queue each url we're intereted in
	go func() {
		for i := 1; i <= numberOfJobs; i++ {
			jobs <- i
		}
		close(jobs)
	}()
	
	// gather results
	var responses []ApiResponseJSON
	for i := 1; i <= numberOfJobs; i++ {
		jobResult := <-results
		responses = append(responses, jobResult)
	}
	wg.Wait()

	// get validity of each song
	var answer []ValidResult
	for _, response := range responses {
		containsInvalidWord := strings.Contains(response.Body, searchTerm)
		answer = append(answer, ValidResult{response, containsInvalidWord})
	}

	// print number of valid and invalid strings
	var validCount int = 0
	var invalidCount int = 0
	for _, ans := range answer {
		if ans.valid {
			validCount += 1
		} else {
			invalidCount += 1
		}
	}
	fmt.Println("Number of valid API responses: ", validCount)
	fmt.Println("Number of invalid API responses: ", invalidCount)
}
