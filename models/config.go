package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Config struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	LiffID             string             `bson:"liff_id" json:"liff_id"`
	ChannelAccessToken string             `bson:"channel_access_token" json:"channel_access_token"`
	ChannelSecret      string             `bson:"channel_secret" json:"channel_secret"`
	Tiers              []TierDetail       `bson:"tiers" json:"tiers"`
	TelegramBotToken   string             `bson:"telegram_bot_token" json:"telegram_bot_token"`
	TelegramChatID     string             `bson:"telegram_chat_id" json:"telegram_chat_id"`
	FirebaseConfig     FirebaseConfig     `bson:"firebase_config" json:"firebase_config"`
	FlexMessages       FlexMessages       `bson:"flex_messages" json:"flexMessages"`
	SiteTemplate       SiteTemplateConfig `bson:"site_template" json:"siteTemplate"`
	ApiEndpoint        string             `bson:"api_endpoint" json:"api_endpoint"`
	ApiKey             string             `bson:"api_key" json:"api_key"`
	LineAt             string             `bson:"line_at" json:"line_at"`
	LineSyncURL        string             `bson:"line_sync_url" json:"line_sync_url"`
}

type FirebaseConfig struct {
	Credential string `bson:"credential" json:"credential"`
	BucketName string `bson:"bucket_name" json:"bucketName"`
}

type TierDetail struct {
	Name                string `bson:"name" json:"name"`
	Period              int    `bson:"period" json:"period"`                             // หน่วยเป็นชั่วโมง
	Target              int    `bson:"target" json:"target"`
	Reward              int    `bson:"reward" json:"reward"`
	MaxLevel            int    `bson:"max_level" json:"max_level"`
	FollowUpHours       int    `bson:"follow_up_hours" json:"follow_up_hours"`           // หน่วยเป็นชั่วโมง
	ExpireRewardHours   int    `bson:"expire_reward_hours" json:"expire_reward_hours"`   // หน่วยเป็นชั่วโมง
	MaxConsecutiveFails int    `bson:"max_consecutive_fails" json:"max_consecutive_fails"`
	NotifyBeforeExpire  int    `bson:"notify_before_expire" json:"notify_before_expire"` // หน่วยเป็นชั่วโมง
	NotifyInterval      int    `bson:"notify_interval" json:"notify_interval"`           // หน่วยเป็นชั่วโมง
	ProcessingDelay     int    `bson:"processing_delay" json:"processing_delay"` // หน่วยเป็นนาที

}

type FlexMessages struct {
	Followup           BaseFlexMessageContent `bson:"followup" json:"followup"`
	MissionSuccess     BaseFlexMessageContent `bson:"mission_success" json:"missionSuccess"`
	MissionFailed      BaseFlexMessageContent `bson:"mission_failed" json:"missionFailed"`
	MissionComplete    BaseFlexMessageContent `bson:"mission_complete" json:"missionComplete"`
	GetReward          BaseFlexMessageContent `bson:"get_reward" json:"getReward"`
	RewardNotification BaseFlexMessageContent `bson:"reward_notification" json:"rewardNotification"`
}

type BaseFlexMessageContent struct {
	Title          string `bson:"title" json:"title"`
	Description    string `bson:"description" json:"description"`
	SubDescription string `bson:"sub_description,omitempty" json:"subDescription,omitempty"`
	ButtonTitle    string `bson:"button_title,omitempty" json:"buttonTitle,omitempty"`
	ButtonUrl      string `bson:"button_url,omitempty" json:"buttonUrl,omitempty"`
	ImageUrl       string `bson:"image_url" json:"imageUrl"`
}

type SiteTemplateConfig struct {
	Logo               string                   `bson:"logo" json:"logo"`
	ProjectName        string                   `bson:"project_name" json:"project_name"`
	Slogan             string                   `bson:"slogan" json:"slogan"`
	PhoneNumberConfirm PhoneNumberConfirmConfig `bson:"phone_number_confirm" json:"phoneNumberConfirm"`
	MainClient         MainClientConfig         `bson:"main_client" json:"mainClient"`
	Slider             SliderConfig             `bson:"slider" json:"slider"`
	MissionClient      MissionClientConfig      `bson:"mission_client" json:"missionClient"`
	HistoryClient      HistoryClientConfig      `bson:"history_client" json:"historyClient"`
	RestartMission     RestartMissionConfig     `bson:"restart_mission" json:"restartMission"`
	Condition          ConditionConfig          `bson:"condition" json:"condition"`
	Contact            ContactConfig            `bson:"contact" json:"contact"`
}

type PhoneNumberConfirmConfig struct {
	Background string `bson:"background" json:"background"`
}

