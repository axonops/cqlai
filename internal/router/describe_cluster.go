package router

import (
	"fmt"
	"strings"
	"github.com/axonops/cqlai/internal/logger"
)

// describeCluster shows cluster information
func (v *CqlCommandVisitorImpl) describeCluster() interface{} {
	// Check if we can use server-side DESCRIBE (Cassandra 4.0+)
	cassandraVersion := v.session.CassandraVersion()
	logger.DebugToFile("describeCluster", fmt.Sprintf("Cassandra version: %s", cassandraVersion))
	
	if isVersion4OrHigher(cassandraVersion) {
		// Use server-side DESCRIBE CLUSTER
		logger.DebugToFile("describeCluster", "Using server-side DESCRIBE CLUSTER")
		return v.session.ExecuteCQLQuery("DESCRIBE CLUSTER")
	}
	
	// Fall back to manual construction for pre-4.0
	iter := v.session.Query("SELECT cluster_name, partitioner, release_version FROM system.local").Iter()

	var clusterName, partitioner, version string
	if iter.Scan(&clusterName, &partitioner, &version) {
		iter.Close()

		// Format as a pretty table consistent with other tables
		var result strings.Builder

		// Shorten partitioner name
		partitionerShort := partitioner
		if strings.Contains(partitioner, ".") {
			parts := strings.Split(partitioner, ".")
			partitionerShort = parts[len(parts)-1]
		}

		// Calculate max width needed
		maxLabelWidth := 17 // "Cassandra Version" is longest
		maxValueWidth := len(clusterName)
		if len(partitionerShort) > maxValueWidth {
			maxValueWidth = len(partitionerShort)
		}
		if len(version) > maxValueWidth {
			maxValueWidth = len(version)
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
			maxLabelWidth, "Cluster Name", maxValueWidth, clusterName))
		result.WriteString(fmt.Sprintf("│ %-*s : %-*s │\n",
			maxLabelWidth, "Partitioner", maxValueWidth, partitionerShort))
		result.WriteString(fmt.Sprintf("│ %-*s : %-*s │\n",
			maxLabelWidth, "Cassandra Version", maxValueWidth, version))

		// Bottom border
		result.WriteString("└")
		result.WriteString(strings.Repeat("─", totalWidth))
		result.WriteString("┘")

		return result.String()
	}
	if err := iter.Close(); err != nil {
		return fmt.Errorf("error describing cluster: %v", err)
	}
	return "Could not retrieve cluster information."
}
