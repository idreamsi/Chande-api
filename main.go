package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "strconv"
    "strings"
    "sync"
    "time"
	"io"

    "github.com/PuerkitoBio/goquery"
    ptime "github.com/yaa110/go-persian-calendar"
)

type Response struct {
    Gold18   Item `json:"gold18"`
    Silver999 Item `json:"silver999"`
}

type Item struct {
    Price int `json:"price"`
    TTL   int `json:"ttl"`
}

type Currency struct {
    Code  string  `json:"code"`
    Name  string  `json:"name"`
    Price float64 `json:"price"`
    Icon  string  `json:"icon"`
    En    string  `json:"en"`
}

type FinalOutput struct {
    Date       string     `json:"date"`
    Currencies []Currency `json:"currencies"`
}

type Country struct {
    Country string `json:"country"`
    En      string `json:"en"`
}

var currencyMap map[string]string

// 🔑 تبدیل اعداد فارسی به انگلیسی
func faToEnDigits(s string) string {
    replacer := strings.NewReplacer(
        "۰", "0", "۱", "1", "۲", "2", "۳", "3", "۴", "4",
        "۵", "5", "۶", "6", "۷", "7", "۸", "8", "۹", "9",
    )
    return replacer.Replace(s)
}

// تبدیل رشته به عدد
func parseNumber(s string) float64 {
    s = faToEnDigits(s)
    s = strings.ReplaceAll(s, "$", "")
    s = strings.ReplaceAll(s, ",", "")
    s = strings.ReplaceAll(s, "٫", ".")
    s = strings.TrimSpace(s)
    f, _ := strconv.ParseFloat(s, 64)
    return f
}

// تغییر متن به Title Case
func toTitleCase(s string) string {
    return strings.Title(strings.ToLower(strings.TrimSpace(s)))
}

// کریپتو
func fetchCryptoData() ([]Currency, error) {
    resp, err := http.Get("https://alanchand.com/crypto-price")
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    doc, err := goquery.NewDocumentFromReader(resp.Body)
    if err != nil {
        return nil, err
    }

    var cryptoData []Currency
    doc.Find("table.cryptoTbl tbody tr").Each(func(i int, s *goquery.Selection) {
        code := strings.ToLower(strings.TrimSpace(s.Find(".symbolCurr").Text()))
        nameFa := strings.TrimSpace(s.Find(".faCurr").Text())
        nameEn := strings.TrimSpace(s.Find(".enCurr").Text())
        tomanStr := s.Find(".tmn").Text()
        dollarStr := s.Find(".dlr").Text()
        icon, _ := s.Find(".CurrIco").Attr("src")
        if !strings.HasPrefix(icon, "http") {
            icon = "https://alanchand.com" + icon
        }

        toman := parseNumber(tomanStr)
        dollar := parseNumber(dollarStr)

        price := dollar
        if code == "usdt" || code == "dai" {
            price = toman
        }
        if code == "btc" {
            priceStr := fmt.Sprintf("%.0f", price)
            price, _ = strconv.ParseFloat(priceStr, 64)
        }

        cryptoData = append(cryptoData, Currency{
            Code:  code,
            Name:  nameFa,
            Price: price,
            Icon:  icon,
            En:    toTitleCase(nameEn),
        })
    })
    return cryptoData, nil
}

// ارزها
func fetchDataCurrency(cm map[string]string) ([]Currency, error) {
    resp, err := http.Get("https://alanchand.com/currencies-price")
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    doc, err := goquery.NewDocumentFromReader(resp.Body)
    if err != nil {
        return nil, err
    }

    var data []Currency
    doc.Find("table tbody tr").Each(func(i int, s *goquery.Selection) {
        codeAttr, _ := s.Attr("onclick")
        code := strings.TrimPrefix(strings.TrimSuffix(codeAttr, "'"), "window.location='/currencies-price/")
        nameFa := strings.TrimSpace(s.Find("td.currName").Text())
        priceStr := strings.TrimSpace(s.Find("td.sellPrice").Contents().First().Text())

        price := parseNumber(priceStr)

        if code != "" {
            // پرچم توی کلاس i.flag-xx هست
            flag := ""
            if iTag := s.Find("td.currName .flag"); iTag.Length() > 0 {
                for _, cls := range strings.Split(iTag.AttrOr("class", ""), " ") {
                    if strings.HasPrefix(cls, "flag-") {
                        flag = strings.TrimPrefix(cls, "flag-")
                        break
                    }
                }
            }
            // اصلاح آیکون برای یورو
            if flag == "eu" {
                flag = "european_union"
            }
            icon := fmt.Sprintf("https://raw.githubusercontent.com/HatScripts/circle-flags/refs/heads/gh-pages/flags/%s.svg", flag)

            enName := cm[flag]
            if enName == "" {
                enName = toTitleCase(code)
            }

            data = append(data, Currency{
                Code:  code,
                Name:  nameFa,
                Price: price,
                Icon:  icon,
                En:    enName,
            })
        }
    })
    return data, nil
}


