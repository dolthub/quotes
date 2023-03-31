package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/go-sql-driver/mysql"
)

const (
	PasswordParameterName = "/amplify/d1bei8pu57dwsm/dev/AMPLIFY_QuotesAPIFunc_DOLT_PASS"
)

type params struct {
	Db       string
	Host     string
	Username string
	Password string
}

func mustGetEnvVar(name string) string {
	value := os.Getenv(name)
	if value == "" {
		panic("missing required environment variable: " + name)
	}
	return value
}

func readParams() params {
	sm := ssm.New(session.Must(session.NewSessionWithOptions(
		session.Options{
			SharedConfigState: session.SharedConfigEnable,
		},
	)))

	result, err := sm.GetParameter(&ssm.GetParameterInput{
		Name:           aws.String(PasswordParameterName),
		WithDecryption: aws.Bool(true),
	})

	if err != nil {
		panic("unable to get the dolt password: " + err.Error())
	}

	return params{
		Db:       mustGetEnvVar("DOLT_DB"),
		Host:     mustGetEnvVar("DOLT_HOST"),
		Username: mustGetEnvVar("DOLT_USER"),
		Password: *result.Parameter.Value,
	}
}

func connectToDolt() (*sql.DB, error) {
	p := readParams()

	cfg := mysql.NewConfig()
	cfg.User = p.Username
	cfg.Passwd = p.Password
	cfg.Addr = p.Host + ":3306"
	cfg.DBName = p.Db
	cfg.ParseTime = true
	cfg.TLSConfig = "skip-verify"

	log.Printf(`creating db connection - host: "%s", db: "%s", user: "%s"`, cfg.Addr, cfg.DBName, cfg.User)

	connector, err := mysql.NewConnector(cfg)
	if err != nil {
		log.Println("failed to create mysql connector: %s", err.Error())
		return nil, err
	}

	return sql.OpenDB(connector), nil
}

var db *sql.DB
var quoteCount int

func init() {
	var err error
	db, err = connectToDolt()
	if err != nil {
		panic("Failed to connect to our database: " + err.Error())
	}

	err = db.QueryRow("SELECT count(*) FROM quotes;").Scan(&quoteCount)
	if err != nil {
		panic("Failed to count the quotes: " + err.Error())
	}

	log.Printf("Connected Successfully. There are %d quotes in our database", quoteCount)
}

type response struct {
	Author string `json:"author"`
	Quote  string `json:"quote"`
}

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	const query = "SELECT author, quote from quotes where id = ?;"

	rId := rng.Int63()%int64(quoteCount) + 1

	var author string
	var quote string
	err := db.QueryRow(query, rId).Scan(&author, &quote)
	if err != nil {
		return events.APIGatewayProxyResponse{}, fmt.Errorf("failed to retrieve quote with id '%d': %w", rId, err)
	}

	responseJson, err := json.Marshal(&response{Quote: quote, Author: author})
	if err != nil {
		return events.APIGatewayProxyResponse{}, fmt.Errorf("failed to marshal response: %w", err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    map[string]string{"Access-Control-Allow-Origin": "*"},
		Body:       string(responseJson),
	}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
