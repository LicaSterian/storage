package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/LicaSterian/storage/api/model"
	"github.com/gofrs/uuid"
)

type tableFields []string

func (tf tableFields) has(field string) bool {
	for _, f := range tf {
		if field == f {
			return true
		}
	}
	return false
}

const (
	minFilesPerPage int = 10
	maxFilesPerPage int = 100

	likeOperation string = "$like"
	eqOperation   string = "$eq"
	neOperation   string = "$ne"
	ltOperation   string = "$lt"
	lteOperation  string = "$lte"
	gtOperation   string = "$gt"
	gteOperation  string = "$gte"
)

// Storage struct
type Storage struct {
	db *sql.DB
	// sanitizeRegex
}

// NewStorage constructor func receives a sql.DB pointer and returns a Storage struct
func NewStorage(db *sql.DB) Storage {
	// re :=
	// fmt.Println(re.MatchString("gopher"))
	return Storage{
		db: db,
		// sanitizeRegex: regexp.MustCompile(`^[A-Za-z0-9@-_.]+$`)
	}
}

func (s Storage) getAllFields(tableName string) (res tableFields, err error) {
	query := fmt.Sprintf("select * from %s where false;", tableName)
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	return rows.Columns()
}

// GetAll func
func (s Storage) GetAll(tableName string, req model.GetAllRequest) (res model.GetAllResponse, statusCode int, err error) {
	// paramCounter := 1
	res.RequestID = req.ID
	res.Data = model.GetAllResponseData{
		Rows: []interface{}{},
	}
	if req.Page < 1 {
		err = errors.New("page param less than 1")
		res.Error = err.Error()
		statusCode = http.StatusBadRequest
		return
	}

	if req.PerPage < minFilesPerPage {
		err = fmt.Errorf("perPage param less than %d", minFilesPerPage)
		res.Error = err.Error()
		statusCode = http.StatusBadRequest
		return
	}
	if req.PerPage > maxFilesPerPage {
		err = fmt.Errorf("perPage param greater than %d", maxFilesPerPage)
		res.Error = err.Error()
		statusCode = http.StatusBadRequest
		return
	}

	fields, err := s.getAllFields(tableName)
	if err != nil {
		err = fmt.Errorf("storage.getAllFields error, does table '%s' exist?", tableName)
		res.Error = err.Error()
		statusCode = http.StatusBadRequest
		return
	}
	// sanitizing req.filterFields
	for _, filterField := range req.FilterFields {
		if !fields.has(filterField) {
			err = fmt.Errorf("filterFields error: table '%s' does not have field: %s", tableName, filterField)
			res.Error = err.Error()
			statusCode = http.StatusBadRequest
			return
		}
	}
	// fmt.Println("fields:", fields)
	// sanitize req.sortBy
	if !fields.has(req.SortBy) {
		err = fmt.Errorf("sortBy error: table '%s' does not have field: %s", tableName, req.SortBy)
		res.Error = err.Error()
		statusCode = http.StatusBadRequest
		return
	}

	// sanitize req.fields
	for _, field := range req.Fields {
		if !fields.has(field) {
			err = fmt.Errorf("fields error: table '%s' does not have field: %s", tableName, field)
			res.Error = err.Error()
			statusCode = http.StatusBadRequest
			return
		}
	}
	// queries
	queryCounter := 1
	queryArgs := []interface{}{}
	var totalRows int64
	countStr := fmt.Sprintf("SELECT COUNT(*) FROM %s ", tableName)
	queryStr := "SELECT "
	if len(req.Fields) == 0 {
		queryStr += "*"
	} else {
		if len(req.Fields) > len(fields) {
			err = fmt.Errorf("number of fields greater than table's '%s' actual fields", tableName)
			res.Error = err.Error()
			statusCode = http.StatusBadRequest
			return
		}
		queryStr += strings.Join(req.Fields, ", ")
	}
	queryStr += fmt.Sprintf(" FROM %s ", tableName)
	if len(req.FilterFields) > 0 {
		if len(req.FilterFields) != len(req.FilterValues) {
			err = errors.New("filterFields length not the same as filterValues length")
			res.Error = err.Error()
			statusCode = http.StatusBadRequest
			return
		}
		if len(req.FilterFields) > len(fields) {
			err = fmt.Errorf("number of filterFields greater than table's '%s' actual fields", tableName)
			res.Error = err.Error()
			statusCode = http.StatusBadRequest
			return
		}
		// where
		for i, filterValue := range req.FilterValues {
			filterValueLen := len(filterValue)
			filterField := req.FilterFields[i]
			if filterValueLen != 2 {
				err = fmt.Errorf("filterField's %s value length not equal to 2", req.FilterFields[i])
				res.Error = err.Error()
				statusCode = http.StatusBadRequest
				return
			}

			operation, okOperation := filterValue[0].(string)
			if !okOperation {
				err = errors.New("operation type not string")
				res.Error = err.Error()
				statusCode = http.StatusBadRequest
				return
			}
			val := filterValue[1]
			if i == 0 {
				countStr += "WHERE "
				queryStr += "WHERE "
			} else {
				countStr += "AND "
				queryStr += "AND "
			}
			switch operation {
			case likeOperation:
				if value, valueOk := val.(string); valueOk {
					countStr += fmt.Sprintf("%s LIKE $%d ", filterField, queryCounter)
					queryStr += fmt.Sprintf("%s LIKE $%d ", filterField, queryCounter)
					queryArgs = append(queryArgs, fmt.Sprintf("%%%s%%", value))
				} else {
					err = fmt.Errorf("value in type %s not string", likeOperation)
					res.Error = err.Error()
					statusCode = http.StatusBadRequest
					return
				}
			case eqOperation:
				countStr += fmt.Sprintf("%s = $%d ", filterField, queryCounter)
				queryStr += fmt.Sprintf("%s = $%d ", filterField, queryCounter)
				queryArgs = append(queryArgs, val)
			case neOperation:
				countStr += fmt.Sprintf("%s <> $%d ", filterField, queryCounter)
				queryStr += fmt.Sprintf("%s <> $%d ", filterField, queryCounter)
				queryArgs = append(queryArgs, val)
			case ltOperation, lteOperation, gtOperation, gteOperation:
				var operationSymbol string
				switch operation {
				case ltOperation:
					operationSymbol = "<"
				case lteOperation:
					operationSymbol = "<="
				case gtOperation:
					operationSymbol = ">"
				case gteOperation:
					operationSymbol = ">="
				}
				countStr += fmt.Sprintf("%s %s $%d ", filterField, operationSymbol, queryCounter)
				queryStr += fmt.Sprintf("%s %s $%d ", filterField, operationSymbol, queryCounter)
				queryArgs = append(queryArgs, val)
			default:
				err = fmt.Errorf("invalid operation %s", operation)
				res.Error = err.Error()
				statusCode = http.StatusBadRequest
				return
			}
			queryCounter++
		}
	}
	countStr += ";"
	// fmt.Println("count str", countStr)
	// fmt.Println("query args", queryArgs)

	sqlTotalRow := s.db.QueryRow(countStr, queryArgs...)
	totalErr := sqlTotalRow.Scan(&totalRows)
	if totalErr != nil {
		err = errors.New("db queryRow total error")
		res.Error = err.Error()
		statusCode = http.StatusInternalServerError
		return
	}

	// order by
	if req.SortBy != "" {
		order := "DESC"
		if req.SortAsc {
			order = "ASC"
		}
		queryStr += fmt.Sprintf("ORDER BY $%d %s ", queryCounter, order)
		queryArgs = append(queryArgs, req.SortBy)
		queryCounter++
	}
	// query pagination
	offset := (req.Page - 1) * req.PerPage
	queryStr += fmt.Sprintf("LIMIT $%d OFFSET $%d;", queryCounter, queryCounter+1)
	queryArgs = append(queryArgs, req.PerPage, offset)
	// fmt.Println("query str", queryStr)
	// fmt.Println("query args", queryArgs)

	// fmt.Printf("req %+v\n", req)
	sqlRows, err := s.db.Query(queryStr, queryArgs...)
	if err != nil {
		err = errors.New("db query error")
		res.Error = err.Error()
		statusCode = http.StatusInternalServerError
		return
	}
	rows := make([]interface{}, 0)
	cols, colsErr := sqlRows.Columns()
	if colsErr != nil {
		err = errors.New("sqlRows.Columns() error")
		res.Error = err.Error()
		statusCode = http.StatusInternalServerError
	}
	colsTypes, colsTypesErr := sqlRows.ColumnTypes()
	if colsTypesErr != nil {
		err = errors.New("sqlRows.ColumnTypes() error")
		res.Error = err.Error()
		statusCode = http.StatusInternalServerError
	}

	for sqlRows.Next() {
		vals := make([]interface{}, len(cols))
		valsMap := make(map[string]interface{})
		for i, colType := range colsTypes {
			var newType interface{}
			switch colType.DatabaseTypeName() {
			case "UUID":
				newType = new(uuid.UUID)
			case "VARCHAR", "TEXT":
				newType = new(string)
			case "INT8", "INT16", "INT32", "INT64":
				newType = new(int64)
			case "DECIMAL":
				newType = new(float64)
			}
			vals[i] = newType
		}
		if scanErr := sqlRows.Scan(vals...); scanErr != nil {
			log.Println("sqlRows scan error:", scanErr)
			err = errors.New("sqlRows scan error")
			res.Error = err.Error()
			statusCode = http.StatusInternalServerError
			return
		}
		for i, col := range cols {
			valsMap[col] = vals[i]
		}
		rows = append(rows, valsMap)
	}
	rowsCloseErr := sqlRows.Close()
	if rowsCloseErr != nil {
		err = errors.New("sqlRows.Close() error")
		res.Error = err.Error()
		statusCode = http.StatusInternalServerError
		return
	}
	if rowsErr := sqlRows.Err(); rowsErr != nil {
		err = errors.New("sqlRows.Err() error")
		res.Error = err.Error()
		statusCode = http.StatusInternalServerError
		return
	}

	res.Success = true
	res.Data.Total = totalRows
	res.Data.Rows = rows

	return
}
