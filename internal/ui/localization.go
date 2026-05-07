package ui

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/0xjuanma/golazo/internal/api"
	"github.com/0xjuanma/golazo/internal/data"
)

const translationCacheFile = "translations_zh.json"

var entityTranslator = newEntityTranslator()
var entityLocalizationEnabled = true

// SetEntityLocalizationEnabled toggles translation of dynamic entities such as teams,
// players, leagues, referees, and venues. Interface text remains Chinese either way.
func SetEntityLocalizationEnabled(enabled bool) {
	entityLocalizationEnabled = enabled
}

// LocalizeEntityName returns a Simplified Chinese display name for dynamic football entities.
// It first checks built-in translations, then uses the persistent machine-translation cache.
func LocalizeEntityName(name string) string {
	return localizeEntityName(name)
}

// WarmEntityTranslations preloads Chinese translations for entity names.
// It is intentionally bounded so large match lists do not create unbounded network traffic.
func WarmEntityTranslations(names ...string) {
	if !entityLocalizationEnabled {
		return
	}

	seen := make(map[string]bool, len(names))
	unique := make([]string, 0, len(names))
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		unique = append(unique, name)
	}

	const maxWarmTranslations = 80
	if len(unique) > maxWarmTranslations {
		unique = unique[:maxWarmTranslations]
	}

	sem := make(chan struct{}, 8)
	var wg sync.WaitGroup
	for _, name := range unique {
		if _, ok := entityNameTranslations[name]; ok || shouldSkipMachineTranslation(name) {
			continue
		}
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			_ = LocalizeEntityName(name)
		}(name)
	}
	wg.Wait()
}

func WarmMatchTranslations(matches []api.Match) {
	names := make([]string, 0, len(matches)*5)
	for _, match := range matches {
		names = append(names,
			match.League.Name,
			match.HomeTeam.Name,
			match.HomeTeam.ShortName,
			match.AwayTeam.Name,
			match.AwayTeam.ShortName,
		)
	}
	WarmEntityTranslations(names...)
}

func WarmMatchDetailsTranslations(details *api.MatchDetails) {
	if details == nil {
		return
	}

	names := []string{
		details.League.Name,
		details.HomeTeam.Name,
		details.HomeTeam.ShortName,
		details.AwayTeam.Name,
		details.AwayTeam.ShortName,
		details.Venue,
		details.Referee,
	}

	for _, event := range details.Events {
		names = append(names, event.Team.Name, event.Team.ShortName)
		if event.Player != nil {
			names = append(names, *event.Player)
		}
		if event.Assist != nil {
			names = append(names, *event.Assist)
		}
	}
	for _, player := range details.HomeStarting {
		names = append(names, player.Name)
	}
	for _, player := range details.AwayStarting {
		names = append(names, player.Name)
	}
	for _, player := range details.HomeSubstitutes {
		names = append(names, player.Name)
	}
	for _, player := range details.AwaySubstitutes {
		names = append(names, player.Name)
	}

	WarmEntityTranslations(names...)
}

func WarmStandingsTranslations(standings []api.LeagueTableEntry) {
	names := make([]string, 0, len(standings)*2)
	for _, entry := range standings {
		names = append(names, entry.Team.Name, entry.Team.ShortName)
	}
	WarmEntityTranslations(names...)
}

func localizeEntityName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return name
	}
	if !entityLocalizationEnabled {
		return name
	}
	if translated, ok := entityNameTranslations[name]; ok {
		return translated
	}
	return entityTranslator.Translate(name)
}

type translatorCache struct {
	UpdatedAt    time.Time         `json:"updated_at"`
	Translations map[string]string `json:"translations"`
}

type entityTranslatorClient struct {
	client   *http.Client
	once     sync.Once
	mu       sync.RWMutex
	cache    map[string]string
	inflight map[string]*translationCall
	path     string
}

type translationCall struct {
	done  chan struct{}
	value string
}

func newEntityTranslator() *entityTranslatorClient {
	return &entityTranslatorClient{
		client: &http.Client{
			Timeout: 1500 * time.Millisecond,
		},
		cache:    make(map[string]string),
		inflight: make(map[string]*translationCall),
	}
}

