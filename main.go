package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	r "github.com/christopherhesse/rethinkgo"
)

var sessions []*r.Session

const (
	dbName string = "movies"
	host   string = "localhost:28015"
)

type Movie struct {
	Title  string
	Year   int
	Actors []*Actor
}

type Actor struct {
	Name string
	Age  int
}

func init() {
	session, err := r.Connect(host, dbName)
	if err != nil {
		log.Fatal("Error creating new session: %s", err)
	}
	var databases []string
	err = r.DbList().Run(session).All(&databases)
	if err != nil {
		log.Printf("Error getting database names: %s", err)
	}
	if len(databases) == 1 {
		err = r.DbDrop(dbName).Run(session).Exec()
		if err != nil {
			log.Printf("Error dropping existing database: %s", err)
		}
	}
	err = r.DbCreate(dbName).Run(session).Exec()
	if err != nil {
		log.Println("Error creating database ", dbName)
	}
	err = r.TableCreate("actors").Run(session).Exec()
	if err != nil {
		log.Println("Error creating table actors")
	}

	sessions = append(sessions, session)
}

func handleIndex(res http.ResponseWriter, req *http.Request) {
	session := sessions[0]
	var actors []Actor

	err := r.Table("actors").Run(session).All(&actors)
	if err != nil {
		log.Fatal("Error getting all actors: %s", err)
		return
	}
	data, err := json.Marshal(actors)
	if err != nil {
		log.Fatalf("Error marshalling json")
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.Write(data)
}

func insertActor(res http.ResponseWriter, req *http.Request) {
	session := sessions[0]

	actor := Actor{}
	err := json.NewDecoder(req.Body).Decode(&actor)
	if err != nil {
		log.Println("Erro decoding body from request")
		res.Header().Set("Content-Type", "application/json; charset=utf-8")
		res.Write([]byte("erro"))
	}

	var response r.WriteResponse
	err = r.Table("actors").Insert(actor).Run(session).One(&response)
	if err != nil {
		log.Fatal(err)
		return
	}
	var data []byte
	data, err = json.Marshal("{'actor':'saved'}")
	if err != nil {
		log.Fatal("Problem inserting actor")
		return
	}
	res.Header().Set("Content-Type", "application/json; charset=utf-8")
	res.Write(data)
}

func updateActor(res http.ResponseWriter, req *http.Request) {
	session := sessions[0]
	name := req.FormValue("name")

	var response map[string]interface{}
	err := r.Table("actors").Get(name).Run(session).One(&response)
	if err != nil {
		log.Fatal("error getting actor on update")
		return
	}
	fmt.Println(response)
}

func deleteActor(res http.ResponseWriter, req *http.Request) {
	session := sessions[0]
	name := req.FormValue("name")
	fmt.Println(name)

	var response r.WriteResponse
	row := r.Map{"Name": name}
	err := r.Table("actors").Filter(row).Delete().Run(session).One(&response)
	if err != nil {
		log.Fatal("Problem removing actor")
		http.Error(res, http.StatusText(404), 404)
	}

	var data []byte
	data, err = json.Marshal("{'actor':'deleted'}")
	res.Header().Set("Content-Type", "application/json; charset=utf-8")
	res.Write(data)
}

func main() {
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/new", insertActor)
	http.HandleFunc("/delete", deleteActor)
	//http.HandleFunc("/update", updateActor)

	err := http.ListenAndServe(":5000", nil)
	if err != nil {
		log.Fatal("Error: %v", err)
	}
}
