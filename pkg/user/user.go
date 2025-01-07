package user

import (
	"encoding/json"
	"errors"

	"github.com/YD-GitGang/serverless-with-go-and-aws---portfolio2---/pkg/validators"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

// 複数箇所に同じ文字列を直書きすると、タイポや文言修正漏れが起こりやすい。変数・定数にしておけば修正があっても1箇所で済む。
var (
	ErrorFailedToFetchRecord     = "error: failed to fetch record"
	ErrorFailedToUnmarshalRecord = "error: failed to unmarshal record"
	ErrorInvalidUserData         = "error: invalid user data"
	ErrorCouldNotDeleteItem      = "error: could not delete item"
	ErrorCouldNotDynamoPutItem   = "error: could not dynamo put item"
	ErrorCouldNotMarshalItem     = "error: could not marshal item"
	ErrorUserDoesNotExists       = "error: user does not exists"
	ErrorUserAlreadyExists       = "error: user already exists"
	ErrorInvalidEmail            = "error: invalid email"
)

type User struct {
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

func FetchUser(email string, tableName string, dynaClient dynamodbiface.DynamoDBAPI) (*User, error) {
	input := &dynamodb.GetItemInput{
		Key:       map[string]*dynamodb.AttributeValue{"email": {S: aws.String(email)}}, // (※3)
		TableName: aws.String(tableName),
	}
	/*
		(※3)
		&dynamodb.AttributeValue{S: aws.String(email)} と書くべきではないのか？について
		Goでは一見 struct を入れているように見えても代入先の型がポインタ型の場合はコンパイラが自動的にアドレスを取ってポインタとして扱う仕様がある。
		Key: map[string]*dynamodb.AttributeValue{"email": &dynamodb.AttributeValue{S: aws.String(email)}}
		↓
		Key: map[string]*dynamodb.AttributeValue{"email":{S: aws.String(email)}}
	*/

	result, err := dynaClient.GetItem(input) // (※2)

	if err != nil {
		return nil, errors.New(ErrorFailedToFetchRecord)
	}

	item := new(User)
	err = dynamodbattribute.UnmarshalMap(result.Item, item) // (※1)
	if err != nil {
		return nil, errors.New(ErrorFailedToUnmarshalRecord)
	}
	/*
		(※1)では(※2)で定義済みの err を上書きしてるが、下記の様にif文に組み込んでif文以降はerrを使用不可にして「:=」でerrを何度も定義でもOK。
			if err := dynamodbattribute.UnmarshalMap(result.Item, item); err != nil {
				return nil, errors.New("Error faild to Unmarshall")
			}
	*/

	return item, nil
}

func FetchUsers(tableName string, dynaClient dynamodbiface.DynamoDBAPI) (*[]User, error) {
	input := &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	}

	result, err := dynaClient.Scan(input)

	if err != nil {
		return nil, errors.New(ErrorFailedToFetchRecord)
	}

	items := new([]User)
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, items)
	if err != nil {
		return nil, errors.New(ErrorFailedToUnmarshalRecord)
	}

	return items, nil
}

func CreateUser(req events.APIGatewayProxyRequest, tableName string, dynaClient dynamodbiface.DynamoDBAPI) (*User, error) {
	var u User
	if err := json.Unmarshal([]byte(req.Body), &u); err != nil { // (※4)
		return nil, errors.New(ErrorInvalidUserData)
	}
	/*
		(※4)
		json.Unmarshal() の第2引数がポインタの理由
		第2引数 u が値（非ポインタ）であった場合、フィールドを書き換えようとしてもそのコピーに対して行うだけで、呼び出し元の変数には反映されないから。
	*/

	if !validators.IsEmailValid(u.Email) {
		return nil, errors.New(ErrorInvalidEmail)
	}

	currentUser, _ := FetchUser(u.Email, tableName, dynaClient) // (※5)
	if currentUser != nil && len(currentUser.Email) != 0 {      // (※6)
		return nil, errors.New(ErrorUserAlreadyExists)
	}

	av, err := dynamodbattribute.MarshalMap(u) // (※7)
	if err != nil {
		return nil, errors.New(ErrorCouldNotMarshalItem)
	}

	input := &dynamodb.PutItemInput{ // (※8)。(※3)ではKeyに入れるものを手動で作ったが、ここでは(※7)の関数で作ったものをItemに入れてる。
		Item:      av,
		TableName: aws.String(tableName),
	}

	_, err = dynaClient.PutItem(input)
	if err != nil {
		return nil, errors.New(ErrorCouldNotDynamoPutItem)
	}

	return &u, nil
}

