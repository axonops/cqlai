package mcp

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/axonops/cqlai/internal/ai"
	"github.com/axonops/cqlai/internal/db"
	"github.com/axonops/cqlai/internal/policy"
	"github.com/axonops/cqlai/internal/router"
	"github.com/axonops/cqlai/internal/session"
)

// cluster is the server's connection, made when it can be rather than before
// the server starts.
//
// An MCP client starts the server when it starts, which is not when the
// cluster is up: a laptop's local cluster, a VPN not yet connected. A server
// that exits then is one the client shows as broken until it is restarted by
// hand. So the server starts regardless, says it is not connected when asked,
// and tries again when a tool is called.
//
// It is a session of the server's own, with a schema cache of its own, even
// in the terminal app: a model's queries must not change the shell's session
// - its consistency, its tracing - and the shell's must not change theirs.
type cluster struct {
	opts  db.SessionOptions
	where string              // which connection, and where it came from, for messages
	scrub func(string) string // keeps the password out of a driver's error
	mgr   *session.Manager

	// retry is how long after a failed attempt the next one waits. A model
	// calling tools in a loop must not turn into a connection storm.
	retry time.Duration

	mu      sync.Mutex
	session *db.Session
	schema  *ai.AI
	failed  string // why the last attempt failed
	tried   time.Time
}

// start makes the first attempt in the background, so the client's
// initialize is answered at once.
func (c *cluster) start() {
	go func() { _, _, _ = c.connected() }()
}

// connected is the session and its schema cache, connecting first if there
// is none and the last attempt was long enough ago. Without one, it says why.
func (c *cluster) connected() (*db.Session, *ai.AI, string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.session != nil {
		return c.session, c.schema, ""
	}
	if c.failed != "" && time.Since(c.tried) < c.retry {
		return nil, nil, c.failed
	}

	c.tried = time.Now()
	sess, err := db.NewSessionWithOptions(c.opts)
	if err != nil {
		c.failed = fmt.Sprintf("not connected to %s: %s", c.where, c.scrub(err.Error()))
		return nil, nil, c.failed
	}
	schema, err := ai.NewSchemaTools(sess)
	if err != nil {
		sess.Close()
		c.failed = fmt.Sprintf("connected to %s, and could not read the schema: %s", c.where, c.scrub(err.Error()))
		return nil, nil, c.failed
	}
	c.session, c.schema, c.failed = sess, schema, ""
	return sess, schema, ""
}

// env is what a call runs against under pol: the session as it is now, or
// the reason there is none.
func (c *cluster) env(pol policy.Policy) ai.ToolEnv {
	sess, schema, reason := c.connected()
	if sess == nil {
		return ai.ToolEnv{Policy: pol, NotConnected: reason}
	}
	return ai.ToolEnv{
		// What the gate needs to tell a scan from a read of one partition,
		// and whether a redaction applies to a table.
		Policy: pol.
			WithPartitionKey(func(keyspace, table string) []string {
				key, _ := sess.TableKey(keyspace, table)
				return key
			}).
			WithColumns(func(keyspace, table string) []string {
				_, columns := sess.TableKey(keyspace, table)
				return columns
			}),
		Session: sess,
		Schema:  schema,
		Describe: func(statement string) interface{} {
			return router.ProcessCommand(statement, sess, c.mgr)
		},
	}
}

// close closes the session, if there is one.
func (c *cluster) close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.session != nil {
		c.session.Close()
		c.session, c.schema = nil, nil
	}
}

// waitForFirstAttempt returns once the first attempt has finished.
func (c *cluster) waitForFirstAttempt(ctx context.Context) {
	for {
		c.mu.Lock()
		done := !c.tried.IsZero() && (c.session != nil || c.failed != "")
		c.mu.Unlock()
		if done || ctx.Err() != nil {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// fixed is a source with one policy and one cluster: the headless server.
type fixed struct {
	pol  policy.Policy
	link *cluster
}

func (f fixed) Policy() policy.Policy { return f.pol }
func (f fixed) Env() ai.ToolEnv       { return f.link.env(f.pol) }
