package config

import "strings"

// Connections that have been kept.
//
// CONNECT edited one set of settings, which was whatever the file held, so
// working with a second cluster meant typing the host, the credentials and the
// SSL paths again every time - and losing the first one by saving the second.
//
// A saved connection is a configuration with only the connection settings in
// it, under a name. That is what lets the window edit one the same way it edits
// the file: the same fields, the same checking, the same reflection over the
// same paths.

// ConnectionName is what a connection is called: the name it was given, or the
// host it connects to.
//
// A name is asked for and rarely wanted - one cluster per host is the ordinary
// case - so an empty one falls back to the host rather than refusing to save.
func ConnectionName(conn Config) string {
	if name := strings.TrimSpace(conn.Name); name != "" {
		return name
	}
	return strings.TrimSpace(conn.Host)
}

// ConnectionSettings is a connection on its own: the settings that say where to
// connect and how, and nothing else.
//
// A connection is kept in the same file as everything else, and writing the
// whole configuration under each one would put a copy of the API keys and the
// output format beside every host.
func ConnectionSettings(conn Config) Config {
	kept := Config{
		Name:           conn.Name,
		Host:           conn.Host,
		Port:           conn.Port,
		Keyspace:       conn.Keyspace,
		Username:       conn.Username,
		Password:       conn.Password,
		ConnectTimeout: conn.ConnectTimeout,
		RequestTimeout: conn.RequestTimeout,
	}
	if conn.SSL != nil {
		ssl := *conn.SSL
		kept.SSL = &ssl
	}
	return kept
}

// SaveConnection puts a connection in the list, in place of the one of the same
// name or at the end when it is new. It reports the name it was saved under.
func (c *Config) SaveConnection(conn Config) string {
	kept := ConnectionSettings(conn)
	kept.Name = ConnectionName(conn)

	for i, existing := range c.Connections {
		if strings.EqualFold(ConnectionName(existing), kept.Name) {
			c.Connections[i] = kept
			return kept.Name
		}
	}

	c.Connections = append(c.Connections, kept)
	return kept.Name
}

// DefaultConnection is the one cqlai opens with: the first of them.
//
// First rather than a name kept somewhere else pointing at one: a name that
// can point at a connection that is not there is a second thing to keep in
// step, and the order is already written down.
func (c *Config) DefaultConnection() (Config, bool) {
	if len(c.Connections) == 0 {
		return Config{}, false
	}
	return c.Connections[0], true
}

// MakeDefault saves a connection and puts it first, which is what makes it the
// one cqlai opens with. It reports the name it was saved under.
func (c *Config) MakeDefault(conn Config) string {
	name := c.SaveConnection(conn)

	for i, existing := range c.Connections {
		if strings.EqualFold(ConnectionName(existing), name) {
			kept := c.Connections[i]
			c.Connections = append(c.Connections[:i], c.Connections[i+1:]...)
			c.Connections = append([]Config{kept}, c.Connections...)
			break
		}
	}

	c.UseConnection(conn)
	return name
}

// Connection is the saved connection of this name, and whether there is one.
func (c *Config) Connection(name string) (Config, bool) {
	for _, conn := range c.Connections {
		if strings.EqualFold(ConnectionName(conn), name) {
			return conn, true
		}
	}
	return Config{}, false
}

// UseConnection puts a connection's settings in place of the ones the
// configuration starts with, so that the connection last saved is the one
// cqlai opens with next time.
func (c *Config) UseConnection(conn Config) {
	c.Name = ConnectionName(conn)
	c.Host = conn.Host
	c.Port = conn.Port
	c.Keyspace = conn.Keyspace
	c.Username = conn.Username
	c.Password = conn.Password
	c.ConnectTimeout = conn.ConnectTimeout
	c.RequestTimeout = conn.RequestTimeout

	c.SSL = nil
	if conn.SSL != nil {
		ssl := *conn.SSL
		c.SSL = &ssl
	}
}
