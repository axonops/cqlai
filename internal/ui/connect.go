package ui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"

	"github.com/axonops/cqlai/internal/config"
	"github.com/axonops/cqlai/internal/db"
	"github.com/axonops/cqlai/internal/logger"
	"github.com/axonops/cqlai/internal/router"
	"github.com/axonops/cqlai/internal/session"
	"github.com/axonops/cqlai/internal/ui/completion"
)

// Connecting to a cluster from inside cqlai.
//
// cqlai used to refuse to start without one, which left the connection that
// stopped it starting as the one thing it could not be used to fix. It starts
// either way now, and CONNECT is how a cluster is chosen: the same window
// PREFERENCES uses, showing the settings a connection takes.

// pressPrefButton carries out whichever button was pressed.
func (m *MainModel) pressPrefButton(label string) (*MainModel, tea.Cmd) {
	switch label {
	case prefCancel:
		m.closePreferences()
		return m, nil
	case prefSave:
		return m.savePreferences()
	case prefConnect:
		return m.connectWith(false)
	case prefSaveAndConnect:
		return m.connectWith(true)
	}
	return m, nil
}

// connectWith tries the connection the window describes, saving it first when
// that is the button that was pressed.
//
// A connection that fails leaves the window open with what went wrong on it:
// what you want next is to change a field and try again, not to type it all
// out a second time.
func (m *MainModel) connectWith(save bool) (*MainModel, tea.Cmd) {
	if !m.preferencesReady() {
		return m, nil
	}

	cfg := m.preferences.cfg
	for _, field := range m.preferences.fields {
		if err := setPrefValue(cfg, field.spec.path, field.value()); err != nil {
			m.preferences.failed = err.Error()
			return m, nil
		}
	}
	prunePrefs(cfg)

	if save {
		written, err := cfg.Save()
		if err != nil {
			m.preferences.failed = "Not saved: " + err.Error()
			return m, nil
		}
		m.preferences.saved = written
	}

	session, err := db.NewSessionWithOptions(db.SessionOptions{
		Host:           cfg.Host,
		Port:           cfg.Port,
		Keyspace:       cfg.Keyspace,
		Username:       cfg.Username,
		Password:       cfg.Password,
		Consistency:    cfg.Consistency,
		SSL:            cfg.SSL,
		ConnectTimeout: cfg.ConnectTimeout,
		RequestTimeout: cfg.RequestTimeout,
	})
	if err != nil {
		logger.DebugfToFile("Connect", "Connecting to %s:%d: %v", cfg.Host, cfg.Port, err)
		m.preferences.failed = err.Error()
		return m, nil
	}

	saved := m.preferences.saved
	m.closePreferences()
	m.adoptSession(session, cfg)

	said := fmt.Sprintf("Connected to %s:%d", cfg.Host, cfg.Port)
	if saved != "" {
		said += ", and the settings saved to " + saved
	}
	return m.report(said)
}

// adoptSession puts a new connection in place of whatever was there.
//
// Everything that was built from the old session has to be built again, or the
// shell goes on completing against the cluster it is no longer talking to and
// showing the schema of one it has left.
func (m *MainModel) adoptSession(newSession *db.Session, cfg *config.Config) {
	if m.session != nil {
		m.session.Close()
	}

	m.session = newSession
	m.config = cfg
	m.connectError = ""

	m.sessionManager = session.NewManager(cfg)
	router.InitRouter(m.sessionManager)
	m.completionEngine = completion.NewCompletionEngine(newSession, m.sessionManager)

	m.statusBar.Host = cfg.Host
	m.statusBar.Username = cfg.Username
	m.statusBar.Keyspace = cfg.Keyspace
	m.statusBar.Consistency = newSession.Consistency()

	// The schema is a different cluster's now, and so is the version being
	// watched: forget both, and let the tree fetch what is there when it is
	// next looked at.
	m.schema = schemaBrowser{}
	m.schemaVersion = ""
}
