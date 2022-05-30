package event

type popularEvents struct {
	PlatformID   string  `json:"platform_id"`
	PlatformName string  `json:"platform_name"`
	Project      string  `json:"project"`
	State        uint8   `json:"-"` // 0 停用 1 启用
	AwayTeam     string  `json:"away_team"`
	HomeTeam     string  `json:"home_team"`
	HomeOdds     float64 `json:"home_odds"`
	AwayOdds     float64 `json:"away_odds"`
	EqOdds       float64 `json:"eq_odds"`
	EventName    string  `json:"event_name"`
	EventAt      int64   `json:"event_at"`
}

type imSportParam struct {
	// 指出体育项目1 足球 2 篮球 3 网球 6 田径 7 羽毛球 8 棒球 11 拳击 15 飞镖 18 草地曲棍球 19 美式足球 21 高尔夫球 23 手球 25 冰上曲棍球 29 赛车运动
	SportId int `json:"SportId"`
	// 1 = 早盘 2 = 今日 3 = 滚球
	Market int `json:"Market"`
	// 1 = 马来盘 2 = 香港盘 3 = 欧洲盘 4 = 印尼盘
	OddsType int `json:"OddsType"`
	// 指出请求是连串过关赛事或者非连串过关赛事
	IsCombo bool `json:"IsCombo"`
	// 指出返回的页数.
	PageRecords int `json:"PageRecords"`
	Page        int `json:"Page"`
	// 新生成的时间戳将持续5分钟
	TimeStamp    string `json:"TimeStamp"`
	BetTypeIds   []int  `json:"BetTypeIds"`
	LanguageCode string `json:"LanguageCode"`
	PeriodIds    []int  `json:"PeriodIds"`
}

type imSportCountParam struct {
	IsCombo bool `json:"IsCombo"`
	// 新生成的时间戳将持续5分钟
	TimeStamp    string `json:"TimeStamp"`
	LanguageCode string `json:"LanguageCode"`
}

// im 体育 数据统计
type imCount struct {
	EarlyFECount int `json:"early_fe_count"` // 早盘
	TodayFECount int `json:"today_fe_count"` // 今日
	RBFECount    int `json:"rbfe_count"`     // 滚球
}

type latest map[string]latestData

type latestData struct {
	EventAt  int64   `json:"event_at"`
	HomeOdds float64 `json:"home_odds"`
	AwayOdds float64 `json:"away_odds"`
	EqOdds   float64 `json:"eq_odds"`
}

type popularEventRedis struct {
	AwayOdds     float64 `json:"away_odds"`
	AwayTeam     string  `json:"away_team"`
	AwayTeamLogo string  `json:"away_team_logo"`
	EqOdds       float64 `json:"eq_odds"`
	EventAt      int     `json:"event_at"`
	EventName    string  `json:"event_name"`
	HomeOdds     float64 `json:"home_odds"`
	HomeTeam     string  `json:"home_team"`
	HomeTeamLogo string  `json:"home_team_logo"`
	ID           string  `json:"id"`
	PlatformID   string  `json:"platform_id"`
	Sort         int     `json:"sort"`
}

const (
	// im体育url
	imURL = "http://ipis-ykyule.imapi.net"
	// IM体育 场馆id
	imPlatformID = "5864536520308659696"
	// VN 越南文 中文 CHS
	imLang              = "VN"
	imOddsTypeMalaysia  = 1 // 马来盘
	imOddsTypeHangKong  = 2 // 香港盘
	imOddsTypeEurope    = 3 // 欧洲盘
	imOddsTypeIndonesia = 4 // 印尼盘

	// 雷火电竞url
	ftURL = "https://api-v4.ely889.com/"
	// 雷火电竞 场馆id
	ftPlatformID = "2854120181948444476"
	//en = english 英文
	//zh = chinese 中文
	//th = thai 泰文
	//vn = viet 越南文
	//id = indonesia bahasa 印尼语
	//ms = malay 马来文
	//jp = japanese 日文
	//kr = korea 韩文
	//es = spanish 西班牙文
	ftLang = "vn"
	// redis 存热门数据的键 key ==> value
	popularRedisKey = "popular_events"
)

var allData = make(latest)
