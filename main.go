package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/lib/pq"

	"github.com/graphql-go/graphql"
)

type Patient struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

func logFatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

var db *sql.DB

func main() {

	pgURL, err := pq.ParseURL(os.Getenv("DB_URL"))
	logFatal(err)

	db, err = sql.Open("postgres", pgURL)
	logFatal(err)

	err = db.Ping()
	logFatal(err)

	//step 1, a patientType

	var patientType = graphql.NewObject(
		graphql.ObjectConfig{
			Name:        "Patient",
			Description: "This is a patient type.",
			Fields: graphql.Fields{
				"id": &graphql.Field{
					Type: graphql.Int,
				},
				"name": &graphql.Field{
					Type: graphql.String,
				},
				"email": &graphql.Field{
					Type: graphql.String,
				},
				"phone": &graphql.Field{
					Type: graphql.String,
				},
			},
		},
	)

	//step 2, a queryType --- queries the database / does not modify/mutate the data

	var queryType = graphql.NewObject(
		graphql.ObjectConfig{
			Name: "Query",
			Fields: graphql.Fields{
				"getPatient": &graphql.Field{
					Type:        patientType,
					Description: "Get a patient by id",
					Args: graphql.FieldConfigArgument{
						"id": &graphql.ArgumentConfig{
							Type: graphql.Int,
						},
					},
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						id, _ := p.Args["id"].(int)
						patient := &Patient{}

						err := db.QueryRow("select id, name, email, phone from patients where id=$1", id).
							Scan(&patient.ID, &patient.Name, &patient.Email, &patient.Phone)

						logFatal(err)

						return patient, nil
					},
				},
				"getPatients": &graphql.Field{
					Type:        graphql.NewList(patientType),
					Description: "Gets a patient list",
					Resolve: func(params graphql.ResolveParams) (interface{}, error) {
						var patients []*Patient

						stmt := "select * from patients"
						rows, err := db.Query(stmt)
						logFatal(err)

						for rows.Next() {
							patient := &Patient{}

							err = rows.Scan(&patient.ID, &patient.Name, &patient.Email,
								&patient.Phone)

							patients = append(patients, patient)
						}

						return patients, nil
					},
				},
			},
		},
	)

	//step 3, a mutationType --- queries the database / but it changes/mutates the data

	var mutationType = graphql.NewObject(
		graphql.ObjectConfig{
			Name: "Mutations",
			Fields: graphql.Fields{
				"create": &graphql.Field{
					Type:        patientType,
					Description: "Creates a new patient",
					Args: graphql.FieldConfigArgument{
						"name": &graphql.ArgumentConfig{
							Type: graphql.NewNonNull(graphql.String),
						},
						"email": &graphql.ArgumentConfig{
							Type: graphql.NewNonNull(graphql.String),
						},
						"phone": &graphql.ArgumentConfig{
							Type: graphql.NewNonNull(graphql.String),
						},
					},
					Resolve: func(params graphql.ResolveParams) (interface{}, error) {
						patient := Patient{}

						name, _ := params.Args["name"].(string)
						email, _ := params.Args["email"].(string)
						phone, _ := params.Args["phone"].(string)

						fmt.Println(name, email, phone)

						stmt := "insert into patients(name, email, phone) values($1, $2, $3) returning id;"
						var lastInsertID int

						err := db.QueryRow(stmt, name, email, phone).Scan(&lastInsertID)

						logFatal(err)

						patient.ID = lastInsertID
						patient.Name = name
						patient.Email = email
						patient.Phone = phone

						return patient, nil
					},
				},
				"update": &graphql.Field{
					Type:        patientType,
					Description: "Updates an existing patient.",
					Args: graphql.FieldConfigArgument{
						"id": &graphql.ArgumentConfig{
							Type: graphql.NewNonNull(graphql.Int),
						},
						"name": &graphql.ArgumentConfig{
							Type: graphql.String,
						},
						"email": &graphql.ArgumentConfig{
							Type: graphql.NewNonNull(graphql.String),
						},
						"phone": &graphql.ArgumentConfig{
							Type: graphql.NewNonNull(graphql.String),
						},
					},
					Resolve: func(params graphql.ResolveParams) (interface{}, error) {
						patient := Patient{}
						id, _ := params.Args["id"].(int)
						name, _ := params.Args["name"].(string)
						email, _ := params.Args["email"].(string)
						phone, _ := params.Args["phone"].(string)

						stmt, err := db.Prepare("update patients set name = $1, email = $2, phone = $3 where id =$4")
						logFatal(err)

						_, err = stmt.Exec(name, email, phone, id)
						logFatal(err)

						patient.ID = id
						patient.Name = name
						patient.Email = email
						patient.Phone = phone

						return patient, nil
					},
				},
				"delete": &graphql.Field{
					Type:        patientType,
					Description: "Delete a patient by id",
					Args: graphql.FieldConfigArgument{
						"id": &graphql.ArgumentConfig{
							Type: graphql.NewNonNull(graphql.Int),
						},
					},
					Resolve: func(params graphql.ResolveParams) (interface{}, error) {
						id, _ := params.Args["id"].(int)

						stmt, err := db.Prepare("delete from patients where id = $1")
						logFatal(err)

						_, err = stmt.Exec(id)
						logFatal(err)

						return nil, nil
					},
				},
			},
		},
	)

	//step 4, a schema -- an object that has the queryType and mutationType

	var schema, _ = graphql.NewSchema(
		graphql.SchemaConfig{
			Query:    queryType,
			Mutation: mutationType,
		},
	)

	//step 5, a graphql method called Do, that takes schema and a requestString and
	//returns a result..

	r := mux.NewRouter()
	r.HandleFunc("/patient", func(w http.ResponseWriter, r *http.Request) {

		result := graphql.Do(graphql.Params{
			Schema:        schema,
			RequestString: r.URL.Query().Get("query"),
		})

		json.NewEncoder(w).Encode(result)
	})

	fmt.Println("Listening on port 8000")
	http.ListenAndServe(":8000", r)
}
