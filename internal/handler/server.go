package handler

import (
	"net/http"

	"github.com/liebeSonne/gophermart/api/server"
)

type Server struct{}

func NewServer() server.ServerInterface {
	return &Server{}
}

func (s *Server) GetUserBalance(w http.ResponseWriter, r *http.Request) {
	// TODO implement me
	_ = w
	_ = r
	panic("implement me")
}

func (s *Server) WithdrawUserBalance(w http.ResponseWriter, r *http.Request) {
	// TODO implement me
	_ = w
	_ = r
	panic("implement me")
}

func (s *Server) LoginUser(w http.ResponseWriter, r *http.Request) {
	// TODO implement me
	_ = w
	_ = r
	panic("implement me")
}

func (s *Server) GetUserOrders(w http.ResponseWriter, r *http.Request) {
	// TODO implement me
	_ = w
	_ = r
	panic("implement me")
}

func (s *Server) UploadUserOrders(w http.ResponseWriter, r *http.Request) {
	// TODO implement me
	_ = w
	_ = r
	panic("implement me")
}

func (s *Server) RegisterUser(w http.ResponseWriter, r *http.Request) {
	// TODO implement me
	_ = w
	_ = r
	panic("implement me")
}

func (s *Server) GetUserWithdrawals(w http.ResponseWriter, r *http.Request) {
	// TODO implement me
	_ = w
	_ = r
	panic("implement me")
}
