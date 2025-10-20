package sqlx

import (
	types "rerng_addicted_api/pkg/model"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// built sql filter
func BuildSQLFilter(filters []types.Filter) (string, []interface{}) {
	var clauses []string
	var params []interface{}
	placeholder := 1

	// Get current OS time
	app_timezone := os.Getenv("APP_TIMEZONE")
	if app_timezone == "" {
		app_timezone = "Asia/Phnom_Penh"
	}
	location, err := time.LoadLocation(app_timezone)
	if err != nil {
		return "", nil
	}

	// convert value types
	convertToTime := func(v interface{}) (time.Time, bool) {
		switch val := v.(type) {
		case string:
			parsed, err := time.ParseInLocation("2006-01-02", val, location)
			if err == nil {
				return parsed, true
			}
		case time.Time:
			return val, true
		}
		return time.Time{}, false
	}

	convertToBool := func(v interface{}) (bool, bool) {
		switch val := v.(type) {
		case bool:
			return val, true
		case string:
			if val == "true" {
				return true, true
			} else if val == "false" {
				return false, true
			}
		}
		return false, false
	}

	for _, f := range filters {
		field := f.Property
		op := strings.ToLower(f.Operator)

		switch op {
		case "eq", "neq", "lt", "lte", "gt", "gte":
			sqlOp := map[string]string{
				"eq":  "=",
				"neq": "!=",
				"lt":  "<",
				"lte": "<=",
				"gt":  ">",
				"gte": ">=",
			}[op]

			// try to convert to time
			if t, ok := convertToTime(f.Value); ok {
				f.Value = t
			} else if b, ok := convertToBool(f.Value); ok {
				f.Value = b
			}

			clauses = append(clauses, fmt.Sprintf("%s %s $%d", field, sqlOp, placeholder))
			params = append(params, f.Value)
			placeholder++

		case "like":
			clauses = append(clauses, fmt.Sprintf("%s LIKE $%d", field, placeholder))
			params = append(params, f.Value)
			placeholder++

		case "in":
			vals, ok := f.Value.([]interface{})
			if !ok || len(vals) == 0 {
				continue
			}
			var ph []string
			for _, v := range vals {
				ph = append(ph, fmt.Sprintf("$%d", placeholder))
				params = append(params, v)
				placeholder++
			}
			clauses = append(clauses, fmt.Sprintf("%s IN (%s)", field, strings.Join(ph, ", ")))

		case "between":
			vals, ok := f.Value.([]interface{})
			if !ok || len(vals) != 2 {
				continue
			}

			start, ok1 := convertToTime(vals[0])
			end, ok2 := convertToTime(vals[1])

			if ok1 && ok2 {
				// Make end date inclusive to the end of the day
				end = end.Add(24 * time.Hour).Add(-time.Second)
				clauses = append(clauses, fmt.Sprintf("%s BETWEEN $%d AND $%d", field, placeholder, placeholder+1))
				params = append(params, start, end)
				placeholder += 2
			} else {
				// Fallback: use original values
				clauses = append(clauses, fmt.Sprintf("%s BETWEEN $%d AND $%d", field, placeholder, placeholder+1))
				params = append(params, vals[0], vals[1])
				placeholder += 2
			}
		}
	}

	if len(clauses) == 0 {
		return "", nil
	}
	return strings.Join(clauses, " AND "), params
}

// built sql sort
func BuildSQLSort(sorts []types.Sort) string {
	var orderClauses []string

	for _, sort := range sorts {
		field := sort.Property
		direction := strings.ToUpper(sort.Direction)

		// ensure the direction is either ASC or DESC
		if direction != "ASC" && direction != "DESC" {
			direction = "ASC"
		}

		// add the sort clause to the list of order clauses
		orderClauses = append(orderClauses, fmt.Sprintf("%s %s", field, direction))
	}

	if len(orderClauses) == 0 {
		return ""
	}

	// join the clauses with commas and return the final order by string
	return "ORDER BY " + strings.Join(orderClauses, ", ")
}

// built sql paging
func BuildPaging(page int, perPage int) string {
	// var params []interface{}

	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}

	offset := (page - 1) * perPage
	limit := perPage

	// params = append(params, offset, limit)

	return fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset)
}

func GetIdByUuid(space_name string, uuid_field_name string, uuid_str string, db *sqlx.Tx) (*int, error) {
	var id int

	// Parse the UUID
	uid, err := uuid.Parse(uuid_str)
	if err != nil {
		return nil, err
	}

	// Define the SQL query
	sql := fmt.Sprintf(`SELECT id FROM %s WHERE %s=$1`, space_name, uuid_field_name)

	// Execute the query
	err = db.Get(&id, sql, uid)
	if err != nil {
		return nil, err
	}

	return &id, nil
}

// GetSeqNextVal returns the next value from a sequence
// SeqResult struct to store sequence result
type SeqResult struct {
	ID int `db:"id"`
}

// Supports both normal DB connection and transactions
func GetSeqNextVal(seqName string, exec sqlx.Ext) (*int, error) {
	var result SeqResult
	sql := `SELECT nextval($1) AS id`

	// Execute query using either DB or transaction
	err := sqlx.Get(exec, &result, sql, seqName)
	if err != nil {
		return nil, fmt.Errorf("failed to get sequence value: %w", err)
	}
	return &result.ID, nil
}

// SetSeqNextVal sets and returns the next sequence value
func SetSeqNextVal(seq_name string, db *sqlx.Tx) (*int, error) {
	var id int

	// Define the SQL query - adjust to PostgreSQL sequence operations
	sql := fmt.Sprintf(`SELECT nextval('%s') as id`, seq_name)

	// Execute the query
	err := db.Get(&id, sql)
	if err != nil {
		return nil, err
	}

	return &id, nil
}

// IsExists checks if a record exists with the given field value
func IsExists(space_name string, field_name string, value interface{}, db *sqlx.Tx) (bool, error) {
	var exists bool

	// Define the SQL query
	sql := fmt.Sprintf(`SELECT EXISTS(SELECT 1 FROM %s WHERE %s=$1 AND deleted_at IS NULL)`, space_name, field_name)

	// Execute the query
	err := db.Get(&exists, sql, value)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// IsExistsWhere checks if a record exists with custom WHERE conditions
func IsExistsWhere(space_name string, where_sqlstr string, args []interface{}, db *sqlx.Tx) (bool, error) {
	var exists bool

	// Define the SQL query
	sql := fmt.Sprintf(`SELECT EXISTS(SELECT 1 FROM %s WHERE %s AND deleted_at IS NULL)`, space_name, where_sqlstr)

	// Execute the query
	err := db.Get(&exists, sql, args...)
	if err != nil {
		return false, err
	}

	return exists, nil
}
