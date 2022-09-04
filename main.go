package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/peppage/kettle"
)

func main() {

	type CsvStruct struct {
		Title                  string  `csv:"Title"`
		AllTotalReviews        int     `csv:"AllTotalReviews"`
		AllTotalPositive       int     `csv:"AllTotalPositive"`
		AllTotalNegative       int     `csv:"AllTotalNegative"`
		JapaneseTotalReviews   int     `csv:"JapaneseTotalReviews"`
		JapaneseTotalPositive  int     `csv:"JapaneseTotalPositive"`
		JapaneseTotalNegative  int     `csv:"JapaneseTotalNegative"`
		AllNegativeRatio       float64 `csv:"AllNegativeRatio"`
		JapaneseNegativeRatio  float64 `csv:"JapaneseNegativeRatio"`
		JapaneseAllReviewRatio float64 `csv:"JapaneseAllReviewRatio"`
	}

	apiKey := GetUserInput("STEAM API KEY")

	httpClient := http.DefaultClient
	steamClient := kettle.NewClient(httpClient, apiKey)

	// 抽出対象のAppId
	appIds := ReadAppIdsFromCsv()
	var csvSlice []CsvStruct

	// ここからAppIdのループ処理
	for _, appId := range appIds {
		time.Sleep(time.Second)
		println("取得開始, AppId:", appId)
		titleChan := make(chan string)
		go GetTitleByAppId(steamClient, int64(appId), titleChan)
		titleName := <-titleChan
		defer close(titleChan)
		if titleName == "err" {
			log.Println("タイトル取得失敗なのでスキップ", appId)
			continue
		}

		time.Sleep(time.Second)
		jaReviewChan := make(chan kettle.QuerySummary)
		go GetStoreReviewByAppId(steamClient, int64(appId), "japanese", jaReviewChan)
		jaReview := <-jaReviewChan
		defer close(jaReviewChan)
		println("ja取得完了", titleName)

		time.Sleep(time.Second)
		allReviewChan := make(chan kettle.QuerySummary)
		go GetStoreReviewByAppId(steamClient, int64(appId), "all", allReviewChan)
		allReview := <-allReviewChan
		defer close(allReviewChan)
		println("all取得完了", titleName)

		csv := CsvStruct{
			Title:                  titleName,
			AllTotalReviews:        allReview.TotalReviews,
			AllTotalPositive:       allReview.TotalPositive,
			AllTotalNegative:       allReview.TotalNegative,
			JapaneseTotalReviews:   jaReview.TotalReviews,
			JapaneseTotalPositive:  jaReview.TotalPositive,
			JapaneseTotalNegative:  jaReview.TotalNegative,
			AllNegativeRatio:       float64(allReview.TotalNegative) / float64(allReview.TotalReviews),
			JapaneseNegativeRatio:  float64(jaReview.TotalNegative) / float64(jaReview.TotalReviews),
			JapaneseAllReviewRatio: float64(jaReview.TotalReviews) / float64(allReview.TotalReviews),
		}

		csvSlice = append(csvSlice, csv)
	}

	fmt.Println("取得完了")

	file, _ := os.OpenFile("sample.csv", os.O_RDWR|os.O_CREATE, os.ModePerm)
	defer file.Close()

	gocsv.MarshalFile(&csvSlice, file)
}

func GetTitleByAppId(steamClient *kettle.Client, appId int64, c chan string) {
	d, _, err := steamClient.Store.AppDetails(appId)
	if err != nil {
		// エラーが帰ってきたということはサーバ負荷が高い可能性もあるので10秒待つ
		time.Sleep(time.Second * 10)

		fmt.Println("AppDetails Error", appId, err)
		c <- "err"
	}

	c <- d.Name
}

func GetStoreReviewByAppId(steamClient *kettle.Client, appId int64, language string, c chan kettle.QuerySummary) {
	d, _, err := steamClient.Store.AppReviews(&kettle.AppReviewsParams{
		AppID:    appId,
		Language: language,
	})
	if err != nil {
		// エラーが帰ってきたということはサーバ負荷が高い可能性もあるので10秒待つ
		time.Sleep(time.Second * 10)

		// 適切なエラーハンドリングが思いつかなかったので負数を返しています
		fmt.Println("GetStoreReview", err)
		var dummy kettle.QuerySummary
		dummy.NumberReviews = -1
		dummy.ReviewScore = -1
		dummy.ReviewScoreDescription = ""
		dummy.TotalPositive = -1
		dummy.TotalNegative = -1
		dummy.TotalReviews = -1
		c <- dummy
		return
	}

	c <- d.QuerySummary
}

func GetUserInput(caption string) string {
	fmt.Print(caption + ":")
	var inVal string
	_, err := fmt.Scan(&inVal)
	if err != nil {
		println(err)
		return ""
	}
	return inVal
}

// []intで返してるけど、実際のAppIdはint64 まぁ上限超えないしいいだろの精神
func ReadAppIdsFromCsv() []int {
	file, err := os.Open("appId.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// 返すデータ
	var appIds []int

	reader := csv.NewReader(file)
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		appId, _ := strconv.Atoi(record[0])
		appIds = append(appIds, appId)
	}

	return appIds
}
