package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/peppage/kettle"
)

// CSVでの出力を前提としているので、複数要素ある場合はこの文字で区切る
// 使う人が都合の良い文字にしてください
const splitChar = "."

func main() {

	type CsvStruct struct {
		Title                  string  `csv:"Title"`
		Genre                  string  `csv:Genre`
		Publisher              string  `csv:Publisher`
		SupportedLanguages     string  `csv:SupportedLanguages`
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
		// 最初の待機
		time.Sleep(time.Second)

		println("取得開始, AppId:", appId)

		appDataChan := make(chan *kettle.AppData)
		defer close(appDataChan)

		// 取得
		go GetAppDataByAppId(steamClient, int64(appId), appDataChan)
		appData := <-appDataChan

		if appData.Name == "err" {
			log.Println("タイトル取得失敗なのでスキップ", appId)
			continue
		}

		// 日本語レビュー取得
		time.Sleep(time.Second)
		jaReviewChan := make(chan kettle.QuerySummary)
		go GetStoreReviewByAppId(steamClient, int64(appId), "japanese", jaReviewChan)
		jaReview := <-jaReviewChan
		defer close(jaReviewChan)
		println("ja取得完了", appData.Name)

		// 全言語のレビュー取得
		time.Sleep(time.Second)
		allReviewChan := make(chan kettle.QuerySummary)
		go GetStoreReviewByAppId(steamClient, int64(appId), "all", allReviewChan)
		allReview := <-allReviewChan
		defer close(allReviewChan)
		println("all取得完了", appData.Name)

		csv := CsvStruct{
			Title:                  appData.Name,
			Genre:                  ConvertGenresToString(appData.Genres),
			Publisher:              strings.Join(appData.Publishers, splitChar),
			SupportedLanguages:     SplitSupportedLanguages(appData.SupportedLanguages),
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

// GetAppDataByAppId AppIdをもとにAppDataを取得する
func GetAppDataByAppId(steamClient *kettle.Client, appId int64, appdata chan *kettle.AppData) {
	d, _, err := steamClient.Store.AppDetails(appId)

	if err != nil {
		// エラーが帰ってきたということはサーバ負荷が高い可能性もあるので10秒待つ
		time.Sleep(time.Second * 10)

		fmt.Println("AppDetails Error", appId, err)
		// とりあえずErrを名前にいれとく
		// TODO: エラーハンドリングしやすい形に修正する
		appdata <- &kettle.AppData{
			Name: "Err",
		}
	}

	appdata <- d
}

// GetStoreReviewByAppId ストアレビューを取得する
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

// GetUserInput ユーザからの入力をパースするための受口
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

// ConvertGenresToString Genreの配列をいいかんじの文字列にする 出力先がcsvなので ,(カンマ) はつかわない
func ConvertGenresToString(genres []kettle.Genre) string {
	// TODO: たぶんもっと良い処理があるので改善したい
	genre := ""
	for i, g := range genres {
		genre += g.Description

		// 続きがあるなら区切り文字を追加する
		if i < (len(genres) - 1) {
			genre += splitChar
		}
	}

	return genre
}

// SplitSupportedLanguages English<strong>*</strong>, Japanese<strong>*</strong>, のような形のものを整形する
func SplitSupportedLanguages(languages string) string {
	// + の削除
	replaced := strings.Replace(languages, "<strong>*</strong>", "", -1)

	// 音声も対応してるぜ！の削除
	replaced = strings.Replace(replaced, "<br>languages with full audio support", "", -1)

	// , だと都合が悪いので置き換え
	replaced = strings.Replace(replaced, ",", splitChar, -1)

	// まだ半角スペースがのこってるので掃除
	replaced = strings.Replace(replaced, " ", "", -1)

	return replaced
}

// ReadAppIdsFromCsv []intで返してるけど、実際のAppIdはint64 まぁ上限超えないしいいだろの精神
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