// طلا
func fetchDigiGoldData() ([]Currency, error) {

//-----------------------------------------------------------------
    /*
	// 1. یک User-Agent واقعی برای مرورگر ست می‌کنیم
    reqq, _ := http.NewRequest("GET", "https://api.digikala.com/non-inventory/v1/prices/", nil)
    reqq.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
    reqq.Header.Set("Accept", "application/json") // صراحتاً اعلام می‌کنیم که JSON می‌خواهیم

    client := &http.Client{}
    respp, errr := client.Do(reqq)
    if errr != nil {
        fmt.Println("خطا در درخواست:", err)
        //return nil
    }
    defer respp.Body.Close()

    // 2. بررسی کد وضعیت HTTP
    fmt.Println("Status Code:", respp.StatusCode) //  403 یا 200 را نشان می‌دهد

    // 3. بررسی هدرها که سرور چه نوع محتوایی برگردانده
    contentTypee := respp.Header.Get("Content-Type")
    fmt.Println("Content-Type:", contentTypee)

    bodyy, _ := io.ReadAll(respp.Body)
    // 4. نگاهی به ۵۰۰ کاراکتر اول بدنه می‌اندازیم تا ببینیم HTML است یا JSON
    fmt.Println("Body (first 500 chars):", string(bodyy[:min(500, len(bodyy))]))

    if strings.Contains(contentTypee, "application/json") {
		fmt.Println("IS JSON")
    } else {
        fmt.Println("BLOCKED")
    }
	*/
//-----------------------------------------------------------------
	
    // درخواست HTTP GET
    resp, err := http.Get("https://idreams.ir/bus/gold/")
    if err != nil {
		fmt.Println("get: ",err)
		//return nil, err
        //return 0, fmt.Errorf("خطا در ارسال درخواست: %w", err)
    }
    defer resp.Body.Close()

/*
	req, _ := http.NewRequest("GET", "https://idreams.ir/bus/gold/", nil)
    req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
    req.Header.Set("Accept", "application/json") // صراحتاً اعلام می‌کنیم که JSON می‌خواهیم

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        fmt.Println("خطا در درخواست:", err)
        //return nil
    }
    defer resp.Body.Close()
*/
	//---------------------------------------------------
	contentType := resp.Header.Get("Content-Type")
    fmt.Println("Content-Type:", contentType)
	if strings.Contains(contentType, "application/json") {
		fmt.Println("IS JSON")
    } else {
        fmt.Println("IS HTML")
    }
	//---------------------------------------------------
	    // بررسی وضعیت پاسخ
    

    // خواندن بدنه پاسخ
    body, err := io.ReadAll(resp.Body)
    if err != nil {
		fmt.Println("io: ",err)
		//return nil, err
        //return 0, fmt.Errorf("خطا در خواندن بدنه: %w", err)
    }
	/*
if resp.StatusCode != http.StatusOK {
    body, _ := io.ReadAll(resp.Body)
    return fmt.Errorf("وضعیت ناموفق: %s, بدنه: %s", resp.Status, string(body))
}*/
if strings.Contains(resp.Header.Get("Content-Type"), "application/json") == false {
    fmt.Println("html: ",err)
	//fmt.Println("body: ",string(body))
	
	// شاید HTML است
}
    // تجزیه JSON
    var data2 Response
    err = json.Unmarshal(body, &data2)
    if err != nil {
		fmt.Println("json: ",err)
		//return nil, err
        //return 0, fmt.Errorf("خطا در تجزیه JSON: %w", err)
    }

    // بازگرداندن قیمت gold18
    //return data.Gold18.Price, nil
fmt.Println("Test")
fmt.Println(float64(data2.Gold18.Price))
    var data []Currency
	data = append(data, Currency{
        Code:  "DigiGold",
        Name:  "طلای دیجی کالا",
        Price: float64(data2.Gold18.Price),
        Icon:  "https://www.digikala.com/wealth/static/img/svg/gold-logo.svg",
        En:    "DigiGold",
        })
    return data, nil
}


