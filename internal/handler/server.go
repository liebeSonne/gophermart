package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"

	"github.com/liebeSonne/gophermart/api/server"
	"github.com/liebeSonne/gophermart/internal/auth"
	"github.com/liebeSonne/gophermart/internal/model"
)

func NewServer(
	userService UserService,
	userOrderService UserOrderService,
	userOrderQueryService UserOrderQueryService,
	userBalanceQueryService UserBalanceQueryService,
	userBalanceService UserBalanceService,
	userBalanceWithDrawnQueryService UserBalanceWithDrawnQueryService,
	tokenService TokenService,
	cookieService CookieService,
	logger *logrus.Logger,
) server.ServerInterface {
	return &Server{
		userService:                      userService,
		userOrderService:                 userOrderService,
		userOrderQueryService:            userOrderQueryService,
		userBalanceQueryService:          userBalanceQueryService,
		userBalanceService:               userBalanceService,
		userBalanceWithDrawnQueryService: userBalanceWithDrawnQueryService,
		tokenService:                     tokenService,
		cookieService:                    cookieService,
		logger:                           logger,
	}
}

type Server struct {
	userService                      UserService
	userOrderService                 UserOrderService
	userOrderQueryService            UserOrderQueryService
	userBalanceQueryService          UserBalanceQueryService
	userBalanceService               UserBalanceService
	userBalanceWithDrawnQueryService UserBalanceWithDrawnQueryService
	tokenService                     TokenService
	cookieService                    CookieService
	logger                           *logrus.Logger
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

	input := CreateUserInput{
		Login:    userCredentials.Login,
		Password: userCredentials.Password,
	}

	user, err := s.userService.Create(ctx, input)
	if err != nil {
		if errors.Is(err, ErrUserLoginExists) {
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
	w.WriteHeader(http.StatusOK)
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

	input := LoginUserInput{
		Login:    userCredentials.Login,
		Password: userCredentials.Password,
	}

	user, err := s.userService.Login(ctx, input)
	if err != nil {
		if errors.Is(err, ErrNotValidUserLoginPassword) || errors.Is(err, ErrUserNotFound) {
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
	w.WriteHeader(http.StatusOK)
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

	input := UploadUserOrderInput{
		OrderID: orderID,
		UserID:  userID,
	}

	_, err = s.userOrderService.Upload(ctx, input)
	if err != nil {
		if errors.Is(err, ErrInvalidOrderID) {
			http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
			return
		}
		if errors.Is(err, ErrUserOrderAlreadyUploadedByOtherUser) {
			http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
			return
		}
		if errors.Is(err, ErrUserOrderAlreadyUploadedByUser) {
			w.WriteHeader(http.StatusOK)
			return
		}

		s.logger.WithError(err).Errorf("error uploading user (userID: '%s') order (orderID: '%s')", userID, orderID)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

//nolint:dupl
func (s *Server) GetUserOrders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	w.Header().Set("Content-Type", "application/json")

	userID, ok := auth.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	items, err := s.userOrderQueryService.FindByUserID(ctx, userID)
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

	enc := json.NewEncoder(w)
	err = enc.Encode(resp)

	if err != nil {
		s.logger.WithError(err).Errorf("error encoding user (userID: %s) orders", userID)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) GetUserBalance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	w.Header().Set("Content-Type", "application/json")

	userID, ok := auth.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	userBalance, err := s.userBalanceQueryService.GetByUserID(ctx, userID)
	if err != nil {
		s.logger.WithError(err).Errorf("error getting user (userID: %s) balance", userID)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	resp, err := convertUserBalanceToAPI(userBalance)
	if err != nil {
		s.logger.WithError(err).Errorf("error converting user (userID: %s) balance", userID)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	enc := json.NewEncoder(w)
	err = enc.Encode(resp)

	if err != nil {
		s.logger.WithError(err).Errorf("error encoding user (userID: %s) orders", userID)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) WithdrawUserBalance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var withdrawnBalanceRequest server.WithdrawUserBalanceRequest
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&withdrawnBalanceRequest)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	userID, ok := auth.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	input := AddWithdrawnInput{
		UserID:  userID,
		OrderID: withdrawnBalanceRequest.Order,
		Amount:  decimal.NewFromFloat(withdrawnBalanceRequest.Sum),
	}

	err = s.userBalanceService.AddWithdrawn(ctx, input)
	if err != nil {
		if errors.Is(err, ErrInvalidOrderID) {
			http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
			return
		}
		if errors.Is(err, ErrUserBalanceIsNotEnough) {
			http.Error(w, http.StatusText(http.StatusPaymentRequired), http.StatusPaymentRequired)
			return
		}

		s.logger.WithError(err).Errorf("error add withdrawn (userID: '%s', orderID: '%s', amount: '%f')", userID, input.OrderID, withdrawnBalanceRequest.Sum)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

//nolint:dupl
func (s *Server) GetUserWithdrawals(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	w.Header().Set("Content-Type", "application/json")

	userID, ok := auth.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	items, err := s.userBalanceWithDrawnQueryService.FindByUserID(ctx, userID)
	if err != nil {
		s.logger.WithError(err).Errorf("error getting user (userID: %s) balance withdrawn", userID)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if len(items) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	resp, err := convertUserBalanceWithdrawnItemsToAPI(items)
	if err != nil {
		s.logger.WithError(err).Errorf("error converting user (userID: %s) balance withdrawn items", userID)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	enc := json.NewEncoder(w)
	err = enc.Encode(resp)

	if err != nil {
		s.logger.WithError(err).Errorf("error encoding user (userID: %s) balance withdrawn items", userID)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
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
