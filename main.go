package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

// Pathentry to store file/folder attributes
type pathEntryDetails struct {
	Fullpath    string    `json:"path"`
	IsDir       bool      `json:"isdirectory"`
	Name        string    `json:"name"`
	Permissions string    `json:"permissions"`
	Size        int64     `json:"size"`
	LastMod     time.Time `json:"lastmodified"`
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
var originalRootDir string
var pathSeparator string

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the HomePage!")
	fmt.Println("Endpoint Hit: homePage")

}

func handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)

	myRouter.Use(CORS)

	myRouter.HandleFunc("/", homePage)
	myRouter.HandleFunc("/ls/", returnDirectoryListingAtPath).Queries("path", "").Methods("GET")
	myRouter.HandleFunc("/ls/", returnDirectoryListingAtPath)
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

func pathNotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	io.WriteString(w, "Specified path not found")
}

func errorWhileReadingPath(w http.ResponseWriter, r *http.Request, err error) {
	w.WriteHeader(http.StatusBadRequest)
	io.WriteString(w, "Unexpected error occurred during operation:"+err.Error())
}

func returnCurrentDirectoryListing(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: returnCurrentDirectoryListing")
	var pe pathEntryDetails
	pe, _ = currentDirStat()
	json.NewEncoder(w).Encode(pe)
}

/*
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

}*/

func returnDirectoryListingAtPath(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimPrefix(r.URL.Query().Get("path"), "/")

	if key == "" {
		key = "."
		fmt.Println("Key was empty, now set to dot " + key)
	}

	fmt.Println("Endpoint Hit: returnDirectoryListingAtPath=" + key)
	wd, _ := os.Getwd()

	fmt.Println("Current workdir=" + wd)

	var entries []pathEntryDetails
	//pe, err := specificDirStat(key)

	tmpDir := key
	//os.Chdir(tmpDir)
	//fmt.Println(syscall.Chdir(tmpDir))
	//err :=

	if syscall.Chdir(tmpDir) != nil {
		pathNotFound(w, r)
		return
	}

	wd, _ = os.Getwd()
	fmt.Println("Current workdir2=" + wd)

	fmt.Println("")
	err := filepath.Walk(".", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}
		if info.IsDir() {

		} else {

		}
		fmt.Printf("visited file or dir: %q %q\n", wd, path)
		entry, err := specificDirStat(path)
		if err != nil {
			fmt.Printf("error walking the path %q: %v\n", tmpDir, err)
			return err
		}
		entries = append(entries, entry)
		return nil
	})
	fmt.Println(syscall.Chdir(originalRootDir))
	if err != nil {
		fmt.Printf("error walking the path %q: %v\n", tmpDir, err)
		errorWhileReadingPath(w, r, err)
		return
	}

	if err != nil {
		fmt.Printf("Erroar hear: %v", err)
	} else {
		json.NewEncoder(w).Encode(entries)
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

	fmt.Print("File Name:", fileStat.Name())         // Base name of the file
	fmt.Print(" Size:", fileStat.Size())             // Length in bytes for regular files
	fmt.Print(" Permissions:", fileStat.Mode())      // File mode bits
	fmt.Print(" Last Modified:", fileStat.ModTime()) // Last modification time
	fmt.Println(" Is Directory: ", fileStat.IsDir())

	wd, _ := os.Getwd()

	details.Fullpath = wd + pathSeparator + path
	details.IsDir = fileStat.IsDir()
	details.Name = fileStat.Name()
	details.Permissions = fileStat.Mode().String()
	details.Size = fileStat.Size()
	details.LastMod = fileStat.ModTime()

	return details, nil
}

func main() {
	wd, _ := os.Getwd()
	originalRootDir = wd

	if runtime.GOOS == "windows" {
		pathSeparator = "\\"
	}

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
