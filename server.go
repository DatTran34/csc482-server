package main

import (
	"csc482/types"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/gorilla/mux"
)

func listAllTeams() (*int64, []types.Table) {
	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")},
	)

	//Create DynamoDB client
	svc := dynamodb.New(sess)

	//using scan api
	params := &dynamodb.ScanInput{
		TableName: aws.String("dtran3-soccer-standings"),
	}
	result, err := svc.Scan(params)
	if err != nil {
		fmt.Println("failed to make Query API call", err)
	}
	fmt.Println(result.Count)
	count := result.Count

	obj := []types.Table{}
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &obj)
	if err != nil {
		fmt.Println("failed to unmarshal Query result items", err)
	}
	teams := obj
	return count, teams
}

func GetAllData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, teams := listAllTeams()
	json.NewEncoder(w).Encode(teams)
}

func GetStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	count, _ := listAllTeams()
	status := types.Status{Table: "dtran3-soccer-standings", RecordCount: count}
	json.NewEncoder(w).Encode(status)
}

func GetSearchData(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	pointStart_ := mux.Vars(r)["pointStart"]
	pointEnd_ := mux.Vars(r)["pointEnd"]

	proper1, err := regexp.MatchString(`^[0-9]+$`, pointStart_)
	proper2, err := regexp.MatchString(`^[0-9]+$`, pointEnd_)

	if proper1 == false || proper2 == false {
		w.WriteHeader(http.StatusBadRequest)
		badMessage := "Your input is wrong formatted. it should be search?pointStart=num1&pointEnd=num2"
		json.NewEncoder(w).Encode(badMessage)
		return
	}

	pointStart, err := strconv.Atoi(pointStart_)
	pointEnd, err := strconv.Atoi(pointEnd_)

	if err != nil {
		log.Fatal(err)
	}

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")},
	)
	if err != nil {
		fmt.Println("ERROR: ", err)
	}

	svc := dynamodb.New(sess)

	condition := expression.Between(expression.Name("points"), expression.Value(pointStart), expression.Value(pointEnd))

	expr, err := expression.NewBuilder().WithFilter(condition).Build()
	if err != nil {
		fmt.Println("ERROR: ", err)
	}

	params := &dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String("dtran3-soccer-standings"),
	}

	out, err := svc.Scan(params)

	if err != nil {
		fmt.Printf("ERROR: %v\n", err.Error())
	}

	resp := []types.Table{}
	err = dynamodbattribute.UnmarshalListOfMaps(out.Items, &resp)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
	}

	json.NewEncoder(w).Encode(resp)
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/dtran3/all", GetAllData).Methods("GET")
	router.HandleFunc("/dtran3/status", GetStatus).Methods("GET")
	router.HandleFunc("/dtran3/search", GetSearchData).Queries("pointStart", "{pointStart:.*}").Queries("pointEnd", "{pointEnd:.*}")
	log.Fatal(http.ListenAndServe(":8080", router))
}
