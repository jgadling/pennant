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

func runHttp(conf *Config, fc *FlagCache, driver StorageDriver) {
	router := pennantRouter(conf, fc, driver)
	handler := handlers.LoggingHandler(os.Stdout, router)

	graceful.AddSignal(syscall.SIGTERM)
	server := graceful.Server{
		Addr:    fmt.Sprintf(":%d", conf.HttpPort),
		Handler: handler,
	}
	err := server.ListenAndServe()
	if err != nil {
		logger.Criticalf("Fatal: %v", err)
	}
	graceful.Wait()
}

func pennantRouter(conf *Config, fc *FlagCache, driver StorageDriver) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	router.Methods("GET").Path("/flagValue/{name}").Handler(FlagValueHandler(fc))
	router.Methods("GET").Path("/flags").Handler(ListFlags(fc))
	router.Methods("POST").Path("/flags").Handler(SaveFlag(driver))
	router.Methods("DELETE").Path("/flags/{name}").Handler(DeleteFlag(fc, driver))
	router.Methods("GET").Path("/flags/{name}").Handler(GetFlag(fc))
	return router
}

type FlagValueResponse struct {
	Status  int    `json:"status"`
	Enabled bool   `json:"enabled"`
	Message string `json:"message"`
}

type FlagListResponse struct {
	Status int      `json:"status"`
	Flags  []string `json:"flags"`
}

type FlagItemResponse struct {
	Status int   `json:"status"`
	Flag   *Flag `json:"flag"`
}

func send(w http.ResponseWriter, status int, resp interface{}) {
	w.WriteHeader(status)
	b, _ := json.Marshal(resp)
	w.Write([]byte(b))
}

func SaveFlag(driver StorageDriver) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)
		r.Body.Close()

		flag, err := LoadAndParseFlag(body)
		if err != nil {
			send(w, 400, FlagValueResponse{
				Status:  400,
				Message: fmt.Sprintf("Error: %v", err),
			})
			return
		}
		re := regexp.MustCompile("^[0-9a-z_-]+$")
		logger.Infof("re is %v", re)
		if len(flag.Name) < 3 || len(flag.Name) > 120 || !re.MatchString(flag.Name) {
			send(w, 400, FlagValueResponse{
				Status:  400,
				Message: "Error: flag name must be 3-120 chars, letters, numbers, underscores and dashes are allowed",
			})
			return
		}

		driverErr := driver.saveFlag(flag)
		if driverErr != nil {
			send(w, 500, FlagValueResponse{
				Status:  500,
				Message: fmt.Sprintf("Error: could not write flag %v", err),
			})
			return
		}
		send(w, 200, FlagValueResponse{
			Status:  200,
			Enabled: true,
		})
		return
	})
}

func ListFlags(fc *FlagCache) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		flagList := make([]string, 0)
		for _, flag := range fc.List() {
			flagList = append(flagList, flag.Name)
		}
		response := FlagListResponse{
			Status: 200,
			Flags:  flagList,
		}
		send(w, response.Status, response)
	})
}

func GetFlag(fc *FlagCache) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		flagName := vars["name"]
		flag, err := fc.Get(flagName)
		if err != nil {
			// The flag didn't exist in the cache, let's send a 404
			send(w, 404, FlagValueResponse{
				Status:  404,
				Enabled: false,
			})
			return
		}
		response := FlagItemResponse{
			Status: 200,
			Flag:   flag,
		}
		send(w, response.Status, response)
	})
}

func DeleteFlag(fc *FlagCache, driver StorageDriver) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		flagName := vars["name"]
		_, err := fc.Get(flagName)
		if err != nil {
			// The flag didn't exist in the cache, let's send a 404
			send(w, 404, FlagValueResponse{
				Status:  404,
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
			Enabled: true,
		})
	})
}

func FlagValueHandler(fc *FlagCache) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		flagName := vars["name"]
		flag, err := fc.Get(flagName)
		if err != nil {
			// The flag didn't exist in the cache, let's send a 404
			send(w, 404, FlagValueResponse{
				Status:  404,
				Enabled: false,
			})
			return
		}
		body, _ := ioutil.ReadAll(r.Body)
		r.Body.Close()

		datas := make(map[string]interface{})
		json.Unmarshal(body, &datas)
		logger.Warningf("datas is %v", datas)
		enabled := flag.GetValue(datas)
		send(w, 200, FlagValueResponse{
			Status:  200,
			Enabled: enabled,
		})
	})
}
