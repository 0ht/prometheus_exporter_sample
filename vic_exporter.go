package main

import (
        "flag"
        "log"
        "net/http"
        "github.com/prometheus/client_golang/prometheus"
        "github.com/prometheus/client_golang/prometheus/promhttp"
        "io/ioutil"
        "encoding/json"
		"github.com/koron/go-dproxy"
		"strconv"
		"time"
		"crypto/tls"
		"fmt"
)

// Metricsの定義
const (
    namespace = "vic_temperature"
)

//構造体構造体は下記のように type と structで定義。関連する変数をひとまとめにする
type vicCollector struct{} // 今回働いてくれるインスタンス

var (
	//コマンドライン引数をString型でフラグを定義、Parseでそれぞれの変数に取得
	//引数名、デフォルト値、ヘルプの文字列
    addr = flag.String("listen-address", ":9080", "The address to listen on for HTTP requests.")
	jst = time.FixedZone("Asia/Tokyo", 9*60*60)
	desc_temperature1 = prometheus.NewDesc(
		"vic_temperature1",
		"vic temperature1 at that timestamp",
		[]string{"vicId"},
		nil,
	)
	desc_temperature2 = prometheus.NewDesc(
		"vic_temperature2",
		"vic temperature2 at that timestamp",
		[]string{"vicId"},
		nil,
	)
	desc_temperature3 = prometheus.NewDesc(
		"vic_temperature3",
		"vic temperature3 at that timestamp",
		[]string{"vicId"},
		nil,
	)
	desc_humidity = prometheus.NewDesc(
		"vic_humidity",
		"vic humidity at that timestamp",
		[]string{"vicId"},
		nil,
	)
	desc_co2 = prometheus.NewDesc(
		"vic_co2",
		"vic co2 at that timestamp",
		[]string{"vicId"},
		nil,
	)
	desc_x = prometheus.NewDesc(
		"vic_x",
		"vic x at that timestamp",
		[]string{"vicId"},
		nil,
	)
	desc_y = prometheus.NewDesc(
		"vic_y",
		"vic y at that timestamp",
		[]string{"vicId"},
		nil,
	)
	desc_z = prometheus.NewDesc(
		"vic_z",
		"vic z at that timestamp",
		[]string{"vicId"},
		nil,
	)

	// Gauge metrics define
    vicInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:	"vicInfo",
            Help:	"vicInfo",
		},
		[]string{"vicId"},
	)
)

// Describeというメソッドを定義、cをレシーバという
func (c vicCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- desc_temperature1
	ch <- desc_temperature2
	ch <- desc_temperature3
	ch <- desc_humidity
	ch <- desc_co2
	ch <- desc_x
	ch <- desc_y
	ch <- desc_z
}

//Prometheusのカスタムコレクタの核心部であるCollectメソッド
//ターゲットのアプリケーションインスタンスから必要なデータを全て取り出して適宜マンジングしてクライアントライブラリに送り返す
func (c vicCollector) Collect(ch chan<- prometheus.Metric) {
	_vicId, _serial_no, _timestamp, _temperature1, _temperature2, _temperature3, _humidity, _co2, _x, _y, _z := getCrbInfo()
	//t := time.Date(2020, 3, 10, 10, 52, 22, 123456789, time.Local)
	fmt.Println("2 JST" + _timestamp)
	//fmt.Println(t)
	//jst := time.FixedZone("Asia/Tokyo", 9*60*60)
	__timestampJST, _ := time.Parse("2006/01/02 15:04:05", _timestamp )
	//__timestampJST := __timestampUTC.In(jst).Add(time.Duration(-9) * time.Hour)
	__timestampUTC := __timestampJST.Add(time.Duration(-9) * time.Hour)
	//__timestamp := time.Now().Add(time.Duration(-1) * time.Hour)
	fmt.Println(__timestampUTC)
	fmt.Println("---------------"+_serial_no)
	//fmt.Println(time.Now())
	//vicInfo.WithLabelValues(_vicId,_serial_no).Set(_temperature1)
	//time.Sleep(10 * time.Second)

		ch <- prometheus.NewMetricWithTimestamp(
			__timestampUTC,
			prometheus.MustNewConstMetric(
				desc_temperature1,
				prometheus.GaugeValue,
				float64(_temperature1),
				_vicId, 
			),
		)

		ch <- prometheus.NewMetricWithTimestamp(
			__timestampUTC,
			prometheus.MustNewConstMetric(
				desc_temperature2,
				prometheus.GaugeValue,
				float64(_temperature2),
				_vicId, 
			),
		)

		ch <- prometheus.NewMetricWithTimestamp(
			__timestampUTC,
			prometheus.MustNewConstMetric(
				desc_temperature3,
				prometheus.GaugeValue,
				float64(_temperature3),
				_vicId, 
			),
		)

		ch <- prometheus.NewMetricWithTimestamp(
			__timestampUTC,
			prometheus.MustNewConstMetric(
				desc_humidity,
				prometheus.GaugeValue,
				float64(_humidity),
				_vicId, 
			),
		)

		ch <- prometheus.NewMetricWithTimestamp(
			__timestampUTC,
			prometheus.MustNewConstMetric(
				desc_co2,
				prometheus.GaugeValue,
				float64(_co2),
				_vicId, 
			),
		)

		ch <- prometheus.NewMetricWithTimestamp(
			__timestampUTC,
			prometheus.MustNewConstMetric(
				desc_x,
				prometheus.GaugeValue,
				float64(_x),
				_vicId, 
			),
		)

		ch <- prometheus.NewMetricWithTimestamp(
			__timestampUTC,
			prometheus.MustNewConstMetric(
				desc_y,
				prometheus.GaugeValue,
				float64(_y),
				_vicId, 
			),
		)

		ch <- prometheus.NewMetricWithTimestamp(
			__timestampUTC,
			prometheus.MustNewConstMetric(
				desc_z,
				prometheus.GaugeValue,
				float64(_z),
				_vicId, 
			),
		)
}

