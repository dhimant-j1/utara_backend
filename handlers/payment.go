package handlers

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	"utara_backend/config"
	"utara_backend/models"
)

// Razorpay API credentials
func getRazorpayCredentials() (keyID, keySecret string) {
	return os.Getenv("RAZORPAY_KEY_ID"), os.Getenv("RAZORPAY_KEY_SECRET")
}

// CreatePayment creates a new Razorpay order and stores payment record
func CreatePayment(c *gin.Context) {
	var req models.CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userObjID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	razorpayOrder, err := createRazorpayOrder(req.Amount, req.Currency, req.Description, req.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create payment order: " + err.Error()})
		return
	}

	// Create payment record
	payment := models.Payment{
		UserID:          userObjID,
		Amount:          req.Amount,
		Currency:        req.Currency,
		Type:            req.Type,
		Status:          models.PaymentStatusCreated,
		RazorpayOrderID: razorpayOrder.ID,
		Description:     req.Description,
		Notes:           req.Notes,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Add request ID if provided
	if req.RequestID != nil {
		reqObjID, err := primitive.ObjectIDFromHex(*req.RequestID)
		if err == nil {
			payment.RequestID = &reqObjID
		}
	}

	result, err := config.DB.Collection("payments").InsertOne(context.Background(), payment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save payment record"})
		return
	}

	payment.ID = result.InsertedID.(primitive.ObjectID)

	// Return payment details with Razorpay key
	razorpayKey, _ := getRazorpayCredentials()
	c.JSON(http.StatusCreated, models.PaymentResponse{
		Payment:     payment,
		RazorpayKey: razorpayKey,
	})
}

// createRazorpayOrder creates an order on Razorpay
func createRazorpayOrder(amount int, currency, description string, notes map[string]string) (*models.RazorpayOrderResponse, error) {
	keyID, keySecret := getRazorpayCredentials()
	
	payload := map[string]interface{}{
		"amount":   amount,
		"currency": currency,
		"receipt":  fmt.Sprintf("receipt_%d", time.Now().Unix()),
		"notes":    notes,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "https://api.razorpay.com/v1/orders", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(keyID, keySecret)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("razorpay API error: %s", string(body))
	}

	var order models.RazorpayOrderResponse
	if err := json.Unmarshal(body, &order); err != nil {
		return nil, err
	}

	return &order, nil
}

// VerifyPayment verifies the Razorpay payment signature
func VerifyPayment(c *gin.Context) {
	var req models.VerifyPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	_, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Verify signature
	if !verifyRazorpaySignature(req.OrderID, req.PaymentID, req.Signature) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment signature"})
		return
	}

	// Find and update payment record
	filter := bson.M{"razorpay_order_id": req.OrderID}
	update := bson.M{
		"$set": bson.M{
			"status":             models.PaymentStatusPaid,
			"razorpay_payment_id": req.PaymentID,
			"razorpay_signature":  req.Signature,
			"updated_at":          time.Now(),
			"paid_at":             time.Now(),
		},
	}

	var payment models.Payment
	err = config.DB.Collection("payments").FindOneAndUpdate(
		context.Background(),
		filter,
		update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&payment)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update payment record"})
		return
	}

	c.JSON(http.StatusOK, payment)
}

// verifyRazorpaySignature verifies the payment signature
func verifyRazorpaySignature(orderID, paymentID, signature string) bool {
	_, keySecret := getRazorpayCredentials()
	
	// Create signature: HMAC_SHA256(order_id|payment_id, secret)
	message := fmt.Sprintf("%s|%s", orderID, paymentID)
	mac := hmac.New(sha256.New, []byte(keySecret))
	mac.Write([]byte(message))
	expectedSignature := hex.EncodeToString(mac.Sum(nil))
	
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// GetUserPayments returns all payments for the current user
func GetUserPayments(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userObjID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	filter := bson.M{"user_id": userObjID}

	// Allow admins to filter by user_id
	role, _ := c.Get("role")
	if role == models.RoleSuperAdmin || role == models.RoleStaff {
		if targetUserID := c.Query("user_id"); targetUserID != "" {
			if objID, err := primitive.ObjectIDFromHex(targetUserID); err == nil {
				filter["user_id"] = objID
			}
		}
	}

	// Filter by status
	if status := c.Query("status"); status != "" {
		filter["status"] = status
	}

	// Filter by type
	if paymentType := c.Query("type"); paymentType != "" {
		filter["type"] = paymentType
	}

	cursor, err := config.DB.Collection("payments").Find(
		context.Background(),
		filter,
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch payments"})
		return
	}
	defer cursor.Close(context.Background())

	var payments []models.Payment
	if err := cursor.All(context.Background(), &payments); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode payments"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"payments": payments,
		"count":    len(payments),
	})
}

