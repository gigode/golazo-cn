package constants

// Menu items
const (
	MenuStats       = "已完赛"
	MenuLiveMatches = "实时比赛"
	MenuSettings    = "设置"
)

// Panel titles
const (
	PanelLiveMatches       = "实时比赛"
	PanelFinishedMatches   = "已完赛"
	PanelMatchDetails      = "比赛详情"
	PanelMatchList         = "比赛列表"
	PanelUpcomingMatches   = "即将开赛"
	PanelMinuteByMinute    = "赛况直播"
	PanelMatchStatistics   = "比赛数据"
	PanelUpdates           = "动态"
	PanelLeaguePreferences = "联赛偏好"
)

// Empty state messages
const (
	EmptyNoLiveMatches     = "暂无实时比赛"
	EmptyNoFinishedMatches = "暂无已完赛"
	EmptySelectMatch       = "请选择一场比赛"
	EmptyNoUpdates         = "暂无动态"
	EmptyNoMatches         = "暂无比赛"
)

// Error messages
const (
	ErrorLoadFailed   = "无法加载数据"
	ErrorMatchDetails = "无法加载比赛详情"
	ErrorRetryHint    = "r：重试"
)

// Help text
const (
	HelpMainMenu           = "↑/↓：导航  Enter：选择  q：退出"
	HelpMatchesView        = "↑/↓：导航  r：刷新  x：数据  s：积分榜  /：筛选  Esc：返回  q：退出"
	HelpSettingsView       = "↑/↓：导航  ←/→：切换标签  Space：勾选  /：筛选  Enter：保存  Esc：返回"
	HelpStatsView          = "h/l：日期范围  j/k：导航  Tab：聚焦详情  聚焦后 ↑/↓：滚动  r：刷新详情  /：筛选  Esc：返回"
	HelpStatsViewUnfocused = "Tab：聚焦详情"
	HelpStatsViewFocused   = "Tab：取消聚焦  s：积分榜  f：阵容  x：全部数据  ↑/↓：滚动"
	HelpStandingsDialog    = "Esc：关闭"
	HelpFormationsDialog   = "Tab/←/→：切换球队  Esc：关闭"
	HelpStatisticsDialog   = "↑/↓：导航  Esc：关闭"

	// Edge case user-facing hints
	ErrorNoStatistics = "暂无比赛数据"
	ErrorNoStandings  = "暂无积分榜"
)

// Status text
const (
	StatusLive            = "直播"
	StatusFinished        = "完场"
	StatusNotStarted      = "对"
	StatusNotStartedShort = "未开"
	StatusFinishedText    = "已结束"
)

// Loading text
const (
	LoadingFetching = "加载中..."
)

// Notification text
const (
	// NotificationTitleGoal is the title shown in goal notifications.
	NotificationTitleGoal = "⚽ GOLAZO!"
)

// Stats labels
const (
	LabelStatus = "状态："
	LabelScore  = "比分："
	LabelLeague = "联赛："
	LabelDate   = "日期："
	LabelVenue  = "场地："
)
