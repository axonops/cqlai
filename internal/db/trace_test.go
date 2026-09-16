package db

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The trace of a paged query.
//
// A paged query is one request per page, each traced separately by Cassandra
// under its own session id. The id was kept in a field of the tracer that was
// copied to the session when the first page came back, so every page after
// that was traced and thrown away; then only the last was kept, which says a
// scan read sixty-five rows when the query read five hundred. All of them are
// kept now.

// TestEveryPageOfAQueryIsTraced.
func TestEveryPageOfAQueryIsTraced(t *testing.T) {
	session := &Session{}
	session.startTracing()
	tracer := &captureTracer{session: session}

	ids, made := session.tracedRequests()
	assert.Empty(t, ids, "nothing has been asked of Cassandra yet")
	assert.Equal(t, 0, made)

	tracer.Trace([]byte("first-page"))
	tracer.Trace([]byte("second-page"))
	tracer.Trace([]byte("third-page"))

	ids, made = session.tracedRequests()
	assert.Equal(t, 3, made)
	require.Len(t, ids, 3, "every page, in the order they were fetched")
	assert.Equal(t, []byte("first-page"), ids[0])
	assert.Equal(t, []byte("third-page"), ids[2])
}

// TestOnlySoManyPagesAreKept.
//
// A query read to the end with auto-fetch on is hundreds of requests, and a
// trace of all of them is a table nobody reads. The most recent are the ones
// worth having, and the count says how many there were.
func TestOnlySoManyPagesAreKept(t *testing.T) {
	session := &Session{}
	session.startTracing()
	tracer := &captureTracer{session: session}

	for i := range tracedPagesKept + 10 {
		tracer.Trace([]byte{byte(i)})
	}

	ids, made := session.tracedRequests()
	assert.Equal(t, tracedPagesKept+10, made, "the query made this many requests")
	require.Len(t, ids, tracedPagesKept, "and this many are kept")
	assert.Equal(t, []byte{byte(tracedPagesKept + 9)}, ids[len(ids)-1], "the most recent among them")
}

// TestTheNextQueryStartsCounting again, rather than carrying on from the last
// one's pages.
func TestTheNextQueryStartsCounting(t *testing.T) {
	session := &Session{}
	tracer := &captureTracer{session: session}

	session.startTracing()
	tracer.Trace([]byte("a"))
	tracer.Trace([]byte("b"))

	session.startTracing()
	ids, made := session.tracedRequests()
	assert.Empty(t, ids)
	assert.Equal(t, 0, made)
}

// TestPagesAreFetchedOnWhicheverGoroutineWantsThem, which is why the traces are
// behind a lock: the driver fetches the next page ahead on one of its own.
func TestPagesAreFetchedOnWhicheverGoroutineWantsThem(t *testing.T) {
	session := &Session{}
	session.startTracing()
	tracer := &captureTracer{session: session}

	var pages sync.WaitGroup
	for range 50 {
		pages.Add(1)
		go func() {
			defer pages.Done()
			tracer.Trace([]byte("a page"))
		}()
	}
	go func() {
		for range 50 {
			_, _ = session.tracedRequests()
		}
	}()
	pages.Wait()

	_, made := session.tracedRequests()
	assert.Equal(t, 50, made)
}