type MainClientConfig struct {
	Background              string `bson:"background" json:"background"`
	NavColor                string `bson:"nav_color" json:"nav_color"`
	NavTextColor            string `bson:"nav_text_color" json:"nav_text_color"`
	ButtonGetText           string `bson:"button_get_text" json:"button_get_text"`
	ButtonGetGradientStart  string `bson:"button_get_gradient_start" json:"button_get_gradient_start"`
	ButtonGetGradientEnd    string `bson:"button_get_gradient_end" json:"button_get_gradient_end"`
	ButtonGetTextColor      string `bson:"button_get_text_color" json:"button_get_text_color"`
	ButtonNextText          string `bson:"button_next_text" json:"button_next_text"`
	ButtonNextGradientStart string `bson:"button_next_gradient_start" json:"button_next_gradient_start"`
	ButtonNextGradientEnd   string `bson:"button_next_gradient_end" json:"button_next_gradient_end"`
	ButtonNextTextColor     string `bson:"button_next_text_color" json:"button_next_text_color"`
}

type SliderConfig struct {
	Images []SliderImage `bson:"images" json:"images"`
}

type SliderImage struct {
	URL  string `bson:"url" json:"url"`
	Link string `bson:"link" json:"link"`
}

type MissionClientConfig struct {
	Background              string       `bson:"background" json:"background"`
	NavColor                string       `bson:"nav_color" json:"nav_color"`
	NavTextColor            string       `bson:"nav_text_color" json:"nav_text_color"`
	ButtonText              string       `bson:"button_text" json:"button_text"`
	ButtonTextColor         string       `bson:"button_text_color" json:"button_text_color"`
	ButtonTextGradientStart string       `bson:"button_text_gradient_start" json:"button_text_gradient_start"`
	ButtonTextGradientEnd   string       `bson:"button_text_gradient_end" json:"button_text_gradient_end"`
	ButtonTextURL           string       `bson:"button_text_url" json:"button_text_url"`
	UsernameText            string       `bson:"username_text" json:"username_text"`
	UsernameTextColor       string       `bson:"username_text_color" json:"username_text_color"`
	Tiers                   []TierConfig `bson:"tiers" json:"tiers"`
}

type TierConfig struct {
	Icon                      string `bson:"icon" json:"icon"`
	BgGradientStart           string `bson:"bg_gradient_start" json:"bg_gradient_start"`
	BgGradientEnd             string `bson:"bg_gradient_end" json:"bg_gradient_end"`
	BgLabel                   string `bson:"bg_label" json:"bg_label"`
	TextLabelColor            string `bson:"text_label_color" json:"text_label_color"`
	ButtonGradientStart       string `bson:"button_gradient_start" json:"button_gradient_start"`
	ButtonGradientEnd         string `bson:"button_gradient_end" json:"button_gradient_end"`
	ButtonTextColor           string `bson:"button_text_color" json:"button_text_color"`
	ActiveButtonGradientStart string `bson:"active_button_gradient_start" json:"active_button_gradient_start"`
	ActiveButtonGradientEnd   string `bson:"active_button_gradient_end" json:"active_button_gradient_end"`
	ActiveButtonTextColor     string `bson:"active_button_text_color" json:"active_button_text_color"`
	TextColor                 string `bson:"text_color" json:"text_color"`
	ProgressBarColor          string `bson:"progress_bar_color" json:"progress_bar_color"`
	ProgressTextColor         string `bson:"progress_text_color" json:"progress_text_color"`
	BoxGradientStart          string `bson:"box_gradient_start" json:"box_gradient_start"`
	BoxGradientEnd            string `bson:"box_gradient_end" json:"box_gradient_end"`
	ColorToggle               string `bson:"color_toggle" json:"color_toggle"`
}

type HistoryClientConfig struct {
	Title            string `bson:"title" json:"title"`
	Subtitle         string `bson:"subtitle" json:"subtitle"`
	ColorTitle       string `bson:"color_title" json:"color_title"`
	ColorSubtitle    string `bson:"color_subtitle" json:"color_subtitle"`
	ColorToggle      string `bson:"color_toggle" json:"color_toggle"`
	BoxGradientStart string `bson:"box_gradient_start" json:"box_gradient_start"`
	BoxGradientEnd   string `bson:"box_gradient_end" json:"box_gradient_end"`
}

type RestartMissionConfig struct {
	Icon        string `bson:"icon" json:"icon"`
	Title       string `bson:"title" json:"title"`
	Subtitle    string `bson:"subtitle" json:"subtitle"`
	ButtonTitle string `bson:"button_title" json:"buttonTitle"`
}

type ConditionConfig struct {
	Items            []string `bson:"items" json:"items"`
	BoxGradientStart string   `bson:"box_gradient_start" json:"box_gradient_start"`
	BoxGradientEnd   string   `bson:"box_gradient_end" json:"box_gradient_end"`
}

type ContactConfig struct {
	Items []ContactItem `bson:"items" json:"items"`
}

type ContactItem struct {
	Icon  string `bson:"icon" json:"icon"`
	Title string `bson:"title" json:"title"`
	URL   string `bson:"url" json:"url"`
}
