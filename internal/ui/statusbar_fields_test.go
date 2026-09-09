package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testStatusBar() StatusBarModel {
	return StatusBarModel{
		Username:    "cassandra",
		Host:        "127.0.0.1",
		Keyspace:    "system",
		Consistency: "LOCAL_ONE",
		PagingSize:  100,
		Tracing:     false,
		AutoFetch:   true,
		Version:     "5.0.9",
	}
}

// TestSegmentsMatchTheRenderedLine is the property everything else rests on:
// the columns used for hit testing have to describe the text actually drawn.
// Checking the two against each other, rather than against the same
// arithmetic, is what catches them drifting apart.
func TestSegmentsMatchTheRenderedLine(t *testing.T) {
	m := testStatusBar()

	rendered := []rune(stripAnsiForTest(m.View(140, DefaultStyles(), "history")))

	for _, seg := range m.segments() {
		require.LessOrEqual(t, seg.end, len(rendered),
			"segment %q runs past the end of the line", seg.label)

		got := string(rendered[seg.start:seg.end])
		assert.Equal(t, seg.label+seg.value, got,
			"columns %d..%d should hold %q but the line has %q",
			seg.start, seg.end, seg.label+seg.value, got)
	}
}

// TestSettingAtResolvesClicks covers turning a column into a setting.
func TestSettingAtResolvesClicks(t *testing.T) {
	m := testStatusBar()

	for _, seg := range m.segments() {
		if !seg.clickable() {
			continue
		}
		for _, col := range []int{seg.start, (seg.start + seg.end) / 2, seg.end - 1} {
			setting, at, ok := m.settingAt(col)
			assert.True(t, ok, "column %d is inside %q and should be clickable", col, seg.setting)
			assert.Equal(t, seg.setting, setting)
			assert.Equal(t, seg.start, at, "the chooser anchors to the start of the field")
		}
	}
}

// TestConnectionFactsAreNotClickable: version, user and host describe the
// connection rather than settings, so clicking them must do nothing rather than
// open an empty list.
func TestConnectionFactsAreNotClickable(t *testing.T) {
	m := testStatusBar()

	for _, seg := range m.segments() {
		if seg.clickable() {
			continue
		}
		_, _, ok := m.settingAt((seg.start + seg.end) / 2)
		assert.False(t, ok, "%q is a fact, not a setting", seg.label)
	}
}

func TestSettingAtOutsideAnyField(t *testing.T) {
	m := testStatusBar()

	_, _, ok := m.settingAt(0) // the padding column
	assert.False(t, ok)

	_, _, ok = m.settingAt(10000) // past the end
	assert.False(t, ok)
}

// TestSegmentsDoNotOverlap: overlapping ranges would make a click ambiguous.
func TestSegmentsDoNotOverlap(t *testing.T) {
	segs := testStatusBar().segments()
	require.NotEmpty(t, segs)

	for i := 1; i < len(segs); i++ {
		assert.GreaterOrEqual(t, segs[i].start, segs[i-1].end,
			"%q overlaps %q", segs[i].label, segs[i-1].label)
	}
}

// TestSettingChoicesMarkTheCurrentValue: the list should open on what is
// already set, so picking the current value is a no-op rather than a surprise.
func TestSettingChoicesMarkTheCurrentValue(t *testing.T) {
	tests := []struct {
		setting string
		current string
		want    string
	}{
		{settingConsistency, "LOCAL_QUORUM", "LOCAL_QUORUM"},
		{settingConsistency, "ONE", "ONE"},
		{settingPaging, "100", "100"},
		{settingTracing, "OFF", "OFF"},
		{settingAutoFetch, "ON", "ON"},
	}

	for _, tt := range tests {
		choices, selected := settingChoices(tt.setting, tt.current)
		require.NotEmpty(t, choices, "%s should offer choices", tt.setting)
		require.Less(t, selected, len(choices))
		assert.Equal(t, tt.want, choices[selected],
			"%s should open on its current value", tt.setting)
	}
}

// TestSettingChoicesFallBackToTheFirst covers a current value that is not in
// the list, which happens for a consistency level we do not offer.
func TestSettingChoicesFallBackToTheFirst(t *testing.T) {
	choices, selected := settingChoices(settingConsistency, "SOMETHING_ELSE")
	require.NotEmpty(t, choices)
	assert.Equal(t, 0, selected)
}

func TestSettingChoicesUnknownSetting(t *testing.T) {
	choices, _ := settingChoices("nonsense", "")
	assert.Nil(t, choices)

	// Keyspaces come from the cluster, so they are not baked in here.
	choices, _ = settingChoices(settingKeyspace, "system")
	assert.Nil(t, choices)
}
