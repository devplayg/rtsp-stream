package server

type Watcher struct {
	server *Server
}

func NewWatcher(server *Server) *Watcher {
	return &Watcher{
		server: server,
	}
}
