package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestALimitBiggerThanAPageStreams. PAGING 100 with LIMIT 300 used to fetch all
// 300 at once, because the limit was compared against a hardcoded 1000 that had
// nothing to do with the page size.
func TestALimitBiggerThanAPageStreams(t *testing.T) {
	s := &Session{}
	s.SetPageSize(100)

	assert.True(t, s.shouldUseStreaming("SELECT * FROM t LIMIT 300"))
	assert.True(t, s.shouldUseStreaming("SELECT * FROM t LIMIT 1000"))
	assert.True(t, s.shouldUseStreaming("SELECT * FROM t LIMIT 101"))
}

// TestALimitInsideAPageDoesNot: there is nothing to page through.
func TestALimitInsideAPageDoesNot(t *testing.T) {
	s := &Session{}
	s.SetPageSize(100)

	assert.False(t, s.shouldUseStreaming("SELECT * FROM t LIMIT 100"))
	assert.False(t, s.shouldUseStreaming("SELECT * FROM t LIMIT 10"))
	assert.False(t, s.shouldUseStreaming("select * from t limit 1"))
}

// TestTheThresholdFollowsThePageSize, rather than being a constant.
func TestTheThresholdFollowsThePageSize(t *testing.T) {
	s := &Session{}

	s.SetPageSize(500)
	assert.False(t, s.shouldUseStreaming("SELECT * FROM t LIMIT 300"),
		"300 fits in a page of 500")

	s.SetPageSize(50)
	assert.True(t, s.shouldUseStreaming("SELECT * FROM t LIMIT 300"),
		"300 does not fit in a page of 50")
}

// TestPagingOffAlwaysStreams: with no page size there is nothing to measure a
// limit against, so the sliding window handles it as it does everything else.
func TestPagingOffAlwaysStreams(t *testing.T) {
	s := &Session{}
	s.SetPageSize(0)

	assert.True(t, s.shouldUseStreaming("SELECT * FROM t LIMIT 10"))
	assert.True(t, s.shouldUseStreaming("SELECT * FROM t"))
}

// TestAQueryWithNoLimitStreams.
func TestAQueryWithNoLimitStreams(t *testing.T) {
	s := &Session{}
	s.SetPageSize(100)

	assert.True(t, s.shouldUseStreaming("SELECT * FROM t"))
	assert.True(t, s.shouldUseStreaming("SELECT * FROM t WHERE id = 1"))
}

// TestAPerPartitionLimitIsStillPaged: PER PARTITION LIMIT 1 matched the
// pattern for LIMIT 1, which fits in a page, so the whole result - one row of
// every partition in the table - was read in one go.
func TestAPerPartitionLimitIsStillPaged(t *testing.T) {
	s := &Session{}
	s.SetPageSize(100)
	assert.True(t, s.shouldUseStreaming("SELECT * FROM ks.t PER PARTITION LIMIT 1"))
	assert.True(t, s.shouldUseStreaming("SELECT * FROM ks.t WHERE s = 'LIMIT 5'"))
	assert.True(t, s.shouldUseStreaming("SELECT * FROM ks.t LIMIT 500"))
	assert.False(t, s.shouldUseStreaming("SELECT * FROM ks.t PER PARTITION LIMIT 1 LIMIT 50"))
	assert.False(t, s.shouldUseStreaming("SELECT * FROM ks.t LIMIT 10"))
}
