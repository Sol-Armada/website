package realtime

import "github.com/sol-armada/sol-bot/database/dbnotify"

// TopicForNotifyChannel maps sol-bot dbnotify channels to websocket topics.
func TopicForNotifyChannel(channel string) (string, bool) {
	switch channel {
	case dbnotify.ChannelMembers:
		return TopicAdminMembers, true
	case dbnotify.ChannelAttendance:
		return TopicAdminAttendance, true
	case dbnotify.ChannelTokens:
		return TopicAdminTokenLedger, true
	default:
		return "", false
	}
}
