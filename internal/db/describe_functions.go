package db

import (
	"fmt"
	"strings"

	"github.com/axonops/cqlai/internal/session"
)

// DescribeFunctionsQuery executes the query to list all functions in the current keyspace (for pre-4.0)
func (s *Session) DescribeFunctionsQuery(currentKeyspace string) ([][]string, error) {
	if currentKeyspace == "" {
		return nil, fmt.Errorf("no keyspace selected")
	}

	query := `SELECT function_name, argument_types, return_type 
	          FROM system_schema.functions 
	          WHERE keyspace_name = ?`

	iter := s.Query(query, currentKeyspace).Iter()

	results := [][]string{{"Function", "Arguments", "Return Type"}}
	var functionName, returnType string
	var argumentTypes []string

	for iter.Scan(&functionName, &argumentTypes, &returnType) {
		args := strings.Join(argumentTypes, ", ")
		results = append(results, []string{functionName, args, returnType})
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}

	return results, nil
}

// FunctionDetails holds the details of a function for describe operations
type FunctionDetails struct {
	Name          string
	ArgumentTypes []string
	ArgumentNames []string
	ReturnType    string
	Language      string
	Body          string
	CalledOnNull  bool
}

// DescribeFunctionQuery executes the query to get detailed information about a specific function
func (s *Session) DescribeFunctionQuery(currentKeyspace string, functionName string) ([]FunctionDetails, error) {
	if currentKeyspace == "" {
		return nil, fmt.Errorf("no keyspace selected")
	}

	// Functions can be overloaded, so we might get multiple results
	query := `SELECT function_name, argument_types, argument_names, return_type, 
	                language, body, called_on_null_input
	          FROM system_schema.functions 
	          WHERE keyspace_name = ? AND function_name = ?`

	iter := s.Query(query, currentKeyspace, functionName).Iter()

	var functions []FunctionDetails
	var name, returnType, language, body string
	var argumentTypes, argumentNames []string
	var calledOnNull bool

	for iter.Scan(&name, &argumentTypes, &argumentNames, &returnType, &language, &body, &calledOnNull) {
		functions = append(functions, FunctionDetails{
			Name:          name,
			ArgumentTypes: argumentTypes,
			ArgumentNames: argumentNames,
			ReturnType:    returnType,
			Language:      language,
			Body:          body,
			CalledOnNull:  calledOnNull,
		})
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}

	return functions, nil
}

// DBDescribeFunctions handles version detection and returns appropriate data
func (s *Session) DBDescribeFunctions(sessionMgr *session.Manager) (interface{}, bool, error) {
	// Check if we can use server-side DESCRIBE (Cassandra 4.0+)
	if s.IsVersion4OrHigher() {
		// Use server-side DESCRIBE FUNCTIONS
		result := s.ExecuteCQLQuery("DESCRIBE FUNCTIONS")
		return result, true, nil // true means server-side was used
	}

	// Fall back to manual construction for pre-4.0
	currentKeyspace := ""
	if sessionMgr != nil {
		currentKeyspace = sessionMgr.CurrentKeyspace()
	}
	if currentKeyspace == "" {
		return nil, false, fmt.Errorf("no keyspace selected")
	}

	results, err := s.DescribeFunctionsQuery(currentKeyspace)
	if err != nil {
		return nil, false, err
	}

	return results, false, nil // false means manual query was used
}

// DBDescribeFunction handles version detection and returns appropriate data
func (s *Session) DBDescribeFunction(sessionMgr *session.Manager, functionName string) (interface{}, []FunctionDetails, error) {
	// Check if we can use server-side DESCRIBE (Cassandra 4.0+)
	if s.IsVersion4OrHigher() {
		// Parse keyspace.function or just function
		var describeCmd string
		if strings.Contains(functionName, ".") {
			describeCmd = fmt.Sprintf("DESCRIBE FUNCTION %s", functionName)
		} else {
			currentKeyspace := ""
			if sessionMgr != nil {
				currentKeyspace = sessionMgr.CurrentKeyspace()
			}
			if currentKeyspace == "" {
				return nil, nil, fmt.Errorf("no keyspace selected")
			}
			describeCmd = fmt.Sprintf("DESCRIBE FUNCTION %s.%s", currentKeyspace, functionName)
		}

		result := s.ExecuteCQLQuery(describeCmd)
		return result, nil, nil // Server-side result, no FunctionDetails needed
	}

	// Fall back to manual construction for pre-4.0
	currentKeyspace := ""
	if sessionMgr != nil {
		currentKeyspace = sessionMgr.CurrentKeyspace()
	}
	if currentKeyspace == "" {
		return nil, nil, fmt.Errorf("no keyspace selected")
	}

	functions, err := s.DescribeFunctionQuery(currentKeyspace, functionName)
	if err != nil {
		return nil, nil, err
	}

	return nil, functions, nil // Manual query result, return FunctionDetails for formatting
}

// DBDescribeFunctionByName returns raw query results for a function with optional keyspace
func (s *Session) DBDescribeFunctionByName(functionName string, keyspaceName string) (interface{}, error) {
	var query string

	if keyspaceName != "" {
		// Specific keyspace provided
		query = fmt.Sprintf("SELECT * FROM system_schema.functions WHERE keyspace_name = '%s' AND function_name = '%s'",
			keyspaceName, functionName)
	} else {
		// No keyspace - search all keyspaces
		query = fmt.Sprintf("SELECT * FROM system_schema.functions WHERE function_name = '%s'", functionName)
	}

	return s.ExecuteCQLQuery(query), nil
}