// طلا
func fetchGoldData() ([]Currency, error) {
    // مپ آیکون‌ها
    var goldIcons = map[string]string{
        "abshodeh": "https://platform.tgju.org/files/images/gold-bar-1622253729.png",
        "18ayar":   "https://platform.tgju.org/files/images/gold-bar-1-1622253841.png",
        "sekkeh":   "https://platform.tgju.org/files/images/gold-1697963730.png",
        "bahar":    "https://platform.tgju.org/files/images/gold-1-1697963918.png",
        "nim":      "https://platform.tgju.org/files/images/money-1697964123.png",
        "rob":      "https://platform.tgju.org/files/images/revenue-1697964369.png",
        "sek":      "https://platform.tgju.org/files/images/parsian-coin-1697964860.png",
        "usd_xau": "https://platform.tgju.org/files/images/gold-1-1622253769.png",
        "xag": "https://platform.tgju.org/files/images/silver-1624079710.png",
    }

    resp, err := http.Get("https://alanchand.com/gold-price")
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    doc, err := goquery.NewDocumentFromReader(resp.Body)
    if err != nil {
        return nil, err
    }

    var data []Currency
    doc.Find("table tbody tr").Each(func(i int, s *goquery.Selection) {
        codeAttr, _ := s.Attr("onclick")
        code := strings.TrimPrefix(strings.TrimSuffix(codeAttr, "'"), "window.location='/gold-price/")
        nameFa := strings.TrimSpace(s.Find("td").First().Text())
        priceStr := strings.TrimSpace(s.Find("td.priceTd").First().Contents().First().Text())

        price := parseNumber(priceStr)

        icon := ""
        if val, exists := goldIcons[code]; exists {
            icon = val
        }

        if code != "" {
            codeName := code
            if codeName == "sek" {
                codeName = "gram"
            }

            data = append(data, Currency{
                Code:  codeName,
                Name:  nameFa,
                Price: price,
                Icon:  icon,
                En:    toTitleCase(code),
            })
        }
    })
    return data, nil
}

// تاریخ جلالی
func getJalaliTime() string {
    loc, _ := time.LoadLocation("Asia/Tehran")
    now := time.Now().In(loc)
    jalali := ptime.New(now)
    return fmt.Sprintf("%04d/%02d/%02d, %02d:%02d",
        jalali.Year(), jalali.Month(), jalali.Day(),
        now.Hour(), now.Minute())
}

func loadCurrencyMap() error {
    data, err := os.ReadFile("currencies.json")
    if err != nil {
        return err
    }

    var countries []Country
    err = json.Unmarshal(data, &countries)
    if err != nil {
        return err
    }

    currencyMap = make(map[string]string)
    for _, c := range countries {
        currencyMap[c.Country] = c.En
    }
    return nil
}

func main() {
    err := loadCurrencyMap()
    if err != nil {
        fmt.Println("Error loading currencies.json:", err)
        return
    }

    var wg sync.WaitGroup
    var currencies, golds, cryptos, digigold []Currency
    var err0, err1, err2, err3 error

    wg.Add(4)
	go func() {
        defer wg.Done()
        digigold, err0 = fetchDigiGoldData()
    }()
    go func() {
        defer wg.Done()
        currencies, err1 = fetchDataCurrency(currencyMap)
    }()
    go func() {
        defer wg.Done()
        golds, err2 = fetchGoldData()
    }()
    go func() {
        defer wg.Done()
        cryptos, err3 = fetchCryptoData()
    }()
    wg.Wait()

    if err0 != nil {
        fmt.Println("Error digi:", err0)
    }
	if err1 != nil {
        fmt.Println("Error currency:", err1)
    }
    if err2 != nil {
        fmt.Println("Error gold:", err2)
    }
    if err3 != nil {
        fmt.Println("Error crypto:", err3)
    }

    finalData := append(append(currencies, golds...), cryptos...)
    finalData = append(digigold, finalData...)


    output := FinalOutput{
        Date:       getJalaliTime(),
        Currencies: finalData,
    }

    jsonData, _ := json.MarshalIndent(output, "", "  ")
    _ = os.WriteFile("arz.json", jsonData, 0644)
    fmt.Println("✅ arz.json ساخته شد")
}

