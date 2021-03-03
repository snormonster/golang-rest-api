package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

// Pathentry to store file/folder attributes
type pathEntryDetails struct {
	Path        string    `json:"path"`
	Filename    string    `json:"filename"`
	Permissions string    `json:"permissions"`
	Size        int64     `json:"size"`
	LastMod     time.Time `json:"lastmodified"`
	IsDir       bool      `json:"isdirectory"`
}

type failedResponse struct {
	Method  string `json:"method"`
	Request string `json:"request"`
	Error   string `json:"error"`
}

type healthResponse struct {
	Status                string    `json:"status"`
	CurrentTime           time.Time `json:"timestamp"`
	LastActivityTimestamp time.Time `json:"lastactivity"`
}

var lastActivity time.Time

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the HomePage!")
	fmt.Println("Endpoint Hit: homePage")

}

func handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)

	myRouter.Use(CORS)

	myRouter.HandleFunc("/", homePage)
	myRouter.HandleFunc("/ls/", returnDirectoryListingAtPath).Queries("path", "").Methods("GET")
	//myRouter.HandleFunc("/ls/", returnCurrentDirectoryListing)
	//myRouter.HandleFunc("/article/", returnSingleArticle).Methods("GET").Queries("foo", "bar", "id", "{id:[0-9]+}")
	myRouter.HandleFunc("/test/", printpathtest).Queries("foo", "").Methods("GET")
	myRouter.HandleFunc("/env/", printenvvar).Methods("GET")
	myRouter.HandleFunc("/health", healthCheck).Methods("GET")
	myRouter.NotFoundHandler = http.HandlerFunc(notFound)

	// finally, instead of passing in nil, we want
	// to pass in our newly created router as the second
	// argument
	//log.Fatal(http.ListenAndServe(":"+os.Getenv("LOCAL_PORT"), myRouter))
	fmt.Println("Starting web service API on local port 10000")
	//fmt.Println("Starting web service API on local port" + os.Getenv("LOCAL_PORT")
	lastActivity = time.Now()
	log.Fatal(http.ListenAndServe(":10000", myRouter))
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	var health healthResponse
	health.CurrentTime = time.Now()
	health.Status = "OK"
	health.LastActivityTimestamp = lastActivity
	lastActivity = time.Now()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(health)
}

func notFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)

	var failed failedResponse

	failed.Error = "Invalid API call"

	if r.Method != "GET" {
		failed.Error = "Invalid HTTP Method"
	}

	failed.Method = r.Method
	failed.Request = ""

	json.NewEncoder(w).Encode(failed)
}

func pathnotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	io.WriteString(w, "Specified path not found")
	//http.ServeFile(w, r, "public/index.html")
}

func returnCurrentDirectoryListing(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: returnCurrentDirectoryListing")
	var pe pathEntryDetails
	pe, _ = currentDirStat()
	json.NewEncoder(w).Encode(pe)
}

func returnDirectoryListingAtPath(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimPrefix(r.URL.Query().Get("path"), "/")

	fmt.Println("Endpoint Hit: returnDirectoryListingAtPath=" + key)
	var pe pathEntryDetails
	pe, err := specificDirStat(key)

	if err != nil {
		print("!!!!!!!!!!!!!!")
	} else {
		json.NewEncoder(w).Encode(pe)
	}

}

func printpathtest(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: printpathtest")
	//json.NewEncoder(w).Encode(r.URL.Query().Get("foo"))
	io.WriteString(w, "default")
}

func printenvvar(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: printenvvar")
	//json.NewEncoder(w).Encode(r.URL.Query().Get("foo"))
	io.WriteString(w, os.Getenv("env"))
}

func currentDirStat() (pathEntryDetails, error) {
	return fileStat(".")
}

func specificDirStat(str string) (pathEntryDetails, error) {
	return fileStat(str)
}

func fileStat(path string) (pathEntryDetails, error) {
	var details pathEntryDetails

	fileStat, err := os.Stat(path)

	if err != nil {
		if os.IsNotExist(err) {
			return details, fmt.Errorf("File or folder does not exist") //	"Folder does not exist"
		}
		log.Panic(err)
		return details, err
	}

	fmt.Println("File Name:", fileStat.Name())       // Base name of the file
	fmt.Print(" Size:", fileStat.Size())             // Length in bytes for regular files
	fmt.Print(" Permissions:", fileStat.Mode())      // File mode bits
	fmt.Print(" Last Modified:", fileStat.ModTime()) // Last modification time
	fmt.Print(" Is Directory: ", fileStat.IsDir())

	details.Filename = fileStat.Name()
	details.Size = fileStat.Size()
	details.LastMod = fileStat.ModTime()
	details.IsDir = fileStat.IsDir()
	details.Permissions = fileStat.Mode().String()
	details.Path = path

	return details, nil
}

func main() {
	//fmt.Println("Application started.\nVersion=" + os.Getenv("VERSION"))
	fmt.Println(time.Now().String() + ": Application started.\nVersion=1.0")

	handleRequests()
	//fmt.Println("Listening on container port" + os.Getenv("LOCAL_PORT"))

}

// CORS Middleware
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set headers
		w.Header().Set("Access-Control-Allow-Headers:", "*")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
		return
	})
}
