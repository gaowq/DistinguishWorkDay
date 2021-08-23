package main

import (
	"context"
	"fmt"
	"html/template"
	"os"
	"strconv"
	"time"
	//"github.com/jackdanger/collectlinks"
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

var map1 map[string]YmData = map[string]YmData{}

func main() {

	//quit := make(chan int)
	//mux := http.NewServeMux()
	//mux.Handle("/", HttpHander{})

	//获取当前路径
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(wd+"/../static"))))
	http.HandleFunc("/edit/", editHandler)
	http.HandleFunc("/getDateType/", getDateHandler)
	http.ListenAndServe(":8080", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

type Page struct {
	Title string
	Body  []byte
}

func loadPage(title string) (*Page, error) {
	filename := "../template/" + title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func getDateHandler(writer http.ResponseWriter, request *http.Request) {
	query := request.URL.Query()
	date1 := query.Get("date")
	//t.Execute(w, p)
	//fmt.Println(date1)

	//日期转化为时间戳
	timeLayout := "2006-01-02"           //转化所需模板
	loc, _ := time.LoadLocation("Local") //获取时区
	tmp, _ := time.ParseInLocation(timeLayout, date1, loc)

	if tmp.Year() > 1 {

		result := dynamicBeforeHandle(tmp)
		fmt.Fprintf(writer, `{"res":"%s"}`, result)
		//fmt.Fprintf(writer, `{"res":"%s"}`, tmp.Format("2006年01月"))
	}
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[len("/edit/"):]
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	t, _ := template.ParseFiles("../template/edit.html")
	t.Execute(w, p)
}

func staticHandle() {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "http://wwww.baidu.com/s?ie=UTF-8&wd=日期", nil)

	req.Header.Set("User-Agent", "Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 5.1)")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("http get error", err)
		return
	}
	//函数结束后关闭相关链接
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("read error", err)
		return
	}

	dom, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		log.Fatalln(err)
	}

	dom.Find("span").Each(func(i int, selection *goquery.Selection) {
		fmt.Println(selection.Text())
	})
}

func dynamicBeforeHandle(searchTime time.Time) string {
	mapKey := searchTime.Format("2006-01")

	//新增数据
	if _, ok := map1[mapKey]; !ok {
		dynamicHandle(searchTime)
	}

	result := ""
	day := searchTime.Day()

	weekStr := searchTime.Weekday().String()
	fmt.Println(weekStr)

	switch weekStr {
	case "Monday":
		result = "工作日"
	case "Tuesday":
		result = "工作日"
	case "Wednesday":
		result = "工作日"
	case "Thursday":
		result = "工作日"
	case "Friday":
		result = "工作日"
	case "Saturday":
		result = "休息日"
	case "Sunday":
		result = "休息日"
	default:
		result = ""
	}

	for i := 0; i < len(map1[mapKey].works); i++ {
		if map1[mapKey].works[i] == day {
			result = "工作日"
		}
	}

	for i := 0; i < len(map1[mapKey].rests); i++ {
		if map1[mapKey].rests[i] == day {
			result = "休息日"
		}
	}

	return result
}

func dynamicHandle(searchTime time.Time) {
	strYm := searchTime.Format("2006年01月")
	url := fmt.Sprintf("http://wwww.baidu.com/s?ie=UTF-8&wd=%s", strYm)

	options := []chromedp.ExecAllocatorOption{
		chromedp.Flag("headless", true),
		chromedp.Flag("blink-settings", "imageEnable=true"),
		chromedp.UserAgent(`Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.131 Safari/537.36`),
	}

	options = append(chromedp.DefaultExecAllocatorOptions[:], options...)

	c, _ := chromedp.NewExecAllocator(context.Background(), options...)

	chromeCtx, cancel := chromedp.NewContext(c, chromedp.WithLogf(log.Printf))
	chromedp.Run(chromeCtx, make([]chromedp.Action, 0, 1)...)

	timeOutCtx, cancel := context.WithTimeout(chromeCtx, 60*time.Second)
	defer cancel()

	fmt.Println("start")
	log.Printf("chrome visit page%sn", url)
	var htmlContent string
	err := chromedp.Run(timeOutCtx,
		chromedp.Navigate(url),
		//需要爬取的网页的url
		//chromedp.WaitVisible(`div[class="List-item"]`),
		chromedp.WaitVisible(`span[class="op-calendar-pc-daynumber"]`),
		//等待某个特定的元素出现
		chromedp.OuterHTML(`document.querySelector("body")`, &htmlContent, chromedp.ByJSPath),
		//生成最终的html文件并保存在htmlContent文件中
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("end")

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		log.Fatal(err)
	}

	//list1 := list.New()
	//list2 := list.New()

	var slice1 []int
	var slice2 []int

	//休息日
	fmt.Println("rest:")
	doc.Find("div.op-calendar-pc-relative>a.op-calendar-pc-table-rest:not(.op-calendar-pc-table-other-month)>.op-calendar-pc-table-holiday-sign+.op-calendar-pc-daynumber").Each(func(i int, selection *goquery.Selection) {
		fmt.Println(selection.Text())

		num1, err1 := strconv.Atoi(selection.Text())
		if err1 != nil {
			log.Fatal(err1)
		}

		slice1 = append(slice1, num1)
		//list1.PushBack(num1)
		//list1.PushBack(parseInt(selection.Text()))
	})

	//op-calendar-pc-table-festival
	fmt.Println("work:")
	doc.Find("div.op-calendar-pc-relative>a.op-calendar-pc-table-work:not(.op-calendar-pc-table-other-month)>.op-calendar-pc-table-holiday-sign+.op-calendar-pc-daynumber").Each(func(i int, selection *goquery.Selection) {
		fmt.Println(selection.Text())

		num2, err2 := strconv.Atoi(selection.Text())
		if err2 != nil {
			log.Fatal(err2)
		}

		//list2.PushBack(num2)
		slice2 = append(slice2, num2)
	})
	//doc.Find("div.op-calendar-pc-relative>a>.op-calendar-pc-table-holiday-sign+.op-calendar-pc-daynumber").Each(func(i int, selection *goquery.Selection) {
	//	fmt.Println(selection.Text())
	//})

	//op-calendar-pc-table-other-month

	//fmt.Println("gap1")
	//
	//doc.Find(".op-calendar-pc-daynumber").Each(func(i int, selection *goquery.Selection) {
	//	fmt.Println(selection.Text())
	//})
	//doc.Find("span").Each(func(i int, selection *goquery.Selection) {
	//	fmt.Println(selection.Text())
	//})

	ymData := new(YmData)
	ymData.rests = slice1
	ymData.works = slice2
	ymData.updatetime = time.Now()
	mapKey := searchTime.Format("2006-01")
	map1[mapKey] = *ymData
}

type YmData struct {
	//ym         string
	rests      []int
	works      []int
	updatetime time.Time
}