/*
package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "strconv"
    "strings"
    "sync"
    "time"

    "github.com/PuerkitoBio/goquery"
    ptime "github.com/yaa110/go-persian-calendar"
)

type Currency struct {
    Code  string  `json:"code"`
    Name  string  `json:"name"`
    Price float64 `json:"price"`
    Icon  string  `json:"icon"`
    En    string  `json:"en"`
}

type FinalOutput struct {
    Date       string     `json:"date"`
    Currencies []Currency `json:"currencies"`
}

type Country struct {
    Country string `json:"country"`
    En      string `json:"en"`
}

var currencyMap map[string]string

// 🔑 تبدیل اعداد فارسی به انگلیسی
func faToEnDigits(s string) string {
    replacer := strings.NewReplacer(
        "۰", "0", "۱", "1", "۲", "2", "۳", "3", "۴", "4",
        "۵", "5", "۶", "6", "۷", "7", "۸", "8", "۹", "9",
    )
    return replacer.Replace(s)
}

// تبدیل رشته به عدد
func parseNumber(s string) float64 {
    s = faToEnDigits(s)
    s = strings.ReplaceAll(s, "$", "")
    s = strings.ReplaceAll(s, ",", "")
    s = strings.ReplaceAll(s, "٫", ".")
    s = strings.TrimSpace(s)
    f, _ := strconv.ParseFloat(s, 64)
    return f
}

// تغییر متن به Title Case
func toTitleCase(s string) string {
    return strings.Title(strings.ToLower(strings.TrimSpace(s)))
}

// کریپتو
func fetchCryptoData() ([]Currency, error) {
    resp, err := http.Get("https://alanchand.com/crypto-price")
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    doc, err := goquery.NewDocumentFromReader(resp.Body)
    if err != nil {
        return nil, err
    }

    var cryptoData []Currency
    doc.Find("table.cryptoTbl tbody tr").Each(func(i int, s *goquery.Selection) {
        code := strings.ToLower(strings.TrimSpace(s.Find(".symbolCurr").Text()))
        nameFa := strings.TrimSpace(s.Find(".faCurr").Text())
        nameEn := strings.TrimSpace(s.Find(".enCurr").Text())
        tomanStr := s.Find(".tmn").Text()
        dollarStr := s.Find(".dlr").Text()
        icon, _ := s.Find(".CurrIco").Attr("src")
        if !strings.HasPrefix(icon, "http") {
            icon = "https://alanchand.com" + icon
        }

        toman := parseNumber(tomanStr)
        dollar := parseNumber(dollarStr)

        price := dollar
        if code == "usdt" || code == "dai" {
            price = toman
        }
        if code == "btc" {
            priceStr := fmt.Sprintf("%.0f", price)
            price, _ = strconv.ParseFloat(priceStr, 64)
        }

        cryptoData = append(cryptoData, Currency{
            Code:  code,
            Name:  nameFa,
            Price: price,
            Icon:  icon,
            En:    toTitleCase(nameEn),
        })
    })
    return cryptoData, nil
}

// ارزها
func fetchDataCurrency(cm map[string]string) ([]Currency, error) {
    resp, err := http.Get("https://alanchand.com/currencies-price")
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    doc, err := goquery.NewDocumentFromReader(resp.Body)
    if err != nil {
        return nil, err
    }

    var data []Currency
    doc.Find("table tbody tr").Each(func(i int, s *goquery.Selection) {
        codeAttr, _ := s.Attr("onclick")
        code := strings.TrimPrefix(strings.TrimSuffix(codeAttr, "'"), "window.location='/currencies-price/")
        nameFa := strings.TrimSpace(s.Find("td.currName").Text())
        priceStr := strings.TrimSpace(s.Find("td.sellPrice").Contents().First().Text())

        price := parseNumber(priceStr)

        if code != "" {
            // پرچم توی کلاس i.flag-xx هست
            flag := ""
            if iTag := s.Find("td.currName .flag"); iTag.Length() > 0 {
                for _, cls := range strings.Split(iTag.AttrOr("class", ""), " ") {
                    if strings.HasPrefix(cls, "flag-") {
                        flag = strings.TrimPrefix(cls, "flag-")
                        break
                    }
                }
            }
            // اصلاح آیکون برای یورو
            if flag == "eu" {
                flag = "european_union"
            }
            icon := fmt.Sprintf("https://raw.githubusercontent.com/HatScripts/circle-flags/refs/heads/gh-pages/flags/%s.svg", flag)

            enName := cm[flag]
            if enName == "" {
                enName = toTitleCase(code)
            }

            data = append(data, Currency{
                Code:  code,
                Name:  nameFa,
                Price: price,
                Icon:  icon,
                En:    enName,
            })
        }
    })
    return data, nil
}

// طلا
func fetchGoldData() ([]Currency, error) {
    // مپ آیکون‌ها
    var goldIcons = map[string]string{
        "abshodeh": "https://platform.tgju.org/files/images/gold-bar-1622253729.png",
        "18ayar":   "https://platform.tgju.org/files/images/gold-bar-1-1622253841.png",
        "sekkeh":   "https://platform.tgju.org/files/images/gold-1697963730.png",
        "bahar":    "https://platform.tgju.org/files/images/gold-1-1697963918.png",
        "nim":      "https://platform.tgju.org/files/images/money-1697964123.png",
        "rob":      "https://platform.tgju.org/files/images/revenue-1697964369.png",
        "sek":      "https://platform.tgju.org/files/images/parsian-coin-1697964860.png",
        "usd_xau": "https://platform.tgju.org/files/images/gold-1-1622253769.png",
        "xag": "https://platform.tgju.org/files/images/silver-1624079710.png",
    }

    resp, err := http.Get("https://alanchand.com/gold-price")
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    doc, err := goquery.NewDocumentFromReader(resp.Body)
    if err != nil {
        return nil, err
    }

    var data []Currency
    doc.Find("table tbody tr").Each(func(i int, s *goquery.Selection) {
        codeAttr, _ := s.Attr("onclick")
        code := strings.TrimPrefix(strings.TrimSuffix(codeAttr, "'"), "window.location='/gold-price/")
        nameFa := strings.TrimSpace(s.Find("td").First().Text())
        priceStr := strings.TrimSpace(s.Find("td.priceTd").First().Contents().First().Text())

        price := parseNumber(priceStr)

        icon := ""
        if val, exists := goldIcons[code]; exists {
            icon = val
        }

        if code != "" {
            codeName := code
            if codeName == "sek" {
                codeName = "gram"
            }

            data = append(data, Currency{
                Code:  codeName,
                Name:  nameFa,
                Price: price,
                Icon:  icon,
                En:    toTitleCase(code),
            })
        }
    })
    return data, nil
}

// تاریخ جلالی
func getJalaliTime() string {
    loc, _ := time.LoadLocation("Asia/Tehran")
    now := time.Now().In(loc)
    jalali := ptime.New(now)
    return fmt.Sprintf("%04d/%02d/%02d, %02d:%02d",
        jalali.Year(), jalali.Month(), jalali.Day(),
        now.Hour(), now.Minute())
}

func loadCurrencyMap() error {
    data, err := os.ReadFile("currencies.json")
    if err != nil {
        return err
    }

    var countries []Country
    err = json.Unmarshal(data, &countries)
    if err != nil {
        return err
    }

    currencyMap = make(map[string]string)
    for _, c := range countries {
        currencyMap[c.Country] = c.En
    }
    return nil
}

func main() {
    err := loadCurrencyMap()
    if err != nil {
        fmt.Println("Error loading currencies.json:", err)
        return
    }

    var wg sync.WaitGroup
    var currencies, golds, cryptos, []Currency
    var err1, err2, err3 error

    wg.Add(3)
    go func() {
        defer wg.Done()
        currencies, err1 = fetchDataCurrency(currencyMap)
    }()
    go func() {
        defer wg.Done()
        golds, err2 = fetchGoldData()
    }()
    go func() {
        defer wg.Done()
        cryptos, err3 = fetchCryptoData()
    }()
    wg.Wait()

    if err1 != nil {
        fmt.Println("Error currency:", err1)
    }
    if err2 != nil {
        fmt.Println("Error gold:", err2)
    }
    if err3 != nil {
        fmt.Println("Error crypto:", err3)
    }

    finalData := append(append(currencies, golds...), cryptos...)

    output := FinalOutput{
        Date:       getJalaliTime(),
        Currencies: finalData,
    }

    jsonData, _ := json.MarshalIndent(output, "", "  ")
    _ = os.WriteFile("arz.json", jsonData, 0644)
    fmt.Println("✅ arz.json ساخته شد")
}
*/