// GetPaymentByID returns a specific payment by ID
func GetPaymentByID(c *gin.Context) {
	paymentID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment ID"})
		return
	}

	var payment models.Payment
	err = config.DB.Collection("payments").FindOne(context.Background(), bson.M{"_id": paymentID}).Decode(&payment)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Payment not found"})
		return
	}

	// Check if user has access to this payment
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")
	
	userObjID, _ := primitive.ObjectIDFromHex(userID.(string))
	if payment.UserID != userObjID && role != models.RoleSuperAdmin && role != models.RoleStaff {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	c.JSON(http.StatusOK, payment)
}

// GetAllPayments returns all payments (admin only)
func GetAllPayments(c *gin.Context) {
	filter := bson.M{}

	// Filter by status
	if status := c.Query("status"); status != "" {
		filter["status"] = status
	}

	// Filter by type
	if paymentType := c.Query("type"); paymentType != "" {
		filter["type"] = paymentType
	}

	// Pagination
	skip := 0
	limit := 100
	
	if skipStr := c.Query("skip"); skipStr != "" {
		if val, err := primitive.ParseDecimal128(skipStr); err == nil {
			if bi, _, err := val.BigInt(); err == nil {
				skip = int(bi.Int64())
			}
		}
	}
	
	if limitStr := c.Query("limit"); limitStr != "" {
		if val, err := primitive.ParseDecimal128(limitStr); err == nil {
			if bi, _, err := val.BigInt(); err == nil {
				limit = int(bi.Int64())
			}
		}
	}

	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "created_at", Value: -1}})
	findOptions.SetSkip(int64(skip))
	findOptions.SetLimit(int64(limit))

	cursor, err := config.DB.Collection("payments").Find(context.Background(), filter, findOptions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch payments"})
		return
	}
	defer cursor.Close(context.Background())

	var payments []models.Payment
	if err := cursor.All(context.Background(), &payments); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode payments"})
		return
	}

	// Get total count
	total, _ := config.DB.Collection("payments").CountDocuments(context.Background(), filter)

	c.JSON(http.StatusOK, gin.H{
		"payments": payments,
		"total":    total,
	})
}

