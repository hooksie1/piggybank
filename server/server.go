package server

import "log"

var (
	databaseKey []byte
	piggyBucket = "piggybank"
)

type Server struct {
	Backend Backend
}

func (s *Server) SetBackend(b Backend) {
	s.Backend = b
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) Serve() {
	log.Printf("starting server")
	s.Backend.Watch()
}
