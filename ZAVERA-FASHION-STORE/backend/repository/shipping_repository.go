package repository

import (
	"database/sql"
	"encoding/json"
	"zavera/models"
)

type ShippingRepository interface {
	// Providers
	GetActiveProviders() ([]models.ShippingProvider, error)
	GetProviderByCode(code string) (*models.ShippingProvider, error)

	// Shipments
	CreateShipment(shipment *models.Shipment) error
	GetShipmentByOrderID(orderID int) (*models.Shipment, error)
	FindByOrderID(orderID int) (*models.Shipment, error) // Alias for GetShipmentByOrderID
	GetShipmentByID(id int) (*models.Shipment, error)
	GetShipmentByResi(resi string) (*models.Shipment, error)
	UpdateShipmentTracking(id int, trackingNumber string) error
	UpdateShipmentStatus(id int, status models.ShipmentStatus) error
	MarkShipmentShipped(id int, trackingNumber string) error
	MarkShipmentDelivered(id int) error
	GetShipmentsForTracking() ([]models.Shipment, error)
	
	// Biteship integration
	UpdateShipmentBiteshipIDs(id int, draftOrderID, orderID, trackingID, waybillID string) error
	UpdateBiteshipTracking(id int, trackingID, waybillID string) error
	GetShipmentByBiteshipTrackingID(trackingID string) (*models.Shipment, error)

	// Shipping Snapshots
	CreateShippingSnapshot(snapshot *models.ShippingSnapshot) error
	GetShippingSnapshotByOrderID(orderID int) (*models.ShippingSnapshot, error)

	// Tracking History
	AddTrackingEvent(event *models.TrackingEvent) error
	GetTrackingHistory(shipmentID int) ([]models.TrackingEvent, error)

	// User Addresses
	CreateAddress(address *models.UserAddress) error
	GetAddressByID(id int) (*models.UserAddress, error)
	GetUserAddresses(userID int) ([]models.UserAddress, error)
	UpdateAddress(address *models.UserAddress) error
	DeleteAddress(id int) error
	SetDefaultAddress(userID, addressID int) error
	GetDefaultAddress(userID int) (*models.UserAddress, error)

	// Subdistricts (Kecamatan)
	GetSubdistrictsByCityID(cityID string) ([]models.Subdistrict, error)
	GetSubdistrictByID(id int) (*models.Subdistrict, error)
	SearchSubdistricts(cityID string, query string) ([]models.Subdistrict, error)
	
	// Biteship Locations
	CreateBiteshipLocation(location *models.BiteshipLocation) error
	GetBiteshipLocationsByUserID(userID int) ([]models.BiteshipLocation, error)
	GetBiteshipLocationByID(id int) (*models.BiteshipLocation, error)
}

type shippingRepository struct {
	db *sql.DB
}

func NewShippingRepository(db *sql.DB) ShippingRepository {
	return &shippingRepository{db: db}
}

// ============================================
// PROVIDERS
// ============================================