// scrape api & parse result json
func getCrbInfo() (string, string, string, float64, float64, float64, float64, float64, float64, float64, float64) {
        url := "https://ec2-13-114-249-134.ap-northeast-1.compute.amazonaws.com/api/data_reference/trakingdata_get"

		transport := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		
		client := &http.Client{
			Transport: transport,
		}
		
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Fatal(err)
		}

		//取得範囲の指定
		//1. 直近１０秒間
		_timestamp_to := time.Now().In(jst).Format("2006/01/02 15:04:05")
		_timestamp_from := time.Now().In(jst).Add(time.Duration(-10) * time.Second).Format("2006/01/02 15:04:05")
		//2. 全量取得(するけれども最初の１エントリのみ表示)
		//_timestamp_to := time.Now().Format("2006/01/02 15:04:05")
		//_timestamp_from := "2020/03/04 00:00:00"
		
		//確認用
		//fmt.Println(_timestamp_from)
		//fmt.Println(_timestamp_to)

		//UTLクエリパラメータの設定
		params := req.URL.Query()
		params.Add("crb_id","0123456789ABCDEFFEDA")
		params.Add("startDate", _timestamp_from)
		params.Add("endDate", _timestamp_to)
		req.URL.RawQuery = params.Encode()

		//Basic認証
		req.SetBasicAuth("crb_iot", "D2018_0256")

		//fmt.Println(req.URL.String())
		res, err_http := client.Do(req)
		//fmt.Println(res)

		//関数が複数の戻り値を返せるので、通常応答とエラーの場合の応答を2つ書いている
		//res, err_http := http.Get(url)
		//err_httpがnilでない場合、つまりエラーが発生している場合の処理
        if err_http != nil {
				//panicとはプログラムの継続的な実行が難しく、どうしよもなくなった時にプログラムを強制的に終了させるために発生するエラー
                panic(err_http.Error())
		}

		//引数の io.Reader から内容を全て読み込んでバイトスライスとして返す
        body, _ := ioutil.ReadAll(res.Body)
		fmt.Println("1 " + string(body))

		//interface{}型 -> どんな型も格納できる特殊な型・型チェックや型変換などに使える
        var tempData interface{}
        err_json := json.Unmarshal(body, &tempData)
        if err_json != nil {
                panic(err_json.Error())
        }

		vicId, _        := dproxy.New(tempData).A(0).M("crbId").String()
		serial_no, _    := dproxy.New(tempData).A(0).M("serial_no").String()
		timestamp, _    := dproxy.New(tempData).A(0).M("timestamp").String() //"timestamp": "2020/03/04 13:52:22",
		temperature1, _ := dproxy.New(tempData).A(0).M("temperature1").String()
		temperature2, _ := dproxy.New(tempData).A(0).M("temperature2").String()
		temperature3, _ := dproxy.New(tempData).A(0).M("temperature3").String()
		humidity, _     := dproxy.New(tempData).A(0).M("humidity").Float64()
		co2, _          := dproxy.New(tempData).A(0).M("co2").Float64()
		x, _            := dproxy.New(tempData).A(0).M("x").Float64()
		y, _            := dproxy.New(tempData).A(0).M("y").Float64()
		z, _            := dproxy.New(tempData).A(0).M("z").Float64()
		//fmt.Println(humidity, co2, x, y, z)

		var _temperature1, _ = strconv.ParseFloat(temperature1, 64)
		var _temperature2, _ = strconv.ParseFloat(temperature2, 64)
		var _temperature3, _ = strconv.ParseFloat(temperature3, 64)
		//var _humidity, _     = strconv.ParseFloat(humidity, 64)
		//var _co2, _          = strconv.ParseFloat(co2, 64)
		//var _x, _            = strconv.ParseFloat(x, 64)
		//var _y, _            = strconv.ParseFloat(y, 64)
		//var _z, _            = strconv.ParseFloat(z, 64)
		//fmt.Println(humidity, co2, _x, _y, _z)
		//
		//minCelsius, _ := dproxy.New(tempData).M("forecasts").A(1).M("temperature").M("min").M("celsius").String()
        //var _fmic, _ = strconv.ParseFloat(minCelsius, 64)
        return vicId, serial_no, timestamp, _temperature1, _temperature2, _temperature3, humidity, co2, x, y, z
}


func main() {
		//定義したフラグをParseして使えるようにする
        flag.Parse()

		var c vicCollector
		prometheus.MustRegister(c)
		
        // Expose the registered metrics via HTTP.
        http.Handle("/metrics", promhttp.Handler())
        log.Fatal(http.ListenAndServe(*addr, nil))
}

