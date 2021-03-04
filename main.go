package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

// PathEntryDetails to store file/folder attributes
type PathEntryDetails struct {
	Fullpath    string    `json:"path"`
	IsDir       bool      `json:"isdirectory"`
	Name        string    `json:"name"`
	Permissions string    `json:"permissions"`
	Size        int64     `json:"size"`
	LastMod     time.Time `json:"lastmodified"`
}

// HealthResponse stores simple health check values
type HealthResponse struct {
	Status                string    `json:"status"`
	CurrentTime           time.Time `json:"timestamp"`
	LastActivityTimestamp time.Time `json:"lastactivity"`
}

// PathSeparator has to be dynamically set
var PathSeparator string
var lastActivity time.Time
var originalWorkDir string
var file os.File

func main() {
	initCloseHandler()
	initLogger()
	handleLogging("Application started.", "INFO")

	wd, _ := os.Getwd()
	originalWorkDir = wd

	if runtime.GOOS == "windows" {
		PathSeparator = "\\"
	} else {
		PathSeparator = "/"
	}

	handleRequests()
	cleanUp("main")
}

func handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.Use(CORS)
	myRouter.HandleFunc("/", homePage)
	myRouter.HandleFunc("/ls", returnDirectoryListingAtPath).Queries("path", "").Methods("GET")
	myRouter.HandleFunc("/health", healthCheck).Methods("GET")
	myRouter.NotFoundHandler = http.HandlerFunc(unexpectedRoute)

	handleLogging("Starting web service API on local port: 8080", "INFO")

	lastActivity = time.Now() // set timestamp before getting blocked by ListenAndServe
	log.Fatal(http.ListenAndServe(":8080", myRouter))
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the HomePage!")
	handleLogging("Hit endpoint: homePage", "DEBUG")
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	handleLogging("Hit endpoint: healthCheck", "DEBUG")

	var health HealthResponse
	health.CurrentTime = time.Now()
	health.Status = "OK"
	health.LastActivityTimestamp = lastActivity
	lastActivity = time.Now()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(health)
}

func unexpectedRoute(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	io.WriteString(w, "Invalid, unknown, or unauthorized API call")
}

func pathNotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	io.WriteString(w, "Specified path not found")
}

func errorWhileReadingPath(w http.ResponseWriter, r *http.Request, err error) {
	w.WriteHeader(http.StatusBadRequest)
	io.WriteString(w, "Unexpected error occurred during operation: "+err.Error())
}

func returnDirectoryListingAtPath(w http.ResponseWriter, r *http.Request) {
	specifiedPath := strings.TrimPrefix(r.URL.Query().Get("path"), "/")

	handleLogging("Hit endpoint: returnDirectoryListingAtPath="+specifiedPath, "DEBUG")

	// if no explicit path is specified, set to relative root
	if specifiedPath == "" {
		specifiedPath = "."
	}

	var entries []PathEntryDetails
	tmpDir := specifiedPath

	// quick check to see if path exists
	if syscall.Chdir(tmpDir) != nil {
		pathNotFound(w, r)
		return
	}

	err := filepath.Walk(".", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			handleLogging("Failure accessing a path="+path+" error="+err.Error(), "ERROR")
			return err
		}

		entry, err := getFileInformation(path)

		if err != nil {
			handleLogging("Error accessing path="+path+" error="+err.Error(), "ERROR")
			return err
		}

		entries = append(entries, entry)
		return nil
	})

	syscall.Chdir(originalWorkDir)

	if err != nil {
		handleLogging("Error while building directory listing, error="+err.Error(), "ERROR")
		errorWhileReadingPath(w, r, err)
	} else {
		json.NewEncoder(w).Encode(entries)
	}
}

func getFileInformation(path string) (PathEntryDetails, error) {
	var details PathEntryDetails

	fileStat, err := os.Stat(path)

	if err != nil {
		if os.IsNotExist(err) {
			return details, fmt.Errorf("File or folder does not exist")
		}
		log.Panic(err)
		return details, err
	}

	wd, _ := os.Getwd()

	details.Fullpath = wd + PathSeparator + path
	details.IsDir = fileStat.IsDir()
	details.Permissions = fileStat.Mode().String()
	details.Size = fileStat.Size()
	details.LastMod = fileStat.ModTime()

	// fileStat returns '.' for folder names, give proper name instead
	if fileStat.Name() != "." {
		details.Name = fileStat.Name()
	} else {
		elements := strings.SplitAfter(wd, PathSeparator)
		lastElement := elements[len(elements)-1]
		details.Name = lastElement
	}

	return details, nil
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

func initCloseHandler() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		handleLogging("App stopped unexpected externally, will cleanup and exit", "WARN")
		cleanUp("external")
		os.Exit(0)
	}()
}

func cleanUp(str string) {
	handleLogging("Application stopped from "+str, "DEBUG")
	file.Close()
}

func initLogger() {
	file, err := os.OpenFile("logfile.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	log.SetOutput(file)
}

func handleLogging(msg, level string) {
	log.Print(level + " " + msg)

	if level == "DEBUG" {
		if os.Getenv("NODEBUG") != "true" {
			fmt.Println(msg)
		}
	} else {
		fmt.Println(msg)
	}
}