func (r *shippingRepository) GetActiveProviders() ([]models.ShippingProvider, error) {
	query := `
		SELECT id, name, code, logo_url, is_active, created_at, updated_at
		FROM shipping_providers
		WHERE is_active = true
		ORDER BY name
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var providers []models.ShippingProvider
	for rows.Next() {
		var p models.ShippingProvider
		err := rows.Scan(&p.ID, &p.Name, &p.Code, &p.LogoURL, &p.IsActive, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			continue
		}
		providers = append(providers, p)
	}

	return providers, nil
}

func (r *shippingRepository) GetProviderByCode(code string) (*models.ShippingProvider, error) {
	query := `
		SELECT id, name, code, logo_url, is_active, created_at, updated_at
		FROM shipping_providers
		WHERE code = $1
	`

	var p models.ShippingProvider
	err := r.db.QueryRow(query, code).Scan(
		&p.ID, &p.Name, &p.Code, &p.LogoURL, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

// ============================================
// SHIPMENTS
// ============================================

func (r *shippingRepository) CreateShipment(shipment *models.Shipment) error {
	query := `
		INSERT INTO shipments (
			order_id, provider_code, provider_name, service_code, service_name,
			cost, etd, weight, status, origin_city_id, origin_city_name,
			destination_city_id, destination_city_name
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, created_at, updated_at
	`

	return r.db.QueryRow(
		query,
		shipment.OrderID, shipment.ProviderCode, shipment.ProviderName,
		shipment.ServiceCode, shipment.ServiceName, shipment.Cost, shipment.ETD,
		shipment.Weight, shipment.Status, shipment.OriginCityID, shipment.OriginCityName,
		shipment.DestinationCityID, shipment.DestinationCityName,
	).Scan(&shipment.ID, &shipment.CreatedAt, &shipment.UpdatedAt)
}

// FindByOrderID is an alias for GetShipmentByOrderID
func (r *shippingRepository) FindByOrderID(orderID int) (*models.Shipment, error) {
	return r.GetShipmentByOrderID(orderID)
}

func (r *shippingRepository) GetShipmentByOrderID(orderID int) (*models.Shipment, error) {
	query := `
		SELECT id, order_id, provider_code, provider_name, service_code, service_name,
		       cost, etd, weight, tracking_number, status, origin_city_id, origin_city_name,
		       destination_city_id, destination_city_name, shipped_at, delivered_at,
		       created_at, updated_at,
		       biteship_draft_order_id, biteship_order_id, biteship_tracking_id, biteship_waybill_id
		FROM shipments
		WHERE order_id = $1
	`

	var s models.Shipment
	var trackingNumber sql.NullString
	var shippedAt, deliveredAt sql.NullTime
	var biteshipDraftOrderID, biteshipOrderID, biteshipTrackingID, biteshipWaybillID sql.NullString

	err := r.db.QueryRow(query, orderID).Scan(
		&s.ID, &s.OrderID, &s.ProviderCode, &s.ProviderName, &s.ServiceCode, &s.ServiceName,
		&s.Cost, &s.ETD, &s.Weight, &trackingNumber, &s.Status, &s.OriginCityID, &s.OriginCityName,
		&s.DestinationCityID, &s.DestinationCityName, &shippedAt, &deliveredAt,
		&s.CreatedAt, &s.UpdatedAt,
		&biteshipDraftOrderID, &biteshipOrderID, &biteshipTrackingID, &biteshipWaybillID,
	)
	if err != nil {
		return nil, err
	}

	if trackingNumber.Valid {
		s.TrackingNumber = trackingNumber.String
	}
	if shippedAt.Valid {
		s.ShippedAt = &shippedAt.Time
	}
	if deliveredAt.Valid {
		s.DeliveredAt = &deliveredAt.Time
	}
	if biteshipDraftOrderID.Valid {
		s.BiteshipDraftOrderID = biteshipDraftOrderID.String
	}
	if biteshipOrderID.Valid {
		s.BiteshipOrderID = biteshipOrderID.String
	}
	if biteshipTrackingID.Valid {
		s.BiteshipTrackingID = biteshipTrackingID.String
	}
	if biteshipWaybillID.Valid {
		s.BiteshipWaybillID = biteshipWaybillID.String
	}

	// Load tracking history
	history, _ := r.GetTrackingHistory(s.ID)
	s.TrackingHistory = history

	return &s, nil
}

func (r *shippingRepository) GetShipmentByID(id int) (*models.Shipment, error) {
	query := `
		SELECT id, order_id, provider_code, provider_name, service_code, service_name,
		       cost, etd, weight, tracking_number, status, origin_city_id, origin_city_name,
		       destination_city_id, destination_city_name, shipped_at, delivered_at,
		       created_at, updated_at
		FROM shipments
		WHERE id = $1
	`

	var s models.Shipment
	var trackingNumber sql.NullString
	var shippedAt, deliveredAt sql.NullTime

	err := r.db.QueryRow(query, id).Scan(
		&s.ID, &s.OrderID, &s.ProviderCode, &s.ProviderName, &s.ServiceCode, &s.ServiceName,
		&s.Cost, &s.ETD, &s.Weight, &trackingNumber, &s.Status, &s.OriginCityID, &s.OriginCityName,
		&s.DestinationCityID, &s.DestinationCityName, &shippedAt, &deliveredAt,
		&s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if trackingNumber.Valid {
		s.TrackingNumber = trackingNumber.String
	}
	if shippedAt.Valid {
		s.ShippedAt = &shippedAt.Time
	}
	if deliveredAt.Valid {
		s.DeliveredAt = &deliveredAt.Time
	}

	return &s, nil
}

func (r *shippingRepository) GetShipmentByResi(resi string) (*models.Shipment, error) {
	query := `
		SELECT id, order_id, provider_code, provider_name, service_code, service_name,
		       cost, etd, weight, tracking_number, status, origin_city_id, origin_city_name,
		       destination_city_id, destination_city_name, shipped_at, delivered_at,
		       created_at, updated_at
		FROM shipments
		WHERE tracking_number = $1
	`

	var s models.Shipment
	var trackingNumber sql.NullString
	var shippedAt, deliveredAt sql.NullTime

	err := r.db.QueryRow(query, resi).Scan(
		&s.ID, &s.OrderID, &s.ProviderCode, &s.ProviderName, &s.ServiceCode, &s.ServiceName,
		&s.Cost, &s.ETD, &s.Weight, &trackingNumber, &s.Status, &s.OriginCityID, &s.OriginCityName,
		&s.DestinationCityID, &s.DestinationCityName, &shippedAt, &deliveredAt,
		&s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if trackingNumber.Valid {
		s.TrackingNumber = trackingNumber.String
	}
	if shippedAt.Valid {
		s.ShippedAt = &shippedAt.Time
	}
	if deliveredAt.Valid {
		s.DeliveredAt = &deliveredAt.Time
	}

	// Load tracking history
	history, _ := r.GetTrackingHistory(s.ID)
	s.TrackingHistory = history

	return &s, nil
}

func (r *shippingRepository) UpdateShipmentTracking(id int, trackingNumber string) error {
	query := `
		UPDATE shipments
		SET tracking_number = $1, updated_at = NOW()
		WHERE id = $2
	`
	_, err := r.db.Exec(query, trackingNumber, id)
	return err
}

func (r *shippingRepository) UpdateShipmentStatus(id int, status models.ShipmentStatus) error {
	query := `
		UPDATE shipments
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`
	_, err := r.db.Exec(query, status, id)
	return err
}

func (r *shippingRepository) MarkShipmentShipped(id int, trackingNumber string) error {
	query := `
		UPDATE shipments
		SET tracking_number = $1, status = 'SHIPPED', shipped_at = NOW(), updated_at = NOW()
		WHERE id = $2
	`
	_, err := r.db.Exec(query, trackingNumber, id)
	return err
}

func (r *shippingRepository) MarkShipmentDelivered(id int) error {
	query := `
		UPDATE shipments
		SET status = 'DELIVERED', delivered_at = NOW(), updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.Exec(query, id)
	return err
}

// ============================================
// SHIPPING SNAPSHOTS
// ============================================

// CreateShippingSnapshot stores the Biteship response at checkout time
func (r *shippingRepository) CreateShippingSnapshot(snapshot *models.ShippingSnapshot) error {
	rawJSON, err := json.Marshal(snapshot.BiteshipRawJSON)
	if err != nil {
		rawJSON = []byte("{}")
	}

	query := `
		INSERT INTO shipping_snapshots (
			order_id, courier, service, cost, etd, origin_city_id, origin_city_name,
			destination_city_id, destination_city_name, destination_district_id, weight,
			origin_area_id, origin_area_name, destination_area_id, destination_area_name,
			biteship_raw_json
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		ON CONFLICT (order_id) DO UPDATE SET
			courier = EXCLUDED.courier,
			service = EXCLUDED.service,
			cost = EXCLUDED.cost,
			etd = EXCLUDED.etd,
			origin_city_id = EXCLUDED.origin_city_id,
			origin_city_name = EXCLUDED.origin_city_name,
			destination_city_id = EXCLUDED.destination_city_id,
			destination_city_name = EXCLUDED.destination_city_name,
			destination_district_id = EXCLUDED.destination_district_id,
			weight = EXCLUDED.weight,
			origin_area_id = EXCLUDED.origin_area_id,
			origin_area_name = EXCLUDED.origin_area_name,
			destination_area_id = EXCLUDED.destination_area_id,
			destination_area_name = EXCLUDED.destination_area_name,
			biteship_raw_json = EXCLUDED.biteship_raw_json
		RETURNING id, created_at
	`

	return r.db.QueryRow(
		query,
		snapshot.OrderID, snapshot.Courier, snapshot.Service, snapshot.Cost, snapshot.ETD,
		snapshot.OriginCityID, snapshot.OriginCityName, snapshot.DestinationCityID,
		snapshot.DestinationCityName, snapshot.DestinationDistrictID, snapshot.Weight,
		snapshot.OriginAreaID, snapshot.OriginAreaName, snapshot.DestinationAreaID,
		snapshot.DestinationAreaName, rawJSON,
	).Scan(&snapshot.ID, &snapshot.CreatedAt)
}

// GetShippingSnapshotByOrderID retrieves the shipping snapshot for an order
func (r *shippingRepository) GetShippingSnapshotByOrderID(orderID int) (*models.ShippingSnapshot, error) {
	query := `
		SELECT id, order_id, courier, service, cost, etd, origin_city_id, 
		       COALESCE(origin_city_name, '') as origin_city_name,
		       destination_city_id, COALESCE(destination_city_name, '') as destination_city_name,
		       COALESCE(destination_district_id, '') as destination_district_id,
		       weight,
		       COALESCE(origin_area_id, '') as origin_area_id,
		       COALESCE(origin_area_name, '') as origin_area_name,
		       COALESCE(destination_area_id, '') as destination_area_id,
		       COALESCE(destination_area_name, '') as destination_area_name,
		       COALESCE(biteship_raw_json, '{}') as biteship_raw_json, created_at
		FROM shipping_snapshots
		WHERE order_id = $1
	`

	var snapshot models.ShippingSnapshot
	var rawJSON []byte

	err := r.db.QueryRow(query, orderID).Scan(
		&snapshot.ID, &snapshot.OrderID, &snapshot.Courier, &snapshot.Service,
		&snapshot.Cost, &snapshot.ETD, &snapshot.OriginCityID, &snapshot.OriginCityName,
		&snapshot.DestinationCityID, &snapshot.DestinationCityName, &snapshot.DestinationDistrictID,
		&snapshot.Weight, &snapshot.OriginAreaID, &snapshot.OriginAreaName,
		&snapshot.DestinationAreaID, &snapshot.DestinationAreaName,
		&rawJSON, &snapshot.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if len(rawJSON) > 0 {
		json.Unmarshal(rawJSON, &snapshot.BiteshipRawJSON)
	}

	return &snapshot, nil
}

// GetShipmentsForTracking returns shipments that need tracking updates
func (r *shippingRepository) GetShipmentsForTracking() ([]models.Shipment, error) {
	query := `
		SELECT id, order_id, provider_code, provider_name, service_code, service_name,
		       cost, etd, weight, tracking_number, status, origin_city_id, origin_city_name,
		       destination_city_id, destination_city_name,
		       COALESCE(biteship_draft_order_id, '') as biteship_draft_order_id,
		       COALESCE(biteship_order_id, '') as biteship_order_id,
		       COALESCE(biteship_tracking_id, '') as biteship_tracking_id,
		       COALESCE(biteship_waybill_id, '') as biteship_waybill_id,
		       shipped_at, delivered_at, created_at, updated_at
		FROM shipments
		WHERE status IN ('SHIPPED', 'IN_TRANSIT', 'OUT_FOR_DELIVERY')
		  AND (
		      (tracking_number IS NOT NULL AND tracking_number != '')
		      OR (biteship_tracking_id IS NOT NULL AND biteship_tracking_id != '')
		  )
		ORDER BY updated_at ASC
		LIMIT 50
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shipments []models.Shipment
	for rows.Next() {
		var s models.Shipment
		var trackingNumber sql.NullString
		var shippedAt, deliveredAt sql.NullTime

		err := rows.Scan(
			&s.ID, &s.OrderID, &s.ProviderCode, &s.ProviderName, &s.ServiceCode, &s.ServiceName,
			&s.Cost, &s.ETD, &s.Weight, &trackingNumber, &s.Status, &s.OriginCityID, &s.OriginCityName,
			&s.DestinationCityID, &s.DestinationCityName,
			&s.BiteshipDraftOrderID, &s.BiteshipOrderID, &s.BiteshipTrackingID, &s.BiteshipWaybillID,
			&shippedAt, &deliveredAt, &s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			continue
		}

		if trackingNumber.Valid {
			s.TrackingNumber = trackingNumber.String
		}
		if shippedAt.Valid {
			s.ShippedAt = &shippedAt.Time
		}
		if deliveredAt.Valid {
			s.DeliveredAt = &deliveredAt.Time
		}

		shipments = append(shipments, s)
	}

	return shipments, nil
}

// ============================================
// TRACKING HISTORY
// ============================================

func (r *shippingRepository) AddTrackingEvent(event *models.TrackingEvent) error {
	rawDataJSON, _ := json.Marshal(event.RawData)

	query := `
		INSERT INTO shipment_tracking_history (
			shipment_id, status, description, location, event_time, raw_data
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`

	return r.db.QueryRow(
		query,
		event.ShipmentID, event.Status, event.Description, event.Location,
		event.EventTime, rawDataJSON,
	).Scan(&event.ID, &event.CreatedAt)
}

func (r *shippingRepository) GetTrackingHistory(shipmentID int) ([]models.TrackingEvent, error) {
	query := `
		SELECT id, shipment_id, status, description, location, event_time, raw_data, created_at
		FROM shipment_tracking_history
		WHERE shipment_id = $1
		ORDER BY event_time DESC, created_at DESC
	`

	rows, err := r.db.Query(query, shipmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.TrackingEvent
	for rows.Next() {
		var e models.TrackingEvent
		var rawDataJSON []byte
		var eventTime sql.NullTime

		err := rows.Scan(
			&e.ID, &e.ShipmentID, &e.Status, &e.Description, &e.Location,
			&eventTime, &rawDataJSON, &e.CreatedAt,
		)
		if err != nil {
			continue
		}

		if eventTime.Valid {
			e.EventTime = &eventTime.Time
		}
		if len(rawDataJSON) > 0 {
			json.Unmarshal(rawDataJSON, &e.RawData)
		}

		events = append(events, e)
	}

	return events, nil
}

// ============================================
// USER ADDRESSES
// ============================================

func (r *shippingRepository) CreateAddress(address *models.UserAddress) error {
	// If setting as default, unset other defaults first
	if address.IsDefault && address.UserID != nil {
		r.db.Exec(`UPDATE user_addresses SET is_default = false WHERE user_id = $1`, *address.UserID)
	}

	query := `
		INSERT INTO user_addresses (
			user_id, label, recipient_name, phone, province_id, province_name,
			city_id, city_name, district_id, district, subdistrict, postal_code, full_address,
			is_default, is_active, area_id, area_name
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, true, $15, $16)
		RETURNING id, created_at, updated_at
	`

	return r.db.QueryRow(
		query,
		address.UserID, address.Label, address.RecipientName, address.Phone,
		address.ProvinceID, address.ProvinceName, address.CityID, address.CityName,
		address.DistrictID, address.District, address.Subdistrict, address.PostalCode, address.FullAddress,
		address.IsDefault, address.AreaID, address.AreaName,
	).Scan(&address.ID, &address.CreatedAt, &address.UpdatedAt)
}

func (r *shippingRepository) GetAddressByID(id int) (*models.UserAddress, error) {
	query := `
		SELECT id, user_id, label, recipient_name, phone, province_id, province_name,
		       city_id, city_name, COALESCE(district_id, ''), district, subdistrict, postal_code, full_address,
		       is_default, is_active, created_at, updated_at, COALESCE(area_id, ''), COALESCE(area_name, '')
		FROM user_addresses
		WHERE id = $1 AND is_active = true
	`

	var a models.UserAddress
	err := r.db.QueryRow(query, id).Scan(
		&a.ID, &a.UserID, &a.Label, &a.RecipientName, &a.Phone,
		&a.ProvinceID, &a.ProvinceName, &a.CityID, &a.CityName,
		&a.DistrictID, &a.District, &a.Subdistrict, &a.PostalCode, &a.FullAddress,
		&a.IsDefault, &a.IsActive, &a.CreatedAt, &a.UpdatedAt, &a.AreaID, &a.AreaName,
	)
	if err != nil {
		return nil, err
	}

	return &a, nil
}

func (r *shippingRepository) GetUserAddresses(userID int) ([]models.UserAddress, error) {
	query := `
		SELECT id, user_id, label, recipient_name, phone, province_id, province_name,
		       city_id, city_name, COALESCE(district_id, ''), district, subdistrict, postal_code, full_address,
		       is_default, is_active, created_at, updated_at, COALESCE(area_id, ''), COALESCE(area_name, '')
		FROM user_addresses
		WHERE user_id = $1 AND is_active = true
		ORDER BY is_default DESC, created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var addresses []models.UserAddress
	for rows.Next() {
		var a models.UserAddress
		err := rows.Scan(
			&a.ID, &a.UserID, &a.Label, &a.RecipientName, &a.Phone,
			&a.ProvinceID, &a.ProvinceName, &a.CityID, &a.CityName,
			&a.DistrictID, &a.District, &a.Subdistrict, &a.PostalCode, &a.FullAddress,
			&a.IsDefault, &a.IsActive, &a.CreatedAt, &a.UpdatedAt, &a.AreaID, &a.AreaName,
		)
		if err != nil {
			continue
		}
		addresses = append(addresses, a)
	}

	return addresses, nil
}

func (r *shippingRepository) UpdateAddress(address *models.UserAddress) error {
	// If setting as default, unset other defaults first
	if address.IsDefault && address.UserID != nil {
		r.db.Exec(`UPDATE user_addresses SET is_default = false WHERE user_id = $1 AND id != $2`,
			*address.UserID, address.ID)
	}

	query := `
		UPDATE user_addresses
		SET label = $1, recipient_name = $2, phone = $3, province_id = $4, province_name = $5,
		    city_id = $6, city_name = $7, district_id = $8, district = $9, subdistrict = $10, postal_code = $11,
		    full_address = $12, is_default = $13, area_id = $15, area_name = $16, updated_at = NOW()
		WHERE id = $14
	`

	_, err := r.db.Exec(
		query,
		address.Label, address.RecipientName, address.Phone, address.ProvinceID, address.ProvinceName,
		address.CityID, address.CityName, address.DistrictID, address.District, address.Subdistrict, address.PostalCode,
		address.FullAddress, address.IsDefault, address.ID, address.AreaID, address.AreaName,
	)
	return err
}

func (r *shippingRepository) DeleteAddress(id int) error {
	// Soft delete
	query := `UPDATE user_addresses SET is_active = false, updated_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *shippingRepository) SetDefaultAddress(userID, addressID int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Unset all defaults
	_, err = tx.Exec(`UPDATE user_addresses SET is_default = false WHERE user_id = $1`, userID)
	if err != nil {
		return err
	}

	// Set new default
	_, err = tx.Exec(`UPDATE user_addresses SET is_default = true WHERE id = $1 AND user_id = $2`, addressID, userID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *shippingRepository) GetDefaultAddress(userID int) (*models.UserAddress, error) {
	query := `
		SELECT id, user_id, label, recipient_name, phone, province_id, province_name,
		       city_id, city_name, COALESCE(district_id, ''), district, subdistrict, postal_code, full_address,
		       is_default, is_active, created_at, updated_at
		FROM user_addresses
		WHERE user_id = $1 AND is_default = true AND is_active = true
		LIMIT 1
	`

	var a models.UserAddress
	err := r.db.QueryRow(query, userID).Scan(
		&a.ID, &a.UserID, &a.Label, &a.RecipientName, &a.Phone,
		&a.ProvinceID, &a.ProvinceName, &a.CityID, &a.CityName,
		&a.DistrictID, &a.District, &a.Subdistrict, &a.PostalCode, &a.FullAddress,
		&a.IsDefault, &a.IsActive, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &a, nil
}

// ============================================
// SUBDISTRICTS (KECAMATAN)
// ============================================

func (r *shippingRepository) GetSubdistrictsByCityID(cityID string) ([]models.Subdistrict, error) {
	query := `
		SELECT id, city_id, name, postal_codes, created_at
		FROM subdistricts
		WHERE city_id = $1
		ORDER BY name
	`

	rows, err := r.db.Query(query, cityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subdistricts []models.Subdistrict
	for rows.Next() {
		var s models.Subdistrict
		var postalCodesArray []byte
		err := rows.Scan(&s.ID, &s.CityID, &s.Name, &postalCodesArray, &s.CreatedAt)
		if err != nil {
			continue
		}
		// Parse PostgreSQL array to Go slice
		if postalCodesArray != nil {
			// Handle PostgreSQL array format: {val1,val2,val3}
			str := string(postalCodesArray)
			if len(str) > 2 && str[0] == '{' && str[len(str)-1] == '}' {
				str = str[1 : len(str)-1]
				if str != "" {
					s.PostalCodes = splitPostalCodes(str)
				}
			}
		}
		subdistricts = append(subdistricts, s)
	}

	return subdistricts, nil
}

func (r *shippingRepository) GetSubdistrictByID(id int) (*models.Subdistrict, error) {
	query := `
		SELECT id, city_id, name, postal_codes, created_at
		FROM subdistricts
		WHERE id = $1
	`

	var s models.Subdistrict
	var postalCodesArray []byte
	err := r.db.QueryRow(query, id).Scan(&s.ID, &s.CityID, &s.Name, &postalCodesArray, &s.CreatedAt)
	if err != nil {
		return nil, err
	}

	if postalCodesArray != nil {
		str := string(postalCodesArray)
		if len(str) > 2 && str[0] == '{' && str[len(str)-1] == '}' {
			str = str[1 : len(str)-1]
			if str != "" {
				s.PostalCodes = splitPostalCodes(str)
			}
		}
	}

	return &s, nil
}

func (r *shippingRepository) SearchSubdistricts(cityID string, query string) ([]models.Subdistrict, error) {
	sqlQuery := `
		SELECT id, city_id, name, postal_codes, created_at
		FROM subdistricts
		WHERE city_id = $1 AND LOWER(name) LIKE LOWER($2)
		ORDER BY name
		LIMIT 20
	`

	rows, err := r.db.Query(sqlQuery, cityID, "%"+query+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subdistricts []models.Subdistrict
	for rows.Next() {
		var s models.Subdistrict
		var postalCodesArray []byte
		err := rows.Scan(&s.ID, &s.CityID, &s.Name, &postalCodesArray, &s.CreatedAt)
		if err != nil {
			continue
		}
		if postalCodesArray != nil {
			str := string(postalCodesArray)
			if len(str) > 2 && str[0] == '{' && str[len(str)-1] == '}' {
				str = str[1 : len(str)-1]
				if str != "" {
					s.PostalCodes = splitPostalCodes(str)
				}
			}
		}
		subdistricts = append(subdistricts, s)
	}

	return subdistricts, nil
}

// Helper function to split PostgreSQL array string
func splitPostalCodes(s string) []string {
	var result []string
	var current string
	inQuote := false

	for _, c := range s {
		switch c {
		case '"':
			inQuote = !inQuote
		case ',':
			if !inQuote {
				if current != "" {
					result = append(result, current)
				}
				current = ""
			} else {
				current += string(c)
			}
		default:
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}


// ============================================
// BITESHIP INTEGRATION
// ============================================

// UpdateShipmentBiteshipIDs updates Biteship-related IDs for a shipment
func (r *shippingRepository) UpdateShipmentBiteshipIDs(id int, draftOrderID, orderID, trackingID, waybillID string) error {
	query := `
		UPDATE shipments
		SET biteship_draft_order_id = COALESCE(NULLIF($1, ''), biteship_draft_order_id),
		    biteship_order_id = COALESCE(NULLIF($2, ''), biteship_order_id),
		    biteship_tracking_id = COALESCE(NULLIF($3, ''), biteship_tracking_id),
		    biteship_waybill_id = COALESCE(NULLIF($4, ''), biteship_waybill_id),
		    updated_at = NOW()
		WHERE id = $5
	`
	_, err := r.db.Exec(query, draftOrderID, orderID, trackingID, waybillID, id)
	return err
}

// UpdateBiteshipTracking updates Biteship tracking info for a shipment
func (r *shippingRepository) UpdateBiteshipTracking(id int, trackingID, waybillID string) error {
	query := `
		UPDATE shipments
		SET biteship_tracking_id = $1,
		    biteship_waybill_id = $2,
		    tracking_number = $2,
		    updated_at = NOW()
		WHERE id = $3
	`
	_, err := r.db.Exec(query, trackingID, waybillID, id)
	return err
}

// GetShipmentByBiteshipTrackingID retrieves a shipment by Biteship tracking ID
func (r *shippingRepository) GetShipmentByBiteshipTrackingID(trackingID string) (*models.Shipment, error) {
	query := `
		SELECT id, order_id, provider_code, provider_name, service_code, service_name,
		       cost, etd, weight, tracking_number, status, origin_city_id, origin_city_name,
		       destination_city_id, destination_city_name,
		       COALESCE(biteship_draft_order_id, '') as biteship_draft_order_id,
		       COALESCE(biteship_order_id, '') as biteship_order_id,
		       COALESCE(biteship_tracking_id, '') as biteship_tracking_id,
		       COALESCE(biteship_waybill_id, '') as biteship_waybill_id,
		       shipped_at, delivered_at, created_at, updated_at
		FROM shipments
		WHERE biteship_tracking_id = $1
	`

	var s models.Shipment
	var trackingNumber sql.NullString
	var shippedAt, deliveredAt sql.NullTime

	err := r.db.QueryRow(query, trackingID).Scan(
		&s.ID, &s.OrderID, &s.ProviderCode, &s.ProviderName, &s.ServiceCode, &s.ServiceName,
		&s.Cost, &s.ETD, &s.Weight, &trackingNumber, &s.Status, &s.OriginCityID, &s.OriginCityName,
		&s.DestinationCityID, &s.DestinationCityName,
		&s.BiteshipDraftOrderID, &s.BiteshipOrderID, &s.BiteshipTrackingID, &s.BiteshipWaybillID,
		&shippedAt, &deliveredAt, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if trackingNumber.Valid {
		s.TrackingNumber = trackingNumber.String
	}
	if shippedAt.Valid {
		s.ShippedAt = &shippedAt.Time
	}
	if deliveredAt.Valid {
		s.DeliveredAt = &deliveredAt.Time
	}

	return &s, nil
}

// ============================================
// BITESHIP LOCATIONS
// ============================================

// CreateBiteshipLocation creates a new Biteship location
func (r *shippingRepository) CreateBiteshipLocation(location *models.BiteshipLocation) error {
	query := `
		INSERT INTO biteship_locations (
			user_id, location_id, area_id, area_name, contact_name, contact_phone, address, postal_code
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at
	`

	return r.db.QueryRow(
		query,
		location.UserID, location.LocationID, location.AreaID, location.AreaName,
		location.ContactName, location.ContactPhone, location.Address, location.PostalCode,
	).Scan(&location.ID, &location.CreatedAt)
}

// GetBiteshipLocationsByUserID retrieves all Biteship locations for a user
func (r *shippingRepository) GetBiteshipLocationsByUserID(userID int) ([]models.BiteshipLocation, error) {
	query := `
		SELECT id, user_id, location_id, area_id, area_name, contact_name, contact_phone, address, postal_code, created_at
		FROM biteship_locations
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locations []models.BiteshipLocation
	for rows.Next() {
		var loc models.BiteshipLocation
		err := rows.Scan(
			&loc.ID, &loc.UserID, &loc.LocationID, &loc.AreaID, &loc.AreaName,
			&loc.ContactName, &loc.ContactPhone, &loc.Address, &loc.PostalCode, &loc.CreatedAt,
		)
		if err != nil {
			continue
		}
		locations = append(locations, loc)
	}

	return locations, nil
}

// GetBiteshipLocationByID retrieves a Biteship location by ID
func (r *shippingRepository) GetBiteshipLocationByID(id int) (*models.BiteshipLocation, error) {
	query := `
		SELECT id, user_id, location_id, area_id, area_name, contact_name, contact_phone, address, postal_code, created_at
		FROM biteship_locations
		WHERE id = $1
	`

	var loc models.BiteshipLocation
	err := r.db.QueryRow(query, id).Scan(
		&loc.ID, &loc.UserID, &loc.LocationID, &loc.AreaID, &loc.AreaName,
		&loc.ContactName, &loc.ContactPhone, &loc.Address, &loc.PostalCode, &loc.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &loc, nil
}