// UpdatePaymentStatus updates payment status (admin only, typically for manual updates or webhooks)
func UpdatePaymentStatus(c *gin.Context) {
	paymentID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment ID"})
		return
	}

	var req struct {
		Status models.PaymentStatus `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	update := bson.M{
		"$set": bson.M{
			"status":     req.Status,
			"updated_at": time.Now(),
		},
	}

	result, err := config.DB.Collection("payments").UpdateOne(context.Background(), bson.M{"_id": paymentID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update payment"})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Payment not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Payment status updated"})
}

// HandleWebhook handles Razorpay webhook events
func HandleWebhook(c *gin.Context) {
	// Verify webhook signature (optional but recommended for production)
	// You should implement webhook signature verification here

	var webhookData map[string]interface{}
	if err := c.ShouldBindJSON(&webhookData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook payload"})
		return
	}

	// Log webhook for debugging
	fmt.Printf("Razorpay Webhook received: %v\n", webhookData)

	event := webhookData["event"].(string)
	payload := webhookData["payload"].(map[string]interface{})

	switch event {
	case "payment.captured":
		paymentEntity := payload["payment"].(map[string]interface{})["entity"].(map[string]interface{})
		orderID := paymentEntity["order_id"].(string)
		paymentID := paymentEntity["id"].(string)
		
		// Update payment record
		filter := bson.M{"razorpay_order_id": orderID}
		update := bson.M{
			"$set": bson.M{
				"status":              models.PaymentStatusPaid,
				"razorpay_payment_id": paymentID,
				"updated_at":          time.Now(),
				"paid_at":             time.Now(),
			},
		}
		config.DB.Collection("payments").UpdateOne(context.Background(), filter, update)

	case "payment.failed":
		paymentEntity := payload["payment"].(map[string]interface{})["entity"].(map[string]interface{})
		orderID := paymentEntity["order_id"].(string)
		failureReason := ""
		if reason, ok := paymentEntity["error_description"].(string); ok {
			failureReason = reason
		}
		
		filter := bson.M{"razorpay_order_id": orderID}
		update := bson.M{
			"$set": bson.M{
				"status":         models.PaymentStatusFailed,
				"failure_reason": failureReason,
				"updated_at":     time.Now(),
			},
		}
		config.DB.Collection("payments").UpdateOne(context.Background(), filter, update)

	case "refund.processed":
		refundEntity := payload["refund"].(map[string]interface{})["entity"].(map[string]interface{})
		paymentID := refundEntity["payment_id"].(string)
		refundID := refundEntity["id"].(string)
		amount := int(refundEntity["amount"].(float64))
		
		filter := bson.M{"razorpay_payment_id": paymentID}
		update := bson.M{
			"$set": bson.M{
				"status":          models.PaymentStatusRefunded,
				"refund_id":       refundID,
				"refunded_amount": amount,
				"updated_at":      time.Now(),
			},
		}
		config.DB.Collection("payments").UpdateOne(context.Background(), filter, update)
	}

	c.JSON(http.StatusOK, gin.H{"status": "received"})
}

// ProcessRefund processes a refund for a payment
func ProcessRefund(c *gin.Context) {
	var req models.RefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get payment record
	paymentID, err := primitive.ObjectIDFromHex(req.PaymentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment ID"})
		return
	}

	var payment models.Payment
	err = config.DB.Collection("payments").FindOne(context.Background(), bson.M{"_id": paymentID}).Decode(&payment)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Payment not found"})
		return
	}

	// Check if payment can be refunded
	if payment.Status != models.PaymentStatusPaid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payment cannot be refunded - not in paid status"})
		return
	}

	if payment.RazorpayPaymentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No Razorpay payment ID found"})
		return
	}

	// Determine refund amount
	refundAmount := req.Amount
	if refundAmount == 0 {
		refundAmount = payment.Amount // Full refund
	}

	// Create refund in Razorpay
	refundResult, err := createRazorpayRefund(payment.RazorpayPaymentID, refundAmount, req.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process refund: " + err.Error()})
		return
	}

	// Update payment record
	filter := bson.M{"_id": paymentID}
	update := bson.M{
		"$set": bson.M{
			"status":          models.PaymentStatusRefunded,
			"refund_id":       refundResult["id"],
			"refunded_amount": refundAmount,
			"updated_at":      time.Now(),
		},
	}

	_, err = config.DB.Collection("payments").UpdateOne(context.Background(), filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update payment record"})
		return
	}

	response := models.RefundResponse{
		RefundID:   refundResult["id"].(string),
		PaymentID:  req.PaymentID,
		Amount:     refundAmount,
		Status:     refundResult["status"].(string),
		ReceiptURL: "",
	}

	// Generate receipt URL if needed
	if receiptURL, ok := refundResult["receipt_url"].(string); ok {
		response.ReceiptURL = receiptURL
	}

	c.JSON(http.StatusOK, response)
}

// createRazorpayRefund creates a refund on Razorpay
func createRazorpayRefund(paymentID string, amount int, reason string) (map[string]interface{}, error) {
	keyID, keySecret := getRazorpayCredentials()

	payload := map[string]interface{}{
		"amount": amount,
	}
	if reason != "" {
		payload["notes"] = map[string]string{"reason": reason}
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("https://api.razorpay.com/v1/payments/%s/refund", paymentID), bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(keyID, keySecret)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("razorpay API error: %s", string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// GetPaymentByRequestID returns payment information for a room request
func GetPaymentByRequestID(c *gin.Context) {
	requestID := c.Param("request_id")
	
	reqObjID, err := primitive.ObjectIDFromHex(requestID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request ID"})
		return
	}

	var payment models.Payment
	err = config.DB.Collection("payments").FindOne(
		context.Background(),
		bson.M{"request_id": reqObjID},
		options.FindOne().SetSort(bson.D{{Key: "created_at", Value: -1}}),
	).Decode(&payment)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No payment found for this request"})
		return
	}

	c.JSON(http.StatusOK, payment)
}
