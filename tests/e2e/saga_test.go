package e2e

import (
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/suite"
)

type SagaTestSuite struct {
	suite.Suite
	client *resty.Client
	
	customerAPI   string
	restaurantAPI string
	orderingAPI   string
	deliveryAPI   string
}

func (s *SagaTestSuite) SetupSuite() {
	s.customerAPI = getEnv("CUSTOMER_API", "http://localhost:8080")
	s.restaurantAPI = getEnv("RESTAURANT_API", "http://localhost:8081")
	s.orderingAPI = getEnv("ORDERING_API", "http://localhost:8083")
	s.deliveryAPI = getEnv("DELIVERY_API", "http://localhost:8084")
	
	s.client = resty.New().SetTimeout(5 * time.Second)
}

func TestSagaTestSuite(t *testing.T) {
	suite.Run(t, new(SagaTestSuite))
}

func (s *SagaTestSuite) TestHappyPath() {
	// 1. Register Customer
	ts := time.Now().UnixNano()
	customerID := s.registerCustomer("John Doe", fmt.Sprintf("john-%d@example.com", ts), "123456789")
	s.NotEmpty(customerID)

	// 2. Setup Restaurant and Menu (Assuming "res-1" exists for test)
	restaurantID := "res-1" 
	
	// 3. Create Order
	orderReq := map[string]interface{}{
		"customer_id":   customerID,
		"restaurant_id": restaurantID,
		"card_token":    "tok_visa_success",
		"items": []map[string]interface{}{
			{
				"product_id": "p1",
				"name":       "Burger",
				"quantity":   1,
				"price":      10.5,
			},
		},
	}

	var orderResp struct {
		Id string `json:"id"`
	}
	resp, err := s.client.R().
		SetBody(orderReq).
		SetResult(&orderResp).
		Post(s.orderingAPI + "/api/v1/orders")

	s.NoError(err)
	s.Equal(http.StatusAccepted, resp.StatusCode())
	orderID := orderResp.Id
	s.NotEmpty(orderID)

	// 4. Poll for Order Status until PREPARING (terminal happy path state) or timeout
	s.waitForOrderStatus(orderID, "PREPARING", 20*time.Second) 
}

func (s *SagaTestSuite) TestPaymentRejected() {
	ts := time.Now().UnixNano()
	customerID := s.registerCustomer("Jane Doe", fmt.Sprintf("jane-%d@example.com", ts), "987654321")
	
	orderReq := map[string]interface{}{
		"customer_id":   customerID,
		"restaurant_id": "res-1",
		"card_token":    "tok_failure", // Triggers failure in payment mock
		"items": []map[string]interface{}{
			{
				"product_id": "p1",
				"name":       "Burger",
				"quantity":   1,
				"price":      10.5,
			},
		},
	}

	var orderResp struct {
		Id string `json:"id"`
	}
	resp, err := s.client.R().
		SetBody(orderReq).
		SetResult(&orderResp).
		Post(s.orderingAPI + "/api/v1/orders")

	s.NoError(err)
	s.Equal(http.StatusAccepted, resp.StatusCode())
	
	// Should end up as REJECTED
	s.waitForOrderStatus(orderResp.Id, "REJECTED", 20*time.Second)
}

func (s *SagaTestSuite) TestRestaurantRejected() {
	ts := time.Now().UnixNano()
	customerID := s.registerCustomer("Rick Sanchez", fmt.Sprintf("rick-%d@example.com", ts), "999999999")
	
	orderReq := map[string]interface{}{
		"customer_id":   customerID,
		"restaurant_id": "res-1",
		"card_token":    "tok_visa_success",
		"items": []map[string]interface{}{
			{
				"product_id": "p-unhappy",
				"name":       "REJECT_ME", // Triggers rejection in restaurant service
				"quantity":   1,
				"price":      20.0,
			},
		},
	}

	var orderResp struct {
		Id string `json:"id"`
	}
	resp, err := s.client.R().
		SetBody(orderReq).
		SetResult(&orderResp).
		Post(s.orderingAPI + "/api/v1/orders")

	s.NoError(err)
	s.Equal(http.StatusAccepted, resp.StatusCode())
	
	// Should end up as CANCELLED (compensating flow after payment authorized)
	s.waitForOrderStatus(orderResp.Id, "CANCELLED", 20*time.Second)
}

func (s *SagaTestSuite) registerCustomer(name, email, phone string) string {
	var respData struct {
		Id string `json:"id"`
	}
	resp, err := s.client.R().
		SetBody(map[string]interface{}{
			"name":  name,
			"email": email,
			"phone": phone,
		}).
		SetResult(&respData).
		Post(s.customerAPI + "/api/v1/customers")

	s.NoError(err)
	s.Equal(http.StatusCreated, resp.StatusCode())
	return respData.Id
}

func (s *SagaTestSuite) waitForOrderStatus(orderID string, expectedStatus string, timeout time.Duration) {
	start := time.Now()
	for time.Since(start) < timeout {
		var statusResp struct {
			Status string `json:"status"`
		}
		resp, err := s.client.R().
			SetResult(&statusResp).
			Get(fmt.Sprintf("%s/api/v1/orders/%s", s.orderingAPI, orderID))

		if err == nil && resp.StatusCode() == http.StatusOK {
			if statusResp.Status == expectedStatus {
				return
			}
			// If we expected a success state but got terminal failure early
			if expectedStatus == "PREPARING" && (statusResp.Status == "REJECTED" || statusResp.Status == "CANCELLED") {
				s.Failf("Order was rejected or cancelled prematurely", "expected %s but got %s", expectedStatus, statusResp.Status)
				return
			}
		}
		time.Sleep(1 * time.Second)
	}
	s.Fail("Timeout waiting for order status " + expectedStatus)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
