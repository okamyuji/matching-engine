package domain

// Gender ユーザーの性別
type Gender string

const (
	GenderMale   Gender = "male"
	GenderFemale Gender = "female"
	GenderOther  Gender = "other"
)

// Prefecture 日本の都道府県
type Prefecture string

const (
	// 北海道
	PrefectureHokkaido Prefecture = "hokkaido"

	// 東北
	PrefectureAomori    Prefecture = "aomori"
	PrefectureIwate     Prefecture = "iwate"
	PrefectureMiyagi    Prefecture = "miyagi"
	PrefectureAkita     Prefecture = "akita"
	PrefectureYamagata  Prefecture = "yamagata"
	PrefectureFukushima Prefecture = "fukushima"

	// 関東
	PrefectureIbaraki  Prefecture = "ibaraki"
	PrefectureTochigi  Prefecture = "tochigi"
	PrefectureGunma    Prefecture = "gunma"
	PrefectureSaitama  Prefecture = "saitama"
	PrefectureChiba    Prefecture = "chiba"
	PrefectureTokyo    Prefecture = "tokyo"
	PrefectureKanagawa Prefecture = "kanagawa"

	// 中部
	PrefectureNiigata   Prefecture = "niigata"
	PrefectureToyama    Prefecture = "toyama"
	PrefectureIshikawa  Prefecture = "ishikawa"
	PrefectureFukui     Prefecture = "fukui"
	PrefectureYamanashi Prefecture = "yamanashi"
	PrefectureNagano    Prefecture = "nagano"
	PrefectureGifu      Prefecture = "gifu"
	PrefectureShizuoka  Prefecture = "shizuoka"
	PrefectureAichi     Prefecture = "aichi"

	// 近畿
	PrefectureMie      Prefecture = "mie"
	PrefectureShiga    Prefecture = "shiga"
	PrefectureKyoto    Prefecture = "kyoto"
	PrefectureOsaka    Prefecture = "osaka"
	PrefectureHyogo    Prefecture = "hyogo"
	PrefectureNara     Prefecture = "nara"
	PrefectureWakayama Prefecture = "wakayama"

	// 中国
	PrefectureTottori   Prefecture = "tottori"
	PrefectureShimane   Prefecture = "shimane"
	PrefectureOkayama   Prefecture = "okayama"
	PrefectureHiroshima Prefecture = "hiroshima"
	PrefectureYamaguchi Prefecture = "yamaguchi"

	// 四国
	PrefectureTokushima Prefecture = "tokushima"
	PrefectureKagawa    Prefecture = "kagawa"
	PrefectureEhime     Prefecture = "ehime"
	PrefectureKochi     Prefecture = "kochi"

	// 九州・沖縄
	PrefectureFukuoka   Prefecture = "fukuoka"
	PrefectureSaga      Prefecture = "saga"
	PrefectureNagasaki  Prefecture = "nagasaki"
	PrefectureKumamoto  Prefecture = "kumamoto"
	PrefectureOita      Prefecture = "oita"
	PrefectureMiyazaki  Prefecture = "miyazaki"
	PrefectureKagoshima Prefecture = "kagoshima"
	PrefectureOkinawa   Prefecture = "okinawa"
)

// BodyType 体型
type BodyType string

const (
	BodyTypeSlim     BodyType = "slim"
	BodyTypeAverage  BodyType = "average"
	BodyTypeAthletic BodyType = "athletic"
	BodyTypeLarge    BodyType = "large"
)

// Education 学歴レベル
type Education string

const (
	EducationHighSchool Education = "high_school"
	EducationVocational Education = "vocational"
	EducationUniversity Education = "university"
	EducationGraduate   Education = "graduate"
)

// IncomeRange 年収範囲
type IncomeRange int

const (
	Income0to200     IncomeRange = 0  // ~200万円
	Income200to400   IncomeRange = 1  // 200-400万円
	Income400to600   IncomeRange = 2  // 400-600万円
	Income600to800   IncomeRange = 3  // 600-800万円
	Income800to1000  IncomeRange = 4  // 800-1000万円
	Income1000to1500 IncomeRange = 5  // 1000-1500万円
	Income1500to2000 IncomeRange = 6  // 1500-2000万円
	Income2000to3000 IncomeRange = 7  // 2000-3000万円
	Income3000to5000 IncomeRange = 8  // 3000-5000万円
	Income5000Plus   IncomeRange = 9  // 5000万円以上
	IncomePrivate    IncomeRange = 10 // 非公開
)

// Level 収入レベル (0-10) を返す
func (i IncomeRange) Level() int {
	return int(i)
}

// MarriageDesire 結婚に対する希望
type MarriageDesire string

const (
	MarriageWantSoon       MarriageDesire = "want_soon"
	MarriageWantEventually MarriageDesire = "want_eventually"
	MarriageUndecided      MarriageDesire = "undecided"
	MarriageNotWant        MarriageDesire = "not_want"
)

// ChildrenDesire 子供に対する希望
type ChildrenDesire string

const (
	ChildrenWant      ChildrenDesire = "want"
	ChildrenNotWant   ChildrenDesire = "not_want"
	ChildrenUndecided ChildrenDesire = "undecided"
)

// SmokingStatus 喫煙状況
type SmokingStatus string

const (
	SmokingNonSmoker  SmokingStatus = "non_smoker"
	SmokingOccasional SmokingStatus = "occasional"
	SmokingSmoker     SmokingStatus = "smoker"
)

// DrinkingStatus 飲酒状況
type DrinkingStatus string

const (
	DrinkingNonDrinker DrinkingStatus = "non_drinker"
	DrinkingSocial     DrinkingStatus = "social"
	DrinkingRegular    DrinkingStatus = "regular"
)

// Photo ユーザーの写真
type Photo struct {
	URL       string `json:"url"`
	IsPrimary bool   `json:"is_primary"`
	Order     int    `json:"order"`
}