func (t *entityTranslatorClient) Translate(name string) string {
	if shouldSkipMachineTranslation(name) {
		return name
	}

	t.once.Do(t.load)

	t.mu.RLock()
	if translated, ok := t.cache[name]; ok {
		t.mu.RUnlock()
		if translated != "" {
			return translated
		}
		return name
	}
	if call, ok := t.inflight[name]; ok {
		t.mu.RUnlock()
		<-call.done
		if call.value != "" {
			return call.value
		}
		return name
	}
	t.mu.RUnlock()

	call := &translationCall{done: make(chan struct{})}

	t.mu.Lock()
	if translated, ok := t.cache[name]; ok {
		t.mu.Unlock()
		if translated != "" {
			return translated
		}
		return name
	}
	if existing, ok := t.inflight[name]; ok {
		t.mu.Unlock()
		<-existing.done
		if existing.value != "" {
			return existing.value
		}
		return name
	}
	t.inflight[name] = call
	t.mu.Unlock()

	translated := t.fetch(name)
	if translated == "" {
		translated = name
	}

	t.mu.Lock()
	t.cache[name] = translated
	delete(t.inflight, name)
	t.mu.Unlock()

	call.value = translated
	close(call.done)

	t.save()
	return translated
}

func shouldSkipMachineTranslation(name string) bool {
	if strings.ContainsAny(name, "的一是在不了有和人这中大为上个国我以要他") {
		return true
	}
	hasLetter := false
	for _, r := range name {
		if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
			hasLetter = true
			break
		}
	}
	if !hasLetter {
		return true
	}
	trimmed := strings.TrimSpace(name)
	return len([]rune(trimmed)) <= 1
}

func (t *entityTranslatorClient) load() {
	cacheDir, err := data.CacheDir()
	if err != nil {
		return
	}
	t.path = filepath.Join(cacheDir, translationCacheFile)

	raw, err := os.ReadFile(t.path)
	if err != nil {
		return
	}

	var disk translatorCache
	if err := json.Unmarshal(raw, &disk); err != nil {
		return
	}
	if disk.Translations == nil {
		return
	}

	t.mu.Lock()
	for source, translated := range disk.Translations {
		if source != "" && translated != "" {
			t.cache[source] = translated
		}
	}
	t.mu.Unlock()
}

