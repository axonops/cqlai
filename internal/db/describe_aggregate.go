package db

import (
	"fmt"
	"strings"
)

// AggregateInfo holds aggregate information for manual describe
type AggregateInfo struct {
	Name          string
	ArgumentTypes []string
	StateFunc     string
	StateType     string
	FinalFunc     string
	InitCond      string
	ReturnType    string
}

// AggregateListInfo holds aggregate list information for manual describe
type AggregateListInfo struct {
	Name          string
	ArgumentTypes []string
	StateType     string
	ReturnType    string
}

// DescribeAggregatesQuery executes the query to list all aggregates (for pre-4.0)
func (s *Session) DescribeAggregatesQuery(keyspace string) ([]AggregateListInfo, error) {
	query := `SELECT aggregate_name, argument_types, state_type, return_type 
	          FROM system_schema.aggregates 
	          WHERE keyspace_name = ?`

	iter := s.Query(query, keyspace).Iter()

	var aggregates []AggregateListInfo
	var aggregateName, stateType, returnType string
	var argumentTypes []string

	for iter.Scan(&aggregateName, &argumentTypes, &stateType, &returnType) {
		aggregates = append(aggregates, AggregateListInfo{
			Name:          aggregateName,
			ArgumentTypes: argumentTypes,
			StateType:     stateType,
			ReturnType:    returnType,
		})
	}
	
	if err := iter.Close(); err != nil {
		return nil, err
	}

	return aggregates, nil
}

// DescribeAggregateQuery executes the query to get aggregate information (for pre-4.0)
func (s *Session) DescribeAggregateQuery(keyspace string, aggregateName string) (*AggregateInfo, error) {
	query := `SELECT aggregate_name, argument_types, state_func, state_type, 
	                final_func, initcond, return_type
	          FROM system_schema.aggregates 
	          WHERE keyspace_name = ? AND aggregate_name = ?`

	iter := s.Query(query, keyspace, aggregateName).Iter()

	var name, stateFunc, stateType, finalFunc, initCond, returnType string
	var argumentTypes []string

	if !iter.Scan(&name, &argumentTypes, &stateFunc, &stateType, &finalFunc, &initCond, &returnType) {
		iter.Close()
		return nil, fmt.Errorf("aggregate '%s' not found in keyspace '%s'", aggregateName, keyspace)
	}
	iter.Close()

	return &AggregateInfo{
		Name:          name,
		ArgumentTypes: argumentTypes,
		StateFunc:     stateFunc,
		StateType:     stateType,
		FinalFunc:     finalFunc,
		InitCond:      initCond,
		ReturnType:    returnType,
	}, nil
}

// DBDescribeAggregates handles version detection and returns appropriate data
func (s *Session) DBDescribeAggregates() (interface{}, []AggregateListInfo, error) {
	// Check if we can use server-side DESCRIBE (Cassandra 4.0+)
	if s.IsVersion4OrHigher() {
		// Use server-side DESCRIBE AGGREGATES
		result := s.ExecuteCQLQuery("DESCRIBE AGGREGATES")
		return result, nil, nil // Server-side result, no AggregateListInfo needed
	}
	
	// Fall back to manual construction for pre-4.0
	currentKeyspace := s.CurrentKeyspace()
	if currentKeyspace == "" {
		return nil, nil, fmt.Errorf("no keyspace selected")
	}

	aggregates, err := s.DescribeAggregatesQuery(currentKeyspace)
	if err != nil {
		return nil, nil, err
	}

	return nil, aggregates, nil // Manual query result, return AggregateListInfo for formatting
}

// DBDescribeAggregate handles version detection and returns appropriate data
func (s *Session) DBDescribeAggregate(aggregateName string) (interface{}, *AggregateInfo, error) {
	// Check if we can use server-side DESCRIBE (Cassandra 4.0+)
	if s.IsVersion4OrHigher() {
		// Parse keyspace.aggregate or just aggregate
		var describeCmd string
		if strings.Contains(aggregateName, ".") {
			describeCmd = fmt.Sprintf("DESCRIBE AGGREGATE %s", aggregateName)
		} else {
			currentKeyspace := s.CurrentKeyspace()
			if currentKeyspace == "" {
				return nil, nil, fmt.Errorf("no keyspace selected")
			}
			describeCmd = fmt.Sprintf("DESCRIBE AGGREGATE %s.%s", currentKeyspace, aggregateName)
		}
		
		result := s.ExecuteCQLQuery(describeCmd)
		return result, nil, nil // Server-side result, no AggregateInfo needed
	}
	
	// Fall back to manual construction for pre-4.0
	currentKeyspace := s.CurrentKeyspace()
	if currentKeyspace == "" {
		return nil, nil, fmt.Errorf("no keyspace selected")
	}

	aggregateInfo, err := s.DescribeAggregateQuery(currentKeyspace, aggregateName)
	if err != nil {
		return nil, nil, err
	}

	return nil, aggregateInfo, nil // Manual query result, return AggregateInfo for formatting
}