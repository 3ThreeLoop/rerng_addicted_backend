package share

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/tarantool/go-tarantool/v2/decimal"
	_ "github.com/tarantool/go-tarantool/v2/uuid"
)

type UserContext struct {
	Id           int
	UserUuid     string
	UserName     string
	LoginSession string
	RoleId       uint64
	Exp          time.Time
	UserAgent    string
	Ip           string
	StatusId     int
}
type Paging struct {
	Page    int `json:"page" query:"page" validate:"required,min=1"`
	Perpage int `json:"per_page" query:"per_page" validate:"required,min=1"`
}
type Sort struct {
	Property  string `json:"property" validate:"required"`
	Direction string `json:"direction" validate:"required,oneof=asc desc"`
}
type Filter struct {
	Property string      `json:"property" validate:"required"`
	Operator string      `json:"operator" validate:"required"`
	Value    interface{} `json:"value" validate:"required"`
}

type FieldUuid struct {
	Uuid uuid.UUID `json:"id"`
}
type FieldId struct {
	Id uint64 `json:"id"`
}

type Status struct {
	Id         int    `json:"id"`
	StatusName string `json:"status_name"`
}

type BroadcastResponse struct {
	Topic string          `json:"topic"`
	Data  json.RawMessage `json:"data"`
}

var StatusData = []Status{
	{Id: 1, StatusName: "Active"},
	{Id: 2, StatusName: "Inactive"},
	{Id: 3, StatusName: "Suspended"},
	{Id: 4, StatusName: "Deleted"},
}

// Platform Mini
type Platform struct {
	ID                     uint64    `json:"id"`
	MembershipPlatformUUID uuid.UUID `json:"membership_platform_uuid"`
	PlatformName           string    `json:"platform_name"`
	PlatformHost           string    `json:"platform_host"`
	PlatformToken          string    `json:"platform_token"`
	PlatformExtraPayload   string    `json:"platform_extra_payload"`
	InternalToken          string    `json:"internal_token"`
	StatusID               uint64    `json:"status_id"`
	Order                  uint64    `json:"order"`
}

type PlayerNewRequest struct {
	ID              uint64          `json:"id"`
	FirstName       string          `json:"first_name"`
	LastName        string          `json:"last_name"`
	UserName        string          `json:"user_name"`
	Password        string          `json:"password"`
	PasswordConfirm string          `json:"password_comfirm"`
	Email           string          `json:"email"`
	RoleID          *int            `json:"role_id"`
	PhoneNumber     *string         `json:"phone_number" `
	Commission      decimal.Decimal `json:"commission"`
}

type Player struct {
	ID         uint64    `json:"id"`
	PlayerUUID uuid.UUID `json:"player_uuid"`
}

type PlayerResponse struct {
	Player []Player `json:"player"`
}

type BroadcastResultData struct {
	Result []Result `json:"result"`
}

type Result struct {
	ResultUUID  uuid.UUID `json:"result_uuid"`
	RoundUUID   uuid.UUID `json:"round_uuid"`
	RoundNo     string    `json:"round_no"`
	ChannelID   uint64    `json:"channel_id"`
	BetTypeID   uint64    `json:"bet_type_id"`
	BetTypeName string    `json:"bet_type_name"`
}

type BroadcastAnnounceData struct {
	Announcement AnnouncementDetails `json:"announcement"`
}

type AnnouncementDetails struct {
	ID                     int    `json:"id"`
	AnnouncementUUID       string `json:"announcement_uuid"`
	AnnouncementDesc       string `json:"announcement_desc"`
	ScheduleAnnounce       string `json:"schedule_announce"`
	ScheduleAnnounceExpire string `json:"schedule_announce_expire"`
	AnnounceRepeat         int    `json:"announce_repeat"`
	StatusID               int    `json:"status_id"`
	ChannelID              int    `json:"channel_id"`
	Order                  int    `json:"order"`
}

type BroadcastFightOddData struct {
	FightOdd []FightOdd `json:"fight_odd"`
}

type FightOdd struct {
	FightOddUUID string `json:"fight_odd_uuid"`
	ChannelID    uint64 `json:"channel_id"`
	RedOdd       string `json:"red_odd"`
	BlueOdd      string `json:"blue_odd"`
	DrawOdd      string `json:"draw_odd"`
}

type BroadcastRoundData struct {
	Rounds []BroadcastRoundStatus `json:"rounds"`
}

type BroadcastRoundStatus struct {
	ID        uint64    `json:"id"`
	RoundNo   string    `json:"round_no"`
	RoundUUID uuid.UUID `json:"round_uuid"`
	StatusID  uint64    `json:"status_id"`
	ChannelID uint64    `json:"channel_id"`
}

type BroadcastChannelData struct {
	Channel []BroadcastChannel `json:"channel"`
}

type BroadcastChannel struct {
	ChannelID   uint64 `json:"channel_id"`
	ChannelName string `json:"channel_name"`
	StreamOne   string `json:"stream_one"`
	StreamTwo   string `json:"stream_two"`
	StatusID    uint64 `json:"status_id"`
}

type BroadcastBalanceData struct {
	PlayerBalance []BroadcastBalance `json:"player_balance"`
}

type BroadcastBalance struct {
	PlayerUUID uuid.UUID       `json:"player_uuid"`
	CurrencyID uint64          `json:"currency_id"`
	Balance    decimal.Decimal `json:"balance"`
}

// for user notification
type BroadcastPlayerNotificationData struct {
	Notifications []BroadcastPlayerNotification `json:"player_notifications"`
}
type BroadcastPlayerNotification struct {
	PlayerUUID  uuid.UUID `json:"player_uuid"`
	Context     string    `json:"context"`
	Subject     string    `json:"subject"`
	Description string    `json:"description"`
	IconID      uint64    `json:"icon_id"`
}

// for bet limit
type BroadcastBetLimitData struct {
	BetLimts []BroadcastBetLimit `json:"bet_limits"`
}
type BroadcastBetLimit struct {
	BetLimitUUID uuid.UUID       `json:"bet_limit_uuid"`
	BetLimit     decimal.Decimal `json:"bet_limit"`
	ChannelID    uint64          `json:"channel_id"`
}

// for announcement banner
type BroadcastAnnouceBannerData struct {
	Announcement AnnouncementBannerDetail `json:"announcement_banner"`
}

type AnnouncementBannerDetail struct {
	TextEN    string `json:"text_en"`
	TextZH    string `json:"text_zh"`
	TextKM    string `json:"text_km"`
	ChannelID int    `json:"channel_id"`
}

type BroadcastNewTokenData struct {
	NewToken []BroadcastNewToken `json:"new_token"`
}
type BroadcastNewToken struct {
	PlayerUUID uuid.UUID `json:"player_uuid"`
	NewToken   string    `json:"new_token"`
}

// for update player fight odd
type BroadcastUpdatePlayerFightOddData struct {
	UpdatePlayerFightOdd []UpdatePlayerFightOdd `json:"fight_odds"`
}

type UpdatePlayerFightOdd struct {
	PlayerUUID uuid.UUID `json:"player_uuid"`
	BetUUID    uuid.UUID `json:"bet_uuid"`
}
