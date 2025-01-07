package handlers

import (
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
)

func apiResponse(status int, body interface{}) (*events.APIGatewayProxyResponse, error) { // (※2)
	resp := events.APIGatewayProxyResponse{Headers: map[string]string{"Content-Type": "application/json"}}
	resp.StatusCode = status

	stringBody, _ := json.Marshal(body) // (※1)
	resp.Body = string(stringBody)      // Marshalするとbyte型のjsonが返ってくるからstringにキャスト

	return &resp, nil
}

/*
(※1)
・nil はマーシャルできるのか
handlers.go の DeleteUser を実行すると、json.Marshal() の第2引数に nil が来ることになる。一見マーシャルできなそうに見えるが、
json.Marshal()は、引数に nil が来た場合JSON の "null" に変換する仕組みをif文で実装しているので問題ない。
よって、DeleteUser を実行すると jsonに "null" が記述され、ターミナルにnullという文字列が表示される。

(※2)
・関数の全体像
Headers, StatusCode, Body と各フィールドに値を入れていってるシンプルなつくり。
*/
