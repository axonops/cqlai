package db

// ConnectionInfo describes the connection, for the panel behind Connection on
// the status line.
//
// The version, the user and the host used to sit on the status line as three
// separate fields, which took a third of it to say things that rarely change.
// They are one field now, and this is what opens behind it - with the
// encryption settings, which were not shown anywhere at all.
type ConnectionInfo struct {
	Version  string // Cassandra release version
	Protocol int    // native protocol version negotiated
	Username string
	Host     string // host and port, as connected

	// Encryption. The rest are meaningless unless Encrypted is set.
	Encrypted         bool
	HostVerified      bool // the server's certificate is checked against its name
	ClientCertificate bool // we present one of our own
	CustomCA          bool // the server's certificate is checked against a CA we were given
}

// ConnectionInfo reports how this session is connected.
//
// It reads the cluster configuration rather than the wire, so it says what was
// asked for. For TLS that is the same thing: gocql fails the connection if the
// handshake does not meet the configuration, so a session that exists is one
// where these held.
func (s *Session) ConnectionInfo() ConnectionInfo {
	info := ConnectionInfo{
		Version:  s.CassandraVersion(),
		Username: s.username,
	}
	if info.Username == "" {
		info.Username = "(anonymous)"
	}

	if s.cluster == nil {
		return info
	}

	info.Protocol = s.cluster.ProtoVersion
	if len(s.cluster.Hosts) > 0 {
		info.Host = s.cluster.Hosts[0]
	}

	if s.cluster.SslOpts == nil {
		return info
	}
	info.Encrypted = true

	if c := s.cluster.SslOpts.Config; c != nil {
		// InsecureSkipVerify is the switch that turns certificate checking
		// off, so saying it plainly is the point: an encrypted connection that
		// verifies nothing is not the same as one that does.
		info.HostVerified = !c.InsecureSkipVerify
		info.ClientCertificate = len(c.Certificates) > 0
		info.CustomCA = c.RootCAs != nil
	}
	return info
}
