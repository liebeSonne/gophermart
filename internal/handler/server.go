package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/liebeSonne/gophermart/api/server"
	"github.com/liebeSonne/gophermart/internal/auth"
	"github.com/liebeSonne/gophermart/internal/handler/cookie"
	"github.com/liebeSonne/gophermart/internal/model"
	"github.com/liebeSonne/gophermart/internal/provider"
	"github.com/liebeSonne/gophermart/internal/service"
)

func NewServer(
	userService service.UserService,
	userOrderService service.UserOrderService,
	userOrderProvider provider.UserOrderProvider,
	tokenService auth.TokenService,
	cookieService cookie.Service,
	logger *logrus.Logger,
) server.ServerInterface {
	return &Server{
		userService:       userService,
		userOrderService:  userOrderService,
		userOrderProvider: userOrderProvider,
		tokenService:      tokenService,
		cookieService:     cookieService,
		logger:            logger,
	}
}

type Server struct {
	userService       service.UserService
	userOrderService  service.UserOrderService
	userOrderProvider provider.UserOrderProvider
	tokenService      auth.TokenService
	cookieService     cookie.Service
	logger            *logrus.Logger
}

//nolint:dupl
func (s *Server) RegisterUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var userCredentials server.UserCredentials
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&userCredentials)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	input := service.CreateUserInput{
		Login:    userCredentials.Login,
		Password: userCredentials.Password,
	}

	user, err := s.userService.Create(ctx, input)
	if err != nil {
		if errors.Is(err, service.ErrUserLoginExists) {
			http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
			return
		}

		s.logger.WithError(err).Errorf("error creating user (login: '%s')", input.Login)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = s.setUserAuthorization(w, r, user)
	if err != nil {
		s.logger.WithError(err).Errorf("error setting user (id: '%s', login: '%s') authorization", user.ID, user.Login)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
}

//nolint:dupl
func (s *Server) LoginUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var userCredentials server.UserCredentials
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&userCredentials)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	input := service.LoginUserInput{
		Login:    userCredentials.Login,
		Password: userCredentials.Password,
	}

	user, err := s.userService.Login(ctx, input)
	if err != nil {
		if errors.Is(err, service.ErrNotValidUserLoginPassword) || errors.Is(err, service.ErrUserNotFound) {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		s.logger.WithError(err).Errorf("error creating user (login: '%s')", input.Login)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = s.setUserAuthorization(w, r, user)
	if err != nil {
		s.logger.WithError(err).Errorf("error setting user (id: '%s', login: '%s') authorization", user.ID, user.Login)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
}

func (s *Server) UploadUserOrders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	orderID := string(body)

	userID, ok := auth.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	input := service.UploadUserOrderInput{
		OrderID: orderID,
		UserID:  userID,
	}

	_, err = s.userOrderService.Upload(ctx, input)
	if err != nil {
		if errors.Is(err, service.ErrInvalidOrderID) {
			http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
			return
		}
		if errors.Is(err, service.ErrUserOrderAlreadyUploadedByOtherUser) {
			http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
			return
		}
		if errors.Is(err, service.ErrUserOrderAlreadyUploadedByUser) {
			w.WriteHeader(http.StatusOK)
			return
		}

		s.logger.WithError(err).Errorf("error uploading user (userID: '%s') order (orderID: '%s')", userID, orderID)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (s *Server) GetUserOrders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	w.Header().Set("Content-Type", "application/json")

	userID, ok := auth.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	items, err := s.userOrderProvider.FindByUserID(ctx, userID)
	if err != nil {
		s.logger.WithError(err).Errorf("error getting user (userID: %s) orders", userID)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if len(items) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	resp, err := convertUserOrdersToAPI(items)
	if err != nil {
		s.logger.WithError(err).Errorf("error converting user (userID: %s) orders", userID)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

	enc := json.NewEncoder(w)
	err = enc.Encode(resp)

	if err != nil {
		s.logger.WithError(err).Errorf("error encoding user (userID: %s) orders", userID)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
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

func (s *Server) GetUserWithdrawals(w http.ResponseWriter, r *http.Request) {
	// TODO implement me
	_ = w
	_ = r
	panic("implement me")
}

func (s *Server) setUserAuthorization(w http.ResponseWriter, r *http.Request, user model.User) error {
	tokenData := auth.Token{
		UserID: user.ID.String(),
	}

	tokenString, err := s.tokenService.Create(tokenData)
	if err != nil {
		return err
	}

	err = s.cookieService.SetAuthToken(tokenString, w, r)
	if err != nil {
		return err
	}

	return nil
}
