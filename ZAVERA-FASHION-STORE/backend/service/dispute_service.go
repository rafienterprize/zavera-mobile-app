package service

import (
	"database/sql"
	"fmt"
	"log"
	"time"
	"zavera/dto"
	"zavera/models"
	"zavera/repository"
)

type DisputeService interface {
	// CRUD
	CreateDispute(req *dto.CreateDisputeRequest, customerUserID *int) (*models.Dispute, error)
	GetDispute(id int) (*models.Dispute, error)
	GetDisputeByCode(code string) (*models.Dispute, error)
	GetDisputesByOrder(orderID int) ([]*models.Dispute, error)
	GetOpenDisputes() ([]*models.Dispute, error)

	// Status management
	StartInvestigation(disputeID int, investigatorID int) error
	RequestEvidence(disputeID int, message string, adminID int) error
	ResolveDispute(disputeID int, req *dto.ResolveDisputeRequest, adminID int) error
	CloseDispute(disputeID int, adminID int) error

	// Messages
	AddMessage(disputeID int, req *dto.AddDisputeMessageRequest, senderType string, senderID *int, senderName string) error
	GetMessages(disputeID int, includeInternal bool) ([]models.DisputeMessage, error)

	// Helpers
	ToDisputeResponse(dispute *models.Dispute) *dto.DisputeResponse
	SetFulfillmentService(fs FulfillmentService)
}

type disputeService struct {
	disputeRepo    repository.DisputeRepository
	orderRepo      repository.OrderRepository
	shippingRepo   repository.ShippingRepository
	refundService  RefundService
	fulfillmentSvc FulfillmentService
	db             *sql.DB
}

func NewDisputeService(
	disputeRepo repository.DisputeRepository,
	orderRepo repository.OrderRepository,
	shippingRepo repository.ShippingRepository,
	refundService RefundService,
	db *sql.DB,
) DisputeService {
	return &disputeService{
		disputeRepo:   disputeRepo,
		orderRepo:     orderRepo,
		shippingRepo:  shippingRepo,
		refundService: refundService,
		db:            db,
	}
}


// SetFulfillmentService sets the fulfillment service (to avoid circular dependency)
func (s *disputeService) SetFulfillmentService(fs FulfillmentService) {
	s.fulfillmentSvc = fs
}

func (s *disputeService) CreateDispute(req *dto.CreateDisputeRequest, customerUserID *int) (*models.Dispute, error) {
	order, err := s.orderRepo.FindByOrderCode(req.OrderCode)
	if err != nil {
		return nil, fmt.Errorf("order not found: %s", req.OrderCode)
	}

	dispute := &models.Dispute{
		DisputeCode:    repository.GenerateDisputeCode(),
		OrderID:        order.ID,
		ShipmentID:     req.ShipmentID,
		DisputeType:    models.DisputeType(req.DisputeType),
		Status:         models.DisputeStatusOpen,
		Title:          req.Title,
		Description:    req.Description,
		CustomerClaim:  req.CustomerClaim,
		CustomerUserID: customerUserID,
		CustomerEmail:  order.CustomerEmail,
		CustomerPhone:  order.CustomerPhone,
		EvidenceURLs:   req.EvidenceURLs,
	}

	if err := s.disputeRepo.Create(dispute); err != nil {
		return nil, err
	}

	msg := &models.DisputeMessage{
		DisputeID:  dispute.ID,
		SenderType: "system",
		SenderName: "System",
		Message:    fmt.Sprintf("Dispute opened: %s", req.Title),
	}
	s.disputeRepo.AddMessage(msg)

	if req.ShipmentID != nil {
		alert := &models.ShipmentAlert{
			ShipmentID:  *req.ShipmentID,
			AlertType:   "dispute",
			AlertLevel:  "critical",
			Title:       "Dispute Opened",
			Description: req.Title,
		}
		s.disputeRepo.CreateAlert(alert)
	}

	log.Printf("📋 Dispute created: %s for order %s", dispute.DisputeCode, req.OrderCode)
	return dispute, nil
}

func (s *disputeService) GetDispute(id int) (*models.Dispute, error) {
	dispute, err := s.disputeRepo.FindByID(id)
	if err != nil {
		return nil, ErrDisputeNotFound
	}
	messages, _ := s.disputeRepo.GetMessages(id, true)
	dispute.Messages = messages
	return dispute, nil
}

