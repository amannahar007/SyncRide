package grpc

import (
	"context"
	"fmt"
	"log"

	pb "github.com/syncride/backend/shared/proto/notification"
)

type NotificationHandler struct {
	pb.UnimplementedNotificationServiceServer
	// In production, you would include:
	// fcmClient *messaging.Client
	// twilioClient *twilio.RestClient
}

func NewNotificationHandler() *NotificationHandler {
	return &NotificationHandler{}
}

func (h *NotificationHandler) SendPushNotification(ctx context.Context, req *pb.PushRequest) (*pb.NotificationResponse, error) {
	// ---------------------------------------------------------
	// MOCK FCM IMPLEMENTATION
	// ---------------------------------------------------------
	// In production, this looks like:
	// message := &messaging.Message{
	// 	 Notification: &messaging.Notification{Title: req.Title, Body: req.Body},
	// 	 Token: req.FcmToken,
	//   Data: req.Data,
	// }
	// response, err := h.fcmClient.Send(ctx, message)

	log.Printf("[FCM MOCK] Sending Push to User %s (Token: %s)", req.UserId, req.FcmToken)
	log.Printf("[FCM MOCK] Title: %s | Body: %s", req.Title, req.Body)

	if len(req.Data) > 0 {
		log.Printf("[FCM MOCK] Payload Data: %v", req.Data)
	}

	mockMessageID := fmt.Sprintf("projects/syncride-app/messages/mock_%s", req.UserId)

	return &pb.NotificationResponse{
		Success:           true,
		ProviderMessageId: mockMessageID,
	}, nil
}

func (h *NotificationHandler) SendSMS(ctx context.Context, req *pb.SMSRequest) (*pb.NotificationResponse, error) {
	// ---------------------------------------------------------
	// MOCK TWILIO IMPLEMENTATION
	// ---------------------------------------------------------
	// In production, this looks like:
	// params := &openapi.CreateMessageParams{}
	// params.SetTo(req.PhoneNumber) // e.g., "+919876543210"
	// params.SetFrom("+1234567890") // Your Twilio number
	// params.SetBody(req.Message)
	// resp, err := h.twilioClient.Api.CreateMessage(params)

	if req.PhoneNumber == "" {
		return &pb.NotificationResponse{
			Success:      false,
			ErrorMessage: "phone number cannot be empty",
		}, nil
	}

	log.Printf("[TWILIO MOCK] Sending SMS to %s", req.PhoneNumber)
	log.Printf("[TWILIO MOCK] Message: %s", req.Message)

	mockTwilioSID := fmt.Sprintf("SM_mock_sid_%s", req.PhoneNumber)

	return &pb.NotificationResponse{
		Success:           true,
		ProviderMessageId: mockTwilioSID,
	}, nil
}
