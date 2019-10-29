package reserve

import (
	"github.com/gin-gonic/gin"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
)

type Service struct {
	checkConcurrency     func(uint64) bool
	allocateReserve      func(ReserveRequest, bool) (Reserve, error)
	listUserFromDB       func(uint64) []Reserve
	listUserFromRegistry func(uint64) []Reserve
}

func NewService(
	checkConcurrency func(uint64) bool,
	allocateReserve func(ReserveRequest, bool) (Reserve, error),
	listUserFromDB func(uint64) []Reserve,
	listUserFromRegistry func(uint64) []Reserve,
) Service {
	return Service{
		checkConcurrency,
		allocateReserve,
		listUserFromDB,
		listUserFromRegistry,
	}
}

func (s *Service) HandleDBRequest(c *gin.Context) {
	var uri CreateURI
	if err := c.ShouldBindUri(&uri); err != nil {
		var errors []validationError
		for _, err := range err.(validator.ValidationErrors) {
			errors = append(errors, NewValidationError(err.Tag(), err.Field()))
		}

		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Invalid uri!",
			"code":    "invalid_uri",
			"errors":  errors,
		})
		return
	}

	c.JSON(http.StatusOK, s.listUserFromDB(uri.UserID))
	return
}

func (s *Service) HandleRegistryRequest(c *gin.Context) {
	var uri CreateURI
	if err := c.ShouldBindUri(&uri); err != nil {
		var errors []validationError
		for _, err := range err.(validator.ValidationErrors) {
			errors = append(errors, NewValidationError(err.Tag(), err.Field()))
		}

		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Invalid uri!",
			"code":    "invalid_uri",
			"errors":  errors,
		})
		return
	}

	c.JSON(http.StatusOK, s.listUserFromRegistry(uri.UserID))
	return
}

func (s *Service) HandleCreation(c *gin.Context) {
	var uri CreateURI
	if err := c.ShouldBindUri(&uri); err != nil {
		var errors []validationError
		for _, err := range err.(validator.ValidationErrors) {
			errors = append(errors, NewValidationError(err.Tag(), err.Field()))
		}

		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Invalid uri!",
			"code":    "invalid_uri",
			"errors":  errors,
		})
		return
	}

	var body Body
	if err := c.ShouldBindJSON(&body); err != nil {
		var errors []validationError
		for _, err := range err.(validator.ValidationErrors) {
			errors = append(errors, NewValidationError(err.Tag(), err.Field()))
		}

		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Invalid reserve!",
			"code":    "invalid_reserve",
			"errors":  errors,
		})
		return
	}

	var headers CreateHeader
	if err := c.ShouldBindHeader(&headers); err != nil {
		var errors []validationError
		for _, err := range err.(validator.ValidationErrors) {
			errors = append(errors, NewValidationError(err.Tag(), err.Field()))
		}

		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Invalid header!",
			"code":    "invalid_header",
			"errors":  errors,
		})
		return
	}

	var query CreateQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		var errors []validationError
		for _, err := range err.(validator.ValidationErrors) {
			errors = append(errors, NewValidationError(err.Tag(), err.Field()))
		}

		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Invalid query parameters!",
			"code":    "invalid_query_parameters",
			"errors":  errors,
		})
		return
	}

	if headers.ClientID == "" && query.ClientID == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Should provide clientID",
			"code":    "absent_client_id",
		})
		return
	}

	if headers.ClientID != "" && query.ClientID != "" && headers.ClientID == query.ClientID {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "clientID does not match",
			"code":    "mismatching_client_ids",
		})
		return
	}

	clientID := headers.ClientID
	if clientID == "" {
		clientID = query.ClientID
	}

	reserve, allocErr := s.allocateReserve(
		ReserveRequest{
			Body:           body,
			UserID:         uri.UserID,
			ClientID:       clientID,
			IdempotencyKey: headers.IdempotencyKey,
		},
		s.checkConcurrency(uri.UserID),
	)
	if allocErr != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, allocErr)
		return
	}

	c.JSON(http.StatusOK, reserve)
	return
}
