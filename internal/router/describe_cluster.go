package router

import (
	"fmt"
	"strings"
)

// describeCluster shows cluster information
func (v *CqlCommandVisitorImpl) describeCluster() interface{} {
	serverResult, clusterInfo, err := v.session.DBDescribeCluster()
	
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}
	
	if serverResult != nil {
		// Server-side DESCRIBE result, return as-is
		return serverResult
	}
	
	// Manual query result - format it
	if clusterInfo == nil {
		return "Could not retrieve cluster information."
	}
	
	// Format as a pretty table consistent with other tables
	var result strings.Builder
	
	// Shorten partitioner name
	partitionerShort := clusterInfo.Partitioner
	if strings.Contains(clusterInfo.Partitioner, ".") {
		parts := strings.Split(clusterInfo.Partitioner, ".")
		partitionerShort = parts[len(parts)-1]
	}
	
	// Calculate max width needed
	maxLabelWidth := 17 // "Cassandra Version" is longest
	maxValueWidth := len(clusterInfo.ClusterName)
	if len(partitionerShort) > maxValueWidth {
		maxValueWidth = len(partitionerShort)
	}
	if len(clusterInfo.Version) > maxValueWidth {
		maxValueWidth = len(clusterInfo.Version)
	}
	
	// Ensure minimum widths
	if maxValueWidth < 30 {
		maxValueWidth = 30
	}
	
	totalWidth := maxLabelWidth + 3 + maxValueWidth + 2 // label + " : " + value + spaces
	
	// Draw top border
	result.WriteString("┌")
	result.WriteString(strings.Repeat("─", totalWidth))
	result.WriteString("┐\n")
	
	// Title
	title := "CLUSTER INFORMATION"
	titlePadding := (totalWidth - len(title)) / 2
	result.WriteString("│")
	result.WriteString(strings.Repeat(" ", titlePadding))
	result.WriteString(title)
	result.WriteString(strings.Repeat(" ", totalWidth-titlePadding-len(title)))
	result.WriteString("│\n")
	
	// Separator
	result.WriteString("├")
	result.WriteString(strings.Repeat("─", totalWidth))
	result.WriteString("┤\n")
	
	// Data rows
	result.WriteString(fmt.Sprintf("│ %-*s : %-*s │\n",
		maxLabelWidth, "Cluster Name", maxValueWidth, clusterInfo.ClusterName))
	result.WriteString(fmt.Sprintf("│ %-*s : %-*s │\n",
		maxLabelWidth, "Partitioner", maxValueWidth, partitionerShort))
	result.WriteString(fmt.Sprintf("│ %-*s : %-*s │\n",
		maxLabelWidth, "Cassandra Version", maxValueWidth, clusterInfo.Version))
	
	// Bottom border
	result.WriteString("└")
	result.WriteString(strings.Repeat("─", totalWidth))
	result.WriteString("┘")
	
	return result.String()
}
