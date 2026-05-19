package model

type Seat struct {
	ID               int64  `json:"id"`
	Code             string `json:"code"`
	ZoneCode         string `json:"zone_code"`
	ZoneName         string `json:"zone_name"`
	SeatType         string `json:"seat_type"`
	FixedOwnerName   string `json:"fixed_owner_name,omitempty"`
	IsActive         bool   `json:"is_active"`
	Availability     string `json:"availability"`
	CurrentBooker    string `json:"current_booker,omitempty"`
	CurrentBookingID *int64 `json:"current_booking_id,omitempty"`
}

type Reservation struct {
	ID          int64  `json:"id"`
	SeatID      int64  `json:"seat_id"`
	SeatCode    string `json:"seat_code"`
	ZoneCode    string `json:"zone_code"`
	ZoneName    string `json:"zone_name"`
	BookerName  string `json:"booker_name"`
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
	Status      string `json:"status"`
	Note        string `json:"note,omitempty"`
	CreatedAt   string `json:"created_at"`
	CancelledAt string `json:"cancelled_at,omitempty"`
}

type CreateReservationRequest struct {
	SeatCode   string `json:"seat_code"`
	BookerName string `json:"booker_name"`
	StartTime  string `json:"start_time"`
	EndTime    string `json:"end_time"`
	Note       string `json:"note"`
}

type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}
