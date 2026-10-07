package validation

import (
	"strings"
)

// What a change will do, said before anyone runs it.
//
// The MCP server never runs a change. A model proposes one, and the user reads
// it and runs it, or not. What the user reads beside it comes from here: fixed
// sentences chosen from what the statement is, not from what the model says
// it does. A model can be wrong about its own statement, or steered; this
// cannot be.

// Facts is what is known about the table a statement changes, beyond the
// statement itself. Empty when it cannot be found.
type Facts struct {
	PartitionKey  []string
	ClusteringKey []string
}

// Implications is what a change will do, worst first, in plain sentences.
func Implications(s Statement, f Facts) []string {
	var out []string
	add := func(lines ...string) {
		for _, line := range lines {
			for _, have := range out {
				if have == line {
					line = ""
				}
			}
			if line != "" {
				out = append(out, line)
			}
		}
	}

	switch s.Command {
	case "DROP":
		add(dropped(s)...)
	case "TRUNCATE":
		add("TRUNCATE removes every row of the table on every node. It cannot be undone from CQL.",
			"Every node has to be up, or it fails part way.",
			snapshotNote)
	case "ALTER":
		add(altered(s)...)
	case "CREATE":
		add(created(s)...)
	case "INSERT":
		add("INSERT overwrites a row with the same primary key, without an error: it is an upsert.")
		if !s.Conditional {
			add("To insert only when there is no such row, add IF NOT EXISTS, at the cost of a lightweight transaction.")
		}
	case "UPDATE":
		add("UPDATE creates the row if there is none: it is an upsert.")
		add(keyCoverage(s, f, "updates")...)
	case "DELETE":
		add(deleted(s, f)...)
	case "BATCH":
		add(batched(s, f)...)
	}

	if s.Conditional {
		add("IF makes this a lightweight transaction: it goes through Paxos, about four round trips between replicas, " +
			"and it should not be mixed with plain writes to the same rows.")
	}
	if s.TTL {
		add("What is written expires after its TTL, and then takes up space as a tombstone until it is compacted away.")
	}
	if s.Kind == KindSchema && s.Command != "TRUNCATE" {
		add("A schema change: run one at a time, and let the cluster agree on it before the next. " +
			"Two at once from different places can leave nodes disagreeing about the schema.")
	}
	return out
}

const durableWritesNote = "durable_writes = false skips the commit log: writes not yet flushed are lost if a node stops."

const snapshotNote = "Cassandra takes a snapshot first only if auto_snapshot is on (the default), " +
	"and the snapshot stays on each node's disk until it is cleared."

func dropped(s Statement) []string {
	switch s.Object {
	case "KEYSPACE":
		return []string{"DROP KEYSPACE deletes the keyspace and every table, type, function and view in it, with all their data, on every node. It cannot be undone from CQL.",
			snapshotNote}
	case "TABLE":
		return []string{"DROP TABLE deletes the table and all its data on every node, with its indexes and materialized views. It cannot be undone from CQL.",
			snapshotNote}
	case "MATERIALIZED VIEW":
		return []string{"DROP MATERIALIZED VIEW deletes the view and its data. Queries that read it fail from then on."}
	case "INDEX":
		return []string{"DROP INDEX removes the index. Queries that use it fail, or need ALLOW FILTERING, from then on."}
	case "TYPE", "FUNCTION", "AGGREGATE":
		return []string{"DROP " + s.Object + " fails while anything still uses it."}
	case "TRIGGER":
		return []string{"DROP TRIGGER stops the trigger running on writes to the table."}
	}
	return nil
}

func altered(s Statement) []string {
	switch {
	case s.Object == "TABLE" && s.Action == "ADD":
		return []string{"Adding a column changes only the schema. Existing rows have no value in it until they are written."}
	case s.Object == "TABLE" && s.Action == "DROP":
		return []string{"Dropping a column makes its data unreadable at once, on every node. It cannot be undone from CQL.",
			"The data leaves the disk only as the table's SSTables compact.",
			"A column of the same name cannot be added back with a different type.",
			"To keep a copy, take a snapshot first (nodetool snapshot)."}
	case s.Object == "TABLE" && s.Action == "RENAME":
		return []string{"Only primary key columns can be renamed. Anything that uses the old name stops working."}
	case s.Object == "TABLE" && s.Action == "ALTER":
		return []string{"Changing a column's type is refused unless the types are compatible."}
	case s.Object == "KEYSPACE":
		return keyspaceOptions(s.Options)
	case s.Object == "TYPE":
		return []string{"A type's fields can be added or renamed, not removed."}
	}
	return tableOptions(s.Options)
}

func keyspaceOptions(options []string) []string {
	var out []string
	for _, o := range options {
		switch o {
		case "replication":
			out = append(out,
				"Changing replication moves no data. Run a full repair in each data centre it changes afterwards "+
					"(nodetool repair -full), or reads at QUORUM and above can miss data.",
				"Lowering the replica count leaves the extra copies on disk until nodetool cleanup runs on each node.")
		case "durable_writes":
			out = append(out, durableWritesNote)
		}
	}
	return out
}