func (s *disputeService) GetDisputeByCode(code string) (*models.Dispute, error) {
	dispute, err := s.disputeRepo.FindByCode(code)
	if err != nil {
		return nil, ErrDisputeNotFound
	}
	messages, _ := s.disputeRepo.GetMessages(dispute.ID, true)
	dispute.Messages = messages
	return dispute, nil
}

func (s *disputeService) GetDisputesByOrder(orderID int) ([]*models.Dispute, error) {
	return s.disputeRepo.FindByOrderID(orderID)
}

func (s *disputeService) GetOpenDisputes() ([]*models.Dispute, error) {
	return s.disputeRepo.FindOpen()
}

func (s *disputeService) StartInvestigation(disputeID int, investigatorID int) error {
	dispute, err := s.disputeRepo.FindByID(disputeID)
	if err != nil {
		return ErrDisputeNotFound
	}

	if dispute.Status.IsFinalStatus() {
		return ErrDisputeAlreadyResolved
	}

	now := time.Now()
	dispute.Status = models.DisputeStatusInvestigating
	dispute.InvestigatorID = &investigatorID
	dispute.InvestigationStartedAt = &now

	if err := s.disputeRepo.Update(dispute); err != nil {
		return err
	}

	msg := &models.DisputeMessage{
		DisputeID:  disputeID,
		SenderType: "system",
		SenderName: "System",
		Message:    "Investigation started",
	}
	s.disputeRepo.AddMessage(msg)
	return nil
}

func (s *disputeService) RequestEvidence(disputeID int, message string, adminID int) error {
	dispute, err := s.disputeRepo.FindByID(disputeID)
	if err != nil {
		return ErrDisputeNotFound
	}

	if dispute.Status.IsFinalStatus() {
		return ErrDisputeAlreadyResolved
	}

	s.disputeRepo.UpdateStatus(disputeID, models.DisputeStatusEvidenceRequired)

	msg := &models.DisputeMessage{
		DisputeID:  disputeID,
		SenderType: "admin",
		SenderID:   &adminID,
		Message:    message,
	}
	s.disputeRepo.AddMessage(msg)
	return nil
}


