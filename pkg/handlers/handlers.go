package handlers

import (
	"net/http"

	"github.com/YD-GitGang/serverless-with-go-and-aws---portfolio2---/pkg/user"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

type ErrorBody struct {
	ErrorMsg *string `json:"error,omitempty"` // (※1)
}

var ErrorMethodNotAllowed = "method not allowed"

func GetUser(req events.APIGatewayProxyRequest, tableName string, dynaClient dynamodbiface.DynamoDBAPI) (
	*events.APIGatewayProxyResponse, error,
) {
	email := req.QueryStringParameters["email"]
	if len(email) > 0 {
		result, err := user.FetchUser(email, tableName, dynaClient)
		if err != nil {
			return apiResponse(http.StatusBadRequest, ErrorBody{aws.String(err.Error())})
		}
		return apiResponse(http.StatusOK, result)
	}

	result, err := user.FetchUsers(tableName, dynaClient)
	if err != nil {
		return apiResponse(http.StatusBadRequest, ErrorBody{aws.String(err.Error())})
	}
	return apiResponse(http.StatusOK, result)
}

func CreateUser(req events.APIGatewayProxyRequest, tableName string, dynaClient dynamodbiface.DynamoDBAPI) (
	*events.APIGatewayProxyResponse, error,
) {
	result, err := user.CreateUser(req, tableName, dynaClient)
	if err != nil {
		return apiResponse(http.StatusBadRequest, ErrorBody{aws.String(err.Error())})
	}
	return apiResponse(http.StatusCreated, result)
}

func UpdateUser(req events.APIGatewayProxyRequest, tableName string, dynaClient dynamodbiface.DynamoDBAPI) (
	*events.APIGatewayProxyResponse, error,
) {
	result, err := user.UpdateUser(req, tableName, dynaClient)
	if err != nil {
		return apiResponse(http.StatusBadRequest, ErrorBody{aws.String(err.Error())})
	}
	return apiResponse(http.StatusOK, result)
}

func DeleteUser(req events.APIGatewayProxyRequest, tableName string, dynaClient dynamodbiface.DynamoDBAPI) (
	*events.APIGatewayProxyResponse, error,
) {
	err := user.DeleteUser(req, tableName, dynaClient)
	if err != nil {
		return apiResponse(http.StatusBadRequest, ErrorBody{aws.String(err.Error())})
	}
	return apiResponse(http.StatusOK, nil)
}

func UnhandledMethod() (*events.APIGatewayProxyResponse, error) {
	return apiResponse(http.StatusMethodNotAllowed, ErrorMethodNotAllowed)
}

/*
(※1)
・なぜ「ErrorMsg string」とせず「ErrorMsg *string」のようにポインタ型を受け取る設計にしたのか？
＜ 理由 1 ＞ nilで未設定のフィールドを表現したい
string 型は空文字列を含む何らかの値を持つため、「未設定（nil）」という状態を表せない。*string であれば、
「指し示す住所が決まっていない空のポインタ(nil)」で未設定を表現できる。

＜ 理由 2 ＞ 「未設定」のフィールドと「空文字」のフィールドを区別するため。
Goで定義した構造体をjsonにマーシャルする際、フィールドの型をstringのポインタにすることで、「空文字」なら空文字をjsonに出力し「未設定」ならomitする
という区別ができる。もしフィールドの型をstringにしてしまったら、「空文字」も「未設定」もどちらもomitされてしまい区別ができない。

------------------------------
例
パターン1：*stringのフィールドに「空文字」を渡す
パターン2：*stringのフィールドを「未設定」にする

type Person struct{
    Name      string   `json:"name,omitempty"`
	Location  *string  `json:"location,omitempty"`
}

// パターン1
func main() {
	b := []byte(`{"name":"", "location":""}`)　　// 空文字を渡す
    var p Person
	if err := json.Unmarshal(b, &p); err != nil{
	    fmt.Println(err)
	}
    fmt.Println(p.Name, p.Location)

	v, _ := json.Marshal(p)
	fmt.Println(string(v))
}

// 出力
(分かりずらいが先頭に空文字がある)
↓
 0xc000014130
{"location":""}

// パターン2
func main() {
	b := []byte(`{"name":""}`)      // 未設定
    var p Person
	if err := json.Unmarshal(b, &p); err != nil{
	    fmt.Println(err)
	}
    fmt.Println(p.Name, p.Location)

	v, _ := json.Marshal(p)
	fmt.Println(string(v))
}

// 出力
(分かりずらいが先頭に空文字がある)
↓
 <nil>
{}
------------------------------

上記の例から分かるように、
パターン1のアンマーシャル後の結果は「空文字」と「アドレス」が出力され、
マーシャル後の結果は「string型の方がomit」され「*string型の方は空文字」が出力された。
パターン2のアンマーシャル後の結果は「空文字」と「nil」が出力され、
マーシャル後の結果は「string型の方はomit」され「*string型の方もomit」された。
*/