func tableOptions(options []string) []string {
	var out []string
	for _, o := range options {
		switch o {
		case "compaction":
			out = append(out, "A new compaction strategy rewrites the table's data over time: expect extra disk reads, writes and space while it runs.")
		case "gc_grace_seconds":
			out = append(out, "gc_grace_seconds is how long a tombstone is kept. Set lower than the time between repairs, deleted data can come back.")
		case "default_time_to_live":
			out = append(out, "default_time_to_live applies to what is written from now on. Rows already written keep the TTL they have.")
		case "compression":
			out = append(out, "Compression applies to SSTables written from now on. Existing ones change only as they compact, or with nodetool upgradesstables -a.")
		}
	}
	if len(out) == 0 && len(options) > 0 {
		out = append(out, "Table options apply from now on, to the data written after this.")
	}
	return out
}

func created(s Statement) []string {
	switch s.Object {
	case "KEYSPACE":
		out := []string{"Check the replication: SimpleStrategy is for one data centre or testing; " +
			"NetworkTopologyStrategy names the replica count in each data centre."}
		for _, o := range s.Options {
			if o == "durable_writes" {
				out = append(out, durableWritesNote)
			}
		}
		return out
	case "TABLE":
		return append([]string{"A table's primary key cannot be changed later. Queries that cannot fix the partition key with = " +
			"read every node; a partition key with few values makes wide partitions."}, tableOptions(s.Options)...)
	case "INDEX":
		return []string{"Building the index reads all of the table's data on every node, which takes disk reads and time on a large table.",
			"A query on the index without the partition key asks every node."}
	case "MATERIALIZED VIEW":
		return []string{"Materialized views are experimental: they are off by default from Cassandra 4.0 (materialized_views_enabled), " +
			"and a view can fall out of step with its table with no way to repair it."}
	case "FUNCTION", "AGGREGATE":
		return []string{"A user-defined function runs code on every node, and needs user_defined_functions_enabled."}
	case "TRIGGER":
		return []string{"A trigger runs a class on the nodes for every write to the table."}
	case "TYPE":
		return []string{"A type can be added to later, but not have fields removed."}
	}
	return nil
}

func deleted(s Statement, f Facts) []string {
	out := keyCoverage(s, f, "deletes")
	if s.Columns {
		out = append(out, "Only the named columns are deleted. The rows themselves stay.")
	}
	return append(out,
		"A delete writes a tombstone. The data leaves reads at once, and leaves the disk only after gc_grace_seconds and a compaction. "+
			"Many tombstones in one partition slow its reads, and can make them fail.")
}

// keyCoverage says how much a DELETE or an UPDATE touches: one row, a range
// of rows, or a whole partition, from which key columns its WHERE fixes.
func keyCoverage(s Statement, f Facts, verb string) []string {
	fixed := map[string]bool{}
	for _, c := range s.Restricted {
		fixed[strings.ToLower(c)] = true
	}
	all := func(columns []string) bool {
		for _, c := range columns {
			if !fixed[strings.ToLower(c)] {
				return false
			}
		}
		return true
	}

	switch {
	case len(f.PartitionKey) == 0:
		if s.Command == "DELETE" {
			return []string{"The table's key could not be found: without the whole primary key in WHERE, this " +
				verb + " a whole partition or a range of rows."}
		}
		return nil
	case !all(f.PartitionKey):
		return []string{"WHERE does not fix the whole partition key (" + strings.Join(f.PartitionKey, ", ") +
			"), which Cassandra refuses: it will not run as written."}
	case s.Command == "UPDATE":
		return nil
	case len(f.ClusteringKey) == 0 || all(f.ClusteringKey):
		return []string{"This " + verb + " one row, with a row tombstone."}
	case !fixed[strings.ToLower(f.ClusteringKey[0])]:
		return []string{"This " + verb + " the whole partition: every row in it, with a partition tombstone."}
	}
	return []string{"This " + verb + " a range of rows in the partition, with a range tombstone."}
}

func batched(s Statement, f Facts) []string {
	var out []string
	switch s.BatchType {
	case "UNLOGGED":
		out = append(out, "An unlogged batch is not atomic: across partitions, some statements can apply and others not.")
	case "COUNTER":
		out = append(out, "A counter batch is not idempotent: retried after a timeout, it can count twice.")
	default:
		out = append(out, "A logged batch applies all of its statements or none, through the batch log, at extra cost. "+
			"It is not isolated: a reader can see some of it before the rest.")
	}
	out = append(out, "A batch is for keeping tables in step, not for speed: a large one across many partitions can overload its coordinator.")
	for _, inner := range s.Inner {
		// The batch's own facts are about its first table only.
		innerFacts := Facts{}
		if len(inner.Names) > 0 && len(s.Names) > 0 && inner.Names[0] == s.Names[0] {
			innerFacts = f
		}
		out = append(out, Implications(inner, innerFacts)...)
	}
	return out
}
