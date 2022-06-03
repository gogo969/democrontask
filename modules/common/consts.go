package common

const (
	Vip1  int = 1
	Vip2  int = 2
	Vip3  int = 3
	Vip4  int = 4
	Vip5  int = 5
	Vip6  int = 6
	Vip7  int = 7
	Vip8  int = 8
	Vip9  int = 9
	Vip10 int = 10
)

// 会员等级调整类型
const (
	MemberLevelUpgrade    = 201 //会员升级
	MemberLevelRelegation = 202 //会员保级
	MemberLevelDowngrade  = 203 //会员降级
	MemberLevelRecover    = 204 //会员等级恢复
)

// 红利类型
const (
	DividendSite      = 211 //平台红利(站点)
	DividendUpgrade   = 212 //升级红利
	DividendBirthday  = 213 //生日红利
	DividendMonthly   = 214 //每月红利
	DividendRedPacket = 215 //红包红利
	DividendMaintain  = 216 //维护补偿
	DividendDeposit   = 217 //存款优惠
	DividendPromo     = 218 //活动红利
	DividendInvite    = 219 //推荐红利
	DividendAdjust    = 220 //红利调整
	DividendResetPlat = 221 //场馆余额负数清零
	DividendAgency    = 222 //代理红利
)

// 红利审核状态
const (
	DividendReviewing    = 231 //红利审核中
	DividendReviewPass   = 232 //红利审核通过
	DividendReviewReject = 233 //红利审核不通过
)

// 存款状态
const (
	DepositConfirming = 361 //确认中
	DepositSuccess    = 362 //存款成功
	DepositCancelled  = 363 //存款已取消
	DepositReviewing  = 364 //存款审核中
)

// 取款状态
const (
	WithdrawReviewing     = 371 //审核中
	WithdrawReviewReject  = 372 //审核拒绝
	WithdrawDealing       = 373 //出款中
	WithdrawSuccess       = 374 //提款成功
	WithdrawFailed        = 375 //出款失败
	WithdrawAbnormal      = 376 //异常订单
	WithdrawAutoPayFailed = 377 // 代付失败
	WithdrawHangup        = 378 // 挂起
	WithdrawDispatched    = 379 // 已派单
)

// 活动奖品审核状态
const (
	PromoGiftReviewing    = 401 //审核中
	PromoGiftReviewPass   = 402 //审核通过
	PromoGiftReviewReject = 403 //审核拒绝
)

// 天天签到派奖状态
const (
	SignPrizeUnReach        = 701 //未达标
	SignPrizeWaitHandOut    = 702 //待派奖
	SignPrizeHandOutSuccess = 703 //派奖成功
	SignPrizeHandOutFailed  = 704 //派奖失败
	SignPrizeInvalid        = 705 //已失效
	SignPrizeReceived       = 706 //已领取
)
