package db

import (
	"context"
	"database/sql"
	"errors"
	"github.com/joho/godotenv"
	"log"
	"os"
)

//import (
//	"context"
//	"database/sql"
//	"errors"
//	"fmt"
//	"generic-db-lib/db"
//	"github.com/joho/godotenv"
//	"log"
//	"os"
//	"regexp"
//	"strings"
//	"task-management/server/commons"
//)

//import (
//"context"
//"database/sql"
//"errors"
//"fmt"
//"log"
//"regexp"
//"strings"
//generatedModel "task-managament/server/api/graph/model"
//"task-managament/server/commons"
//dbModel "task-managament/server/models"
//)

//var saveQuery = fmt.Sprintf("INSERT INTO %s (%s) VALUES ($1,$2,$3,$4,$5) RETURNING %s;", repositoryTableName, insertFields, returnFields)
//var getQuery = fmt.Sprintf("SELECT %s FROM %s WHERE service_tag = $1;", returnFields, repositoryTableName)

type databaseRepository struct {
	con *sql.DB
}

//func GMin[T constraints.Ordered](x, y T, error) T {
//	if x < y {
//		return x
//	}
//	return y
//}

const (
	repositoryTableName = "task"
	insertFields        = " id, content, title, views, timestamp "
	returnFields        = insertFields
)

func NewRepository() (*databaseRepository, error) {
	err := godotenv.Load(".env")
	if err == nil {

	}
	ctx := context.Background()
	config := &Config{}
	*config = parseEnv()
	connection, err := NewDatabase(*config)
	if err != nil {
		log.Print(ctx, err, "unable to setup database")
		return nil, errors.New("unable to setup database")
	}

	err = Migrate(connection, config)
	if err != nil {
		log.Print(ctx, err, "unable to setup database")
		return nil, errors.New("failed during migration")
	}
	return &databaseRepository{connection}, nil
}