func (t *entityTranslatorClient) save() {
	if t.path == "" {
		return
	}

	t.mu.RLock()
	disk := translatorCache{
		UpdatedAt:    time.Now(),
		Translations: make(map[string]string, len(t.cache)),
	}
	for source, translated := range t.cache {
		if source != "" && translated != "" && source != translated {
			disk.Translations[source] = translated
		}
	}
	t.mu.RUnlock()

	raw, err := json.MarshalIndent(disk, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(t.path, raw, 0644)
}

func (t *entityTranslatorClient) fetch(name string) string {
	endpoint := "https://translate.googleapis.com/translate_a/single?client=gtx&sl=auto&tl=zh-CN&dt=t&q=" + url.QueryEscape(name)
	resp, err := t.client.Get(endpoint)
	if err != nil {
		return ""
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if err != nil {
		return ""
	}

	var payload []any
	if err := json.Unmarshal(body, &payload); err != nil {
		return ""
	}
	if len(payload) == 0 {
		return ""
	}

	segments, ok := payload[0].([]any)
	if !ok {
		return ""
	}

	var b strings.Builder
	for _, segment := range segments {
		fields, ok := segment.([]any)
		if !ok || len(fields) == 0 {
			continue
		}
		text, ok := fields[0].(string)
		if ok {
			b.WriteString(text)
		}
	}

	return strings.TrimSpace(b.String())
}

func localizeRegion(region string) string {
	switch region {
	case "Europe":
		return "欧洲"
	case "Americas":
		return "美洲"
	case "Global":
		return "全球"
	default:
		return region
	}
}

var entityNameTranslations = map[string]string{
	"1. Division":                        "丹麦甲级联赛",
	"1. Divisjon":                        "挪威甲级联赛",
	"2. Bundesliga":                      "德乙",
	"3. Liga":                            "德丙",
	"A-League":                           "澳超",
	"AFC Bournemouth":                    "伯恩茅斯",
	"AFC Champions League Elite":         "亚冠精英联赛",
	"Africa Cup of Nations":              "非洲杯",
	"Allsvenskan":                        "瑞典超",
	"Arsenal":                            "阿森纳",
	"Aston Villa":                        "阿斯顿维拉",
	"Austrian Bundesliga":                "奥地利甲级联赛",
	"Belgian First Division":             "比利时甲级联赛",
	"Botola Pro":                         "摩洛哥职业联赛",
	"Brasileirão Série A":                "巴甲",
	"Brasileirão Série B":                "巴乙",
	"Brentford":                          "布伦特福德",
	"Brighton & Hove Albion":             "布莱顿",
	"Bundesliga":                         "德甲",
	"Burnley":                            "伯恩利",
	"CAF Champions League":               "非洲冠军联赛",
	"CONCACAF Champions Cup":             "中北美冠军杯",
	"CONCACAF Gold Cup":                  "中北美金杯赛",
	"CONCACAF Nations League":            "中北美国联",
	"Carioca":                            "里约州锦标赛",
	"Chelsea":                            "切尔西",
	"Chinese League One":                 "中甲",
	"Club Friendlies":                    "俱乐部友谊赛",
	"Copa America":                       "美洲杯",
	"Copa Colombia":                      "哥伦比亚杯",
	"Copa Libertadores":                  "解放者杯",
	"Copa Sudamericana":                  "南美杯",
	"Copa del Rey":                       "国王杯",
	"Copa do Brasil":                     "巴西杯",
	"Coppa Italia":                       "意大利杯",
	"Coupe de France":                    "法国杯",
	"Crystal Palace":                     "水晶宫",
	"DFB Pokal":                          "德国杯",
	"EFL Championship":                   "英冠",
	"EFL League One":                     "英甲",
	"EFL League Two":                     "英乙",
	"Egyptian Premier League":            "埃及超",
	"Ekstraklasa":                        "波兰超",
	"Eliteserien":                        "挪超",
	"Eredivisie":                         "荷甲",
	"Everton":                            "埃弗顿",
	"FA Cup":                             "足总杯",
	"FIFA Club World Cup":                "世俱杯",
	"FIFA World Cup":                     "世界杯",
	"Finalissima":                        "欧美杯",
	"Frauen-Bundesliga":                  "德国女足甲级联赛",
	"Fulham":                             "富勒姆",
	"Indian Super League":                "印度超",
	"International Friendlies":           "国际友谊赛",
	"J. League":                          "J 联赛",
	"K League 1":                         "K 联赛 1",
	"La Liga":                            "西甲",
	"League of Ireland First Division":   "爱尔兰甲级联赛",
	"League of Ireland Premier Division": "爱尔兰超级联赛",
	"Leeds United":                       "利兹联",
	"Liga 1":                             "秘鲁甲级联赛",
	"Liga F":                             "西班牙女足甲级联赛",
	"Liga MX":                            "墨西哥超",
	"Liga Portugal 2":                    "葡甲",
	"Liga Portugal 2 Qualification":      "葡甲附加赛",
	"Liga Profesional":                   "阿根廷职业联赛",
	"Ligue 1":                            "法甲",
	"Ligue 2":                            "法乙",
	"Liverpool":                          "利物浦",
	"MLS":                                "美职联",
	"Manchester City":                    "曼城",
	"Manchester United":                  "曼联",
	"Mineiro":                            "米内罗州锦标赛",
	"NWSL":                               "美国女足联赛",
	"Newcastle United":                   "纽卡斯尔联",
	"Nordeste":                           "巴西东北杯",
	"Nottingham Forest":                  "诺丁汉森林",
	"Old Trafford":                       "老特拉福德",
	"Paulista":                           "圣保罗州锦标赛",
	"Premier League":                     "英超",
	"Premier Soccer League":              "南非超级联赛",
	"Primera A":                          "哥伦比亚甲级联赛",
	"Primera B":                          "哥伦比亚乙级联赛",
	"Primera Division":                   "甲级联赛",
	"Primeira Liga":                      "葡超",
	"Primeira Liga Qualification":        "葡超附加赛",
	"Qatar Stars League":                 "卡塔尔星联赛",
	"Recopa Sudamericana":                "南美优胜者杯",
	"Regionalliga":                       "德国地区联赛",
	"Russian Premier League":             "俄超",
	"Saudi Pro League":                   "沙特职业联赛",
	"Scottish Premiership":               "苏超",
	"Selhurst Park":                      "塞尔赫斯特公园球场",
	"Serie A":                            "意甲",
	"Serie A Femminile":                  "意大利女足甲级联赛",
	"Serie B":                            "意乙",
	"St. James' Park":                    "圣詹姆斯公园球场",
	"Stamford Bridge":                    "斯坦福桥",
	"Sunderland":                         "桑德兰",
	"Super League 1":                     "希腊超",
	"Supercopa de España":                "西班牙超级杯",
	"Supercopa do Brasil":                "巴西超级杯",
	"Superligaen":                        "丹麦超",
	"Swiss Super League":                 "瑞士超",
	"Süper Lig":                          "土超",
	"Taça da Liga":                       "葡萄牙联赛杯",
	"Taça de Portugal":                   "葡萄牙杯",
	"Tottenham Hotspur":                  "托特纳姆热刺",
	"UEFA Champions League":              "欧冠",
	"UEFA Conference League":             "欧协联",
	"UEFA Euro":                          "欧洲杯",
	"UEFA Europa League":                 "欧联杯",
	"UEFA Nations League":                "欧国联",
	"Ukrainian Premier League":           "乌克兰超",
	"West Ham United":                    "西汉姆联",
	"Women's DFB Pokal":                  "德国女足杯",
	"Women's FIFA World Cup":             "女足世界杯",
	"Women's Super League":               "英格兰女超",
	"Women's UEFA Champions League":      "女足欧冠",
	"Women's UEFA Euro":                  "女足欧洲杯",
	"Wolverhampton Wanderers":            "狼队",
}

func localizeCountry(country string) string {
	switch country {
	case "Africa":
		return "非洲"
	case "Argentina":
		return "阿根廷"
	case "Asia":
		return "亚洲"
	case "Australia":
		return "澳大利亚"
	case "Austria":
		return "奥地利"
	case "Belgium":
		return "比利时"
	case "Brazil":
		return "巴西"
	case "Chile":
		return "智利"
	case "China":
		return "中国"
	case "Colombia":
		return "哥伦比亚"
	case "Denmark":
		return "丹麦"
	case "Ecuador":
		return "厄瓜多尔"
	case "Egypt":
		return "埃及"
	case "England":
		return "英格兰"
	case "Europe":
		return "欧洲"
	case "France":
		return "法国"
	case "Germany":
		return "德国"
	case "Greece":
		return "希腊"
	case "India":
		return "印度"
	case "International":
		return "国际"
	case "Ireland":
		return "爱尔兰"
	case "Italy":
		return "意大利"
	case "Japan":
		return "日本"
	case "Mexico":
		return "墨西哥"
	case "Morocco":
		return "摩洛哥"
	case "Netherlands":
		return "荷兰"
	case "North America":
		return "北美"
	case "Norway":
		return "挪威"
	case "Peru":
		return "秘鲁"
	case "Poland":
		return "波兰"
	case "Portugal":
		return "葡萄牙"
	case "Qatar":
		return "卡塔尔"
	case "Russia":
		return "俄罗斯"
	case "Saudi Arabia":
		return "沙特阿拉伯"
	case "Scotland":
		return "苏格兰"
	case "South Africa":
		return "南非"
	case "South America":
		return "南美"
	case "South Korea":
		return "韩国"
	case "Spain":
		return "西班牙"
	case "Sweden":
		return "瑞典"
	case "Switzerland":
		return "瑞士"
	case "Turkey":
		return "土耳其"
	case "Ukraine":
		return "乌克兰"
	case "Uruguay":
		return "乌拉圭"
	case "USA":
		return "美国"
	default:
		return country
	}
}

func localizeStatLabel(label string) string {
	normalized := strings.ToLower(strings.TrimSpace(label))
	normalized = strings.ReplaceAll(normalized, "_", " ")

	switch normalized {
	case "possession", "ball possession", "ballpossesion":
		return "控球率"
	case "total shots":
		return "射门"
	case "shots on target", "on target", "shotsontarget":
		return "射正"
	case "accurate passes":
		return "传球成功"
	case "fouls", "fouls committed":
		return "犯规"
	case "corners", "corner kicks":
		return "角球"
	case "offsides":
		return "越位"
	case "yellow cards":
		return "黄牌"
	case "red cards":
		return "红牌"
	case "big chances":
		return "绝佳机会"
	case "big chances missed":
		return "错失绝佳机会"
	case "xg", "expected goals":
		return "预期进球"
	case "shots":
		return "射门"
	case "passes":
		return "传球"
	case "tackles":
		return "抢断"
	case "saves":
		return "扑救"
	default:
		if label == "" {
			return ""
		}
		return label
	}
}
