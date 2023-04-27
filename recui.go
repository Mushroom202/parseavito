// url for search img in yandex.ru https://yandex.ru/images/search?rpt=imageview&source=collections&url=urltofile
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gocolly/colly"
	_ "github.com/lib/pq"
)

const (
	host          = "localhost"
	port          = 5432
	user          = "postgres"
	password      = "132"
	dbname        = "postgres"
	scraperUrlApi = "http://api.scraperapi.com?api_key=72b06c509284c5689e1cd3bab9a7a1a7&url="
)

var (
	mainurl           string = "https://www.avito.ru/"
	cityForSelect     string = "москва"
	categoryForSelect string = "одежда"
	maxpage           int    = 1
)
var city map[string]string = map[string]string{"москва": "moskva"}
var category map[string]string = map[string]string{"одежда": "odezhda_obuv_aksessuary"}

// api + main doman + city + category + ?p=[page]
var url string = fmt.Sprintf("%v%v%v/%v?p=", scraperUrlApi, mainurl, city[cityForSelect], category[categoryForSelect])

type goods struct {
	name  string
	price string
	urls  string
	img   string
}

type FakeBrowserHeadersResponse struct {
	Result []map[string]string `json:"result"`
}

func RandomHeader(headersList []map[string]string) map[string]string {
	randomIndex := rand.Intn(len(headersList))
	return headersList[randomIndex]
}

func GetHeadersList() []map[string]string {

	// ScrapeOps Browser Headers API Endpint
	scrapeopsAPIKey := "3e3c6bc6-0b4e-40c4-a6b0-a54e3df02823"
	scrapeopsAPIEndpoint := "http://headers.scrapeops.io/v1/browser-headers?api_key=" + scrapeopsAPIKey

	req, _ := http.NewRequest("GET", scrapeopsAPIEndpoint, nil)
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Make Request
	resp, err := client.Do(req)
	if err == nil && resp.StatusCode == 200 {
		defer resp.Body.Close()

		// Convert Body To JSON
		var fakeBrowserHeadersResponse FakeBrowserHeadersResponse
		json.NewDecoder(resp.Body).Decode(&fakeBrowserHeadersResponse)
		return fakeBrowserHeadersResponse.Result
	}

	var emptySlice []map[string]string
	return emptySlice
}
func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}
func avito() {

	// connection string
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	// open database
	db, err := sql.Open("postgres", psqlconn)
	CheckError(err)

	// close database
	defer db.Close()

	// check db
	err = db.Ping()
	CheckError(err)

	fmt.Println("Connected!")

	//var listof []string
	headersList := GetHeadersList()
	//initialision of collector
	c := colly.NewCollector()
	//Navigate on site pages
	c.OnRequest(func(r *colly.Request) {
		randomHeader := RandomHeader(headersList)
		for key, value := range randomHeader {
			r.Headers.Set(key, value)
		}
		fmt.Println("Scraping:", r.URL)
	})
	//code of connect to the site
	c.OnResponse(func(r *colly.Response) {
		fmt.Println("Status:", r.StatusCode)

	})
	c.OnHTML("div.iva-item-content-rejJg", func(h *colly.HTMLElement) {

		getgoods := goods{
			name:  h.ChildText("div.iva-item-content-rejJg>div.iva-item-body-KLUuy>div.iva-item-titleStep-pdebR>a>h3"),
			price: h.ChildText("div.iva-item-content-rejJg>div.iva-item-body-KLUuy>div.iva-item-priceStep-uq2CQ>span>span"),
			urls:  "https://www.avito.ru" + h.ChildAttr("div.iva-item-content-rejJg>div.iva-item-body-KLUuy>div.iva-item-titleStep-pdebR>a", "href"),
			img:   h.ChildAttr("div.iva-item-slider-pYwHo>a>div>div>ul>li>div img", "src"),
		}
		insertDynStmt := `insert into "avito"("name", "price", "url", "img") values($1, $2, $3, $4)`
		_, e := db.Exec(insertDynStmt, getgoods.name, getgoods.price, getgoods.urls, getgoods.img)
		/*
			page, err := strconv.Atoi(h.ChildText("a.styles-module-item_size_s-hBQnY>.styles-module-text_size_s-LNY0Q"))
			if err != nil {
				log.Fatal(err)
			}
		*/
		//fmt.Println("Колличество страниц: ", page%1000)
		//fmt.Println(h.ChildAttr("div.iva-item-slider-pYwHo>a>div>div>ul>li", "data-marker"))
		//fmt.Printf("Name:%v,\nPrice: %v,\nUrls: %v,\nImg:%v\n//////////////////////////\n", getgoods.name, getgoods.price, getgoods.urls, getgoods.img)
		CheckError(e)
	})
	c.OnHTML("ul.styles-module-root-OK422", func(h *colly.HTMLElement) {
		max, err := strconv.Atoi(h.ChildText(".styles-module-listItem_last-_ZfSe"))
		if err != nil {
			log.Fatal(err)
		}
		maxpage = max
		//fmt.Println(1 >= maxpage)
	})
	//c.Visit("http://api.scraperapi.com?api_key=72b06c509284c5689e1cd3bab9a7a1a7&url=https://www.avito.ru/")
	//c.Visit(fmt.Sprintf("%v%v", url, 1))
	for i := 1; i <= maxpage; i++ {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		sec := r.Intn(2) + 1
		time.Sleep(time.Duration(sec) * time.Second)
		fmt.Println(maxpage, i, sec)
		c.Visit(fmt.Sprintf("%v%v", url, i))
	}

}
func main() {
	avito()

}