//func NewRepository(config Config) (repositories.Repository, error) {
//	con, err := NewDatabase(config)
//	if err != nil {
//		return nil, errors.New("unable to setup database")
//	}
//
//	err = Migrate(con)
//	if err != nil {
//		return nil, errors.Wrapf(err, "Failed while migration")
//	}
//	return &databaseRepository{con}, nil
//}
//
//func (repo *databaseRepository) Save(id string, content string, title string, views int, timestamp string) (*dbModel.Task, error) {
//	taskInput := dbModel.Task{}
//	err := repo.con.QueryRow(saveQuery, id, content, title, views, timestamp).Scan(&taskInput.ID,
//		&taskInput.Content, &taskInput.Title, &taskInput.Views, &taskInput.Timestamp)
//	if err == nil {
//		return &taskInput, nil
//	}
//	return nil, err
//}
//
//func (repo *databaseRepository) Update(ctx context.Context, input generatedModel.TaskInput) (model *generatedModel.Task, err error) {
//	//TODO: update below print statement with database persist functionality
//	return nil, err
//}
//
//func (repo *databaseRepository) GetList() (model []dbModel.Task, _ error) {
//	baseQuery := fmt.Sprintf("SELECT %s FROM %s ", returnFields, repositoryTableName)
//
//	var rows *sql.Rows
//	var err error = nil
//	listQuery := fmt.Sprintf(baseQuery)
//	rows, err = repo.con.Query(listQuery)
//	if err != nil {
//		return []dbModel.Task{}, err
//	}
//
//	var list []dbModel.Task
//	for rows.Next() {
//		var taskItem dbModel.Task
//		err := rows.Scan(&taskItem.ID, &taskItem.Title, &taskItem.Content, &taskItem.Views, &taskItem.Timestamp)
//		if err != nil {
//			return []dbModel.Task{}, err
//		}
//		list = append(list, taskItem)
//	}
//	if err = rows.Err(); err != nil {
//		return []dbModel.Task{}, err
//	}
//	return list, nil
//}
//
//func (repo *databaseRepository) GetById(ctx context.Context, id int) (model *dbModel.Task, err error) {
//	//TODO: update below print statement with database persist functionality
//	return nil, err
//}
//
//func (repo *databaseRepository) GetByQuery(queryInputMap map[string]string) ([]dbModel.Task, error) {
//
//	baseQuery := ""
//	logicalOperator := queryInputMap["logical_operator"]
//
//	var queryFields [2]string
//	var queryValues [2]string
//	var comparisonValues [2]string
//
//	if logicalOperator != "" {
//
//		comparisonoperator1 := queryInputMap["comparison_operator_1"]
//		comparisonoperator2 := queryInputMap["comparison_operator_2"]
//
//		queryFields[0] = strings.Trim(strings.Trim(regexp.MustCompile(`\((.*?)\,+`).FindString(comparisonoperator1), "("), ",")
//		queryValues[0] = strings.Trim(strings.Trim(regexp.MustCompile(`\,(.*?)\)+`).FindString(comparisonoperator1), ","), ")")
//		queryFields[1] = strings.Trim(strings.Trim(regexp.MustCompile(`\((.*?)\,+`).FindString(comparisonoperator2), "("), ",")
//		queryValues[1] = strings.Trim(strings.Trim(regexp.MustCompile(`\,(.*?)\)+`).FindString(comparisonoperator2), ","), ")")
//
//		firstOperator := regexp.MustCompile(`^[^\(]+`).FindString(comparisonoperator1)
//		secondOperator := regexp.MustCompile(`^[^\(]+`).FindString(comparisonoperator2)
//
//		switch firstOperator {
//
//		case "EQUAL":
//			comparisonValues[0] = "="
//		case "GREATER_THAN":
//			comparisonValues[0] = ">"
//		case "LESS_THAN":
//			comparisonValues[0] = "<"
//		}
//
//		switch secondOperator {
//
//		case "EQUAL":
//			comparisonValues[1] = "="
//		case "GREATER_THAN":
//			comparisonValues[1] = ">"
//		case "LESS_THAN":
//			comparisonValues[1] = "<"
//		}
//
//		baseQuery = fmt.Sprintf("SELECT %s FROM %s WHERE %s%s $1 %s %s%s $2;", returnFields, repositoryTableName, queryFields[0], comparisonValues[0], logicalOperator, queryFields[1], comparisonValues[1])
//	} else {
//		baseQuery = fmt.Sprintf("SELECT %s FROM %s WHERE %s%s $1", returnFields, repositoryTableName, queryFields[0], comparisonValues[0])
//
//	}
//
//	var rows *sql.Rows
//	var err error = nil
//
//	listQuery := fmt.Sprintf(baseQuery)
//
//	if len(queryValues) == 2 {
//		rows, err = repo.con.Query(listQuery, queryValues[0], queryValues[1])
//	} else {
//		rows, err = repo.con.Query(listQuery, queryValues[0])
//	}
//
//	if err != nil {
//		return []dbModel.Task{}, err
//	}
//
//	log.Print(rows)
//
//	var list []dbModel.Task
//	for rows.Next() {
//		var taskItem dbModel.Task
//		err := rows.Scan(&taskItem.ID, &taskItem.Title, &taskItem.Content, &taskItem.Views, &taskItem.Timestamp)
//		if err != nil {
//			return []dbModel.Task{}, err
//		}
//		list = append(list, taskItem)
//	}
//	if err = rows.Err(); err != nil {
//		return []dbModel.Task{}, err
//	}
//	return list, nil
//}
//

func parseEnv() Config {
	dbHost := os.Getenv("DB-HOST")
	dbPort := os.Getenv("DB-PORT")
	dbUser := os.Getenv("DB-USER")
	dbPassword := os.Getenv("DB-PASSWORD")
	database := os.Getenv("DATABASE")
	if dbHost == "" && dbPort == "" && dbUser == "" && dbPassword == "" && database == "" {
		//error - missing env vars
	}
	return Config{
		Host:     dbHost,
		Port:     dbPort,
		User:     dbUser,
		Password: dbPassword,
		Database: database,
	}
}