func UpdateUser(req events.APIGatewayProxyRequest, tableName string, dynaClient dynamodbiface.DynamoDBAPI) (*User, error) {
	var u User
	if err := json.Unmarshal([]byte(req.Body), &u); err != nil {
		return nil, errors.New(ErrorInvalidUserData)
	}

	currentUser, _ := FetchUser(u.Email, tableName, dynaClient)
	if currentUser != nil && len(currentUser.Email) == 0 {
		return nil, errors.New(ErrorUserDoesNotExists)
	}

	av, err := dynamodbattribute.MarshalMap(u)
	if err != nil {
		return nil, errors.New(ErrorCouldNotMarshalItem)
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(tableName),
	}
	_, err = dynaClient.PutItem(input)
	if err != nil {
		return nil, errors.New(ErrorCouldNotDynamoPutItem)
	}

	return &u, nil
}

func DeleteUser(req events.APIGatewayProxyRequest, tableName string, dynaClient dynamodbiface.DynamoDBAPI) error {
	email := req.QueryStringParameters["email"]
	input := &dynamodb.DeleteItemInput{
		Key:       map[string]*dynamodb.AttributeValue{"email": {S: aws.String(email)}},
		TableName: aws.String(tableName),
	}
	_, err := dynaClient.DeleteItem(input)
	if err != nil {
		return errors.New(ErrorCouldNotDeleteItem)
	}

	return nil
}

/*
(※5)
・FetchUser した時、DynamoDB にユーザー情報がなかった場合の返り値
結論「nil」ではなく「&User{Email: "", FirstName: "", LastName: ""}」が返ってくる。
FetchUser 内の(※2)の result, err := dynaClient.GetItem(input) の戻り値は下記のように定義されていて、
dynamodb.GetItemOutput 型の構造体のポインタと error が返ってくる。
type GetItemOutput struct {
	Item map[string]*AttributeValue
	// ... 省略 ...
}
つまり(※2)の result は *dynamodb.GetItemOutput。(※7)と(※8)の逆なだけ。DynamoDB にデータを送るときは(※7)で
マーシャルして *dynamodb.AttributeValue(dynamodb にある AttributeValue のポインタ、つまり *AttributeValue)にして PutItemInput の
Item フィールドに入れてる。GetItemInput は Item フィールドじゃなくて Key フィールドを使ってるだけ。△△ItemInput も △△ItemOutputも中身は
多分だいたい同じ。
DynamoDB に指定キーのアイテムが見つからない場合、GetItemOutput は返るが、中身の Item フィールドが nil となる。
GetItem ドキュメント:「If there is no matching item, Item is an empty map (or nil).」
つまり、(※2)の result には Item フィールドが nil の *dynamodb.GetItemOutput が返ってきて、
それを(※1)の err = dynamodbattribute.UnmarshalMap(result.Item, item) でアンマーシャルして、
中身が {Email: "", FirstName: "", LastName: ""} の User 型のポインタとなり、それが FetchUser の返り値 item となる。
よって(※5)の currentUser は nil ではなく空の User 構造体のポインタ &User{Email: "", FirstName: "", LastName: ""} になる。
currentUser が nil になるのは、ユーザー情報が登録されていない時じゃなくて、エラーの時。FetchUser 関数は、
「DynamoDB操作のエラー」と「ユーザが見つからない」を区別して、エラー時は(nil, error)を返し、ユーザなし時は(&User{}, nil)を返す設計。
因みに、(※5)では簡単のために(nil, error)という返り値を想定していない。だから、返り値のエラーを省略しエラーハンドリングをハショッているし、
currentUser が nil の時の条件分岐もハショッている。

(※6)
currentUser.Email を実行した際に currentUser が nil だった場合パニックが起こるから currentUser != nil が必要。
*/

/*
【全体像】
＜ユーザー → Lambda＞
byte型の json の HTTPリクエストを、Goで用意した User構造体にアンマーシャル。※A
( user.go の json.Unmarshal() )

＜Lambda → DynamoDB＞
※Aを DynamoDB用の map[string]＊dynamodb.AttributeValue にマーシャルして、△△ItemInput 構造体の Itemフィールドに入れる。
( user.go の dynamodbattribute.MarshalMap() )

＜DynamoDB → Lambda＞
DynamoDB形式の △△ItemOutput の Itemフィールドから取り出した map[string]＊dynamodb.AttributeValue をGoで用意した User構造体にアンマーシャル。※B
( user.go の dynamodbattribute.UnmarshalMap() )

＜Lambda → ユーザー＞
※Bを json にマーシャルして、HTTPレスポンスとして APIGatewayProxyResponse構造体の bodyフィールドに入れる。
( api_response.go の json.Marshal() )
*/
