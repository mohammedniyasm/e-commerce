package request

type ListNotificationsRequest struct {
	Page       int  `form:"page"`
	Limit      int  `form:"limit"`
	UnreadOnly bool `form:"unread_only"`
}