func (s *disputeService) ResolveDispute(disputeID int, req *dto.ResolveDisputeRequest, adminID int) error {
	dispute, err := s.disputeRepo.FindByID(disputeID)
	if err != nil {
		return ErrDisputeNotFound
	}

	if dispute.Status.IsFinalStatus() {
		return ErrDisputeAlreadyResolved
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	resolution := models.DisputeStatus(req.Resolution)
	now := time.Now()

	query := `
		UPDATE disputes SET
			status = $1, resolution = $1, resolution_notes = $2,
			resolution_amount = $3, resolved_by = $4, resolved_at = $5,
			investigation_completed_at = $5, updated_at = NOW()
		WHERE id = $6
	`
	_, err = tx.Exec(query, resolution, req.ResolutionNotes, req.ResolutionAmount, adminID, now, disputeID)
	if err != nil {
		return err
	}

	switch resolution {
	case models.DisputeStatusResolvedRefund:
		if req.CreateRefund && s.refundService != nil {
			order, _ := s.orderRepo.FindByID(dispute.OrderID)
			if order != nil {
				refundReq := &dto.RefundRequest{
					OrderCode:      order.OrderCode,
					RefundType:     "FULL",
					Reason:         fmt.Sprintf("Dispute resolution: %s", req.ResolutionNotes),
					IdempotencyKey: fmt.Sprintf("dispute-%d-refund", disputeID),
				}
				refund, err := s.refundService.CreateRefund(refundReq, &adminID)
				if err == nil {
					s.disputeRepo.LinkRefund(disputeID, refund.ID)
				}
			}
		}

	case models.DisputeStatusResolvedReship:
		if req.CreateReship && s.fulfillmentSvc != nil && dispute.ShipmentID != nil {
			reshipReq := &dto.ReshipRequest{
				Reason:     fmt.Sprintf("Dispute resolution: %s", req.ResolutionNotes),
				CostBearer: "company",
			}
			newShipment, err := s.fulfillmentSvc.CreateReship(*dispute.ShipmentID, reshipReq, fmt.Sprintf("admin:%d", adminID))
			if err == nil {
				s.disputeRepo.LinkReship(disputeID, newShipment.ID)
			}
		}
	}

	msg := &models.DisputeMessage{
		DisputeID:  disputeID,
		SenderType: "system",
		SenderName: "System",
		Message:    fmt.Sprintf("Dispute resolved: %s - %s", resolution, req.ResolutionNotes),
	}
	s.disputeRepo.AddMessage(msg)

	if dispute.ShipmentID != nil {
		s.db.Exec(`UPDATE shipment_alerts SET resolved = true, resolution_notes = 'Dispute resolved' WHERE shipment_id = $1 AND alert_type = 'dispute' AND resolved = false`, *dispute.ShipmentID)
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	log.Printf("✅ Dispute %s resolved as %s by admin %d", dispute.DisputeCode, resolution, adminID)
	return nil
}

func (s *disputeService) CloseDispute(disputeID int, adminID int) error {
	dispute, err := s.disputeRepo.FindByID(disputeID)
	if err != nil {
		return ErrDisputeNotFound
	}

	if !dispute.Status.IsResolved() {
		return fmt.Errorf("can only close resolved disputes, current status: %s", dispute.Status)
	}

	if err := s.disputeRepo.UpdateStatus(disputeID, models.DisputeStatusClosed); err != nil {
		return err
	}

	msg := &models.DisputeMessage{
		DisputeID:  disputeID,
		SenderType: "system",
		SenderName: "System",
		Message:    "Dispute closed",
	}
	s.disputeRepo.AddMessage(msg)

	log.Printf("📋 Dispute %s closed by admin %d", dispute.DisputeCode, adminID)
	return nil
}

func (s *disputeService) AddMessage(disputeID int, req *dto.AddDisputeMessageRequest, senderType string, senderID *int, senderName string) error {
	dispute, err := s.disputeRepo.FindByID(disputeID)
	if err != nil {
		return ErrDisputeNotFound
	}

	if dispute.Status.IsFinalStatus() {
		return fmt.Errorf("cannot add message to closed dispute")
	}

	msg := &models.DisputeMessage{
		DisputeID:      disputeID,
		SenderType:     senderType,
		SenderID:       senderID,
		SenderName:     senderName,
		Message:        req.Message,
		AttachmentURLs: req.AttachmentURLs,
		IsInternal:     req.IsInternal,
	}

	return s.disputeRepo.AddMessage(msg)
}

func (s *disputeService) GetMessages(disputeID int, includeInternal bool) ([]models.DisputeMessage, error) {
	_, err := s.disputeRepo.FindByID(disputeID)
	if err != nil {
		return nil, ErrDisputeNotFound
	}
	return s.disputeRepo.GetMessages(disputeID, includeInternal)
}


func (s *disputeService) ToDisputeResponse(dispute *models.Dispute) *dto.DisputeResponse {
	resp := &dto.DisputeResponse{
		ID:                 dispute.ID,
		DisputeCode:        dispute.DisputeCode,
		OrderID:            dispute.OrderID,
		ShipmentID:         dispute.ShipmentID,
		DisputeType:        string(dispute.DisputeType),
		Status:             string(dispute.Status),
		Title:              dispute.Title,
		Description:        dispute.Description,
		CustomerClaim:      dispute.CustomerClaim,
		CustomerEmail:      dispute.CustomerEmail,
		EvidenceURLs:       dispute.EvidenceURLs,
		InvestigationNotes: dispute.InvestigationNotes,
		Resolution:         string(dispute.Resolution),
		ResolutionNotes:    dispute.ResolutionNotes,
		ResolutionAmount:   dispute.ResolutionAmount,
		ReshipShipmentID:   dispute.ReshipShipmentID,
		CreatedAt:          dispute.CreatedAt.Format(time.RFC3339),
	}

	order, _ := s.orderRepo.FindByID(dispute.OrderID)
	if order != nil {
		resp.OrderCode = order.OrderCode
	}

	if dispute.ResponseDeadline != nil {
		t := dispute.ResponseDeadline.Format(time.RFC3339)
		resp.ResponseDeadline = &t
	}
	if dispute.ResolutionDeadline != nil {
		t := dispute.ResolutionDeadline.Format(time.RFC3339)
		resp.ResolutionDeadline = &t
	}
	if dispute.ResolvedAt != nil {
		t := dispute.ResolvedAt.Format(time.RFC3339)
		resp.ResolvedAt = &t
	}

	for _, m := range dispute.Messages {
		resp.Messages = append(resp.Messages, dto.DisputeMessageResponse{
			ID:             m.ID,
			SenderType:     m.SenderType,
			SenderName:     m.SenderName,
			Message:        m.Message,
			AttachmentURLs: m.AttachmentURLs,
			IsInternal:     m.IsInternal,
			CreatedAt:      m.CreatedAt.Format(time.RFC3339),
		})
	}

	return resp
}
