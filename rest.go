package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"syscall"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/zenazn/goji/graceful"
)

// Run HTTP server
func runHTTP(conf *Config, fc *FlagCache, driver StorageDriver) {
	router := pennantRouter(fc, driver)
	handler := handlers.LoggingHandler(os.Stdout, router)

	graceful.AddSignal(syscall.SIGTERM)
	server := graceful.Server{
		Addr:    fmt.Sprintf(":%d", conf.HTTPPort),
		Handler: handler,
	}
	err := server.ListenAndServe()
	if err != nil {
		logger.Criticalf("Fatal: %v", err)
	}
	graceful.Wait()
}

// Configure routes
func pennantRouter(fc *FlagCache, driver StorageDriver) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	// Supporting a posted request body via POST and query string params via GET?
	router.Methods("GET").Path("/flagValue/{name}").Handler(FlagValueHandler(fc))
	router.Methods("POST").Path("/flagValue/{name}").Handler(FlagValueHandler(fc))

	// CRUD's for flags
	router.Methods("GET").Path("/flags").Handler(ListFlags(fc))
	router.Methods("POST").Path("/flags").Handler(SaveFlag(driver))
	router.Methods("DELETE").Path("/flags/{name}").Handler(DeleteFlag(fc, driver))
	router.Methods("GET").Path("/flags/{name}").Handler(GetFlag(fc))
	return router
}

// FlagValueResponse exposes whether or not a flag is enabled
type FlagValueResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Enabled bool   `json:"enabled"`
}

// FlagItemResponse is a single flag definition
type FlagItemResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Flag    *Flag  `json:"flag"`
}

// FlagListResponse is a list of flag definitions
type FlagListResponse struct {
	Status  int      `json:"status"`
	Message string   `json:"message"`
	Flags   []string `json:"flags"`
}

// Send a response back to the client
func send(w http.ResponseWriter, status int, resp interface{}) {
	w.WriteHeader(status)
	b, _ := json.Marshal(resp)
	w.Write([]byte(b))
}

// SaveFlag updates a flag
func SaveFlag(driver StorageDriver) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)
		r.Body.Close()

		flag, err := LoadAndParseFlag(body)
		if err != nil {
			send(w, 400, FlagItemResponse{
				Status:  400,
				Message: fmt.Sprintf("Error: %v", err),
				Flag:    flag,
			})
			return
		}
		re := regexp.MustCompile("^[0-9a-z_-]+$")
		if len(flag.Name) < 3 || len(flag.Name) > 120 || !re.MatchString(flag.Name) {
			send(w, 400, FlagItemResponse{
				Status:  400,
				Message: "Error: flag name must be 3-120 chars, letters, numbers, underscores and dashes are allowed",
				Flag:    flag,
			})
			return
		}

		driverErr := driver.saveFlag(flag)
		if driverErr != nil {
			send(w, 500, FlagItemResponse{
				Status:  500,
				Message: fmt.Sprintf("Error: could not write flag %v", err),
				Flag:    flag,
			})
			return
		}
		send(w, 200, FlagItemResponse{
			Status:  200,
			Message: "OK",
			Flag:    flag,
		})
		return
	})
}

// ListFlags returns a list of all flags
func ListFlags(fc *FlagCache) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		flagList := make([]string, 0)
		for _, flag := range fc.List() {
			flagList = append(flagList, flag.Name)
		}
		response := FlagListResponse{
			Status:  200,
			Message: "OK",
			Flags:   flagList,
		}
		send(w, response.Status, response)
	})
}

// GetFlag gets the definition of a single flag
func GetFlag(fc *FlagCache) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		flagName := vars["name"]
		flag, err := fc.Get(flagName)
		if err != nil {
			// The flag didn't exist in the cache, let's send a 404
			send(w, 404, FlagItemResponse{
				Status:  404,
				Message: "flag not found",
				Flag:    flag,
			})
			return
		}
		response := FlagItemResponse{
			Status:  200,
			Message: "OK",
			Flag:    flag,
		}
		send(w, response.Status, response)
	})
}

// DeleteFlag deletes a flag
func DeleteFlag(fc *FlagCache, driver StorageDriver) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		flagName := vars["name"]
		_, err := fc.Get(flagName)
		if err != nil {
			// The flag didn't exist in the cache, let's send a 404
			send(w, 404, FlagValueResponse{
				Status:  404,
				Message: "Not Found",
				Enabled: false,
			})
			return
		}

		driverErr := driver.deleteFlag(flagName)
		if driverErr != nil {
			send(w, 500, FlagValueResponse{
				Status:  500,
				Message: fmt.Sprintf("Error: could not delete flag %v", err),
			})
			return
		}

		send(w, 200, FlagValueResponse{
			Status:  200,
			Message: "OK",
			Enabled: true,
		})
	})
}

// FlagValueHandler performs policy evaluations on a document to determine whether a flag is
// enabled.
func FlagValueHandler(fc *FlagCache) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		flagName := vars["name"]
		flag, err := fc.Get(flagName)
		if err != nil {
			// The flag didn't exist in the cache, let's send a 404
			send(w, 404, FlagValueResponse{
				Status:  404,
				Message: "Flag not found",
				Enabled: false,
			})
			return
		}
		body, _ := ioutil.ReadAll(r.Body)
		r.Body.Close()

		// Support request body for POST
		datas := make(map[string]interface{})
		json.Unmarshal(body, &datas)

		// Support Query string data for GET
		// ONLY SUPPORTS STRINGS THO :'(
		queryData := r.URL.Query()
		for k, v := range queryData {
			if len(v) == 1 {
				datas[k] = v[0]
				continue
			}
			datas[k] = v
		}

		logger.Debugf("Document is %v", datas)
		enabled := flag.GetValue(datas)
		send(w, 200, FlagValueResponse{
			Status:  200,
			Message: "OK",
			Enabled: enabled,
		})
	})
}
