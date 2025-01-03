package main

import (
	"os"

	"github.com/YD-GitGang/serverless-with-go-and-aws---portfolio2---/pkg/handlers"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

var dynaClient dynamodbiface.DynamoDBAPI // (※1)

/*
(※1)
「DynamoDBクライアント」とは。
「DynamoDBクライアント」とは、GoプログラムがDynamoDBという「サーバー」にリクエストを送るための「通信の窓口オブジェクト」のこと。
多くの文脈では、「クライアント」＝「サーバーからデータを受け取るもの（例：ブラウザやPC、スマホアプリ）」を指す。
プログラムがDynamoDBに対してAPI呼び出しを行うとき、その処理を担当する「窓口」オブジェクトのことをクライアントと呼ぶ。
dynamodb.New(...) や dynamodb.NewFromConfig(...) といった関数を呼び出すと、DynamoDBへの接続に必要な機能をまとめたオブジェクトが返ってくる。
DynamoDBを利用したいプログラム側にとって、DynamoDBはサーバー（データを保持し、APIで提供）。
そのDynamoDBサーバーにリクエストを送る役割を担うのが、Goプログラム内にある「クライアントオブジェクト」。
ネットワーク通信で言う「クライアントとサーバー」の関係が、プログラム(クライアント) → DynamoDB(サーバー) の形で成り立っている。
*/

func main() {
	region := os.Getenv("AWS_REGION")
	awsSession, err := session.NewSession(&aws.Config{Region: aws.String(region)}) // (※2)

	if err != nil {
		return
	}

	dynaClient = dynamodb.New(awsSession)
	lambda.Start(handler)
}

const tableName = "go-serverless"

func handler(req events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) { // (※3)
	switch req.HTTPMethod {
	case "GET":
		return handlers.GetUser(req, tableName, dynaClient)
	case "POST":
		return handlers.CreateUser(req, tableName, dynaClient)
	case "PUT":
		return handlers.UpdateUser(req, tableName, dynaClient)
	case "DELETE":
		return handlers.DeleteUser(req, tableName, dynaClient)
	default:
		return handlers.UnhandledMethod()
	}
}

/*
(※2)
・なぜライブラリの作者は「aws.Config」とせず「&aws.Config」のようにポインタ型を受け取る設計にしたのか？
＜ 理由 1 ＞ nil かどうかの判定ができる。-------
ポインタであれば nil 判定が簡単にでき、「引数を省略または nil にしたときはデフォルト値を使う」といったロジックを組み込みやすくなる。

例
func NewClient(cfg *Config) *Client {
    // nilチェックができる
    if cfg == nil {
        // 呼び出し元が nil を渡した時はデフォルト設定で初期化
        cfg = &Config{
            Region:  "us-east-1",
            Timeout: 30,
        }
    }
}

上記の if cfg == nil {...} で、呼び出し元が何も設定せずに nil を渡した場合に、一括でデフォルト設定を作れる。
値渡し（func NewClient(cfg Config) *Client) にしてしまうと、「cfgがnilかどうか」を判定する術がなくなる。構造体は必ず何らかの値を持ち、
nil という概念がない。「設定を省略したい = nilを渡す」という選択肢がなくなるので、「すべてのフィールドがゼロ値かどうか」を見て判断するなど、
もう少し複雑なロジックを組む必要がある。

＜ 理由 2 ＞ 「未設定のゼロ」なのか「ゼロという値」なのかという混乱を避けるため-------
例
// 値渡しバージョン
func NewClient(cfg Config) *Client {
    // cfg そのものはnilにはなり得ない。なので省略したい場合も cfg = Config{}
    // 何が「省略された」のかを判定したいなら、フィールドごとに「ゼロ値かどうか」を判定するしかない
    if cfg.Region == "" {
        cfg.Region = "us-east-1"
    }
    if cfg.Timeout == 0 {
        cfg.Timeout = 30
    }
}
cfg は nil にはなり得ない。なので省略したい場合も cfg = Config{} とすることになる。
この構造体が「意図的にregionを空文字にした」のか「設定忘れ」なのか見分けがつかない。「ゼロという値なのか」「未設定なのか」が曖昧になる。
ポインタだと「nilなら未設定」「0はちゃんとした値」などの区別ができる。

・なぜライブラリの作者は「region」とせず「aws.String(region)」のようにポインタ型を受け取る設計にしたのか？
「省略」と「空文字」を区別したい。
Region: nil なら「リージョン未指定（デフォルトに従う）」、Region: aws.String("") なら「空文字をセット」といった形で区別できる。
値渡しの場合は「設定忘れの空文字」と「意図した空文字」が同じ扱いになってしまう。もともと AWS の多くのサービスは JSON ベースでやりとりを
する際に、「フィールドが存在しない(none)」「フィールドが存在するが null」「フィールドが存在して文字列が空」などを区別するニーズがある。
そのため、Go用のSDKでも ポインタを使うことで「未設定」と「空文字」の違いを反映しやすい設計になっている。

・aws.Configとは
aws.Config は、AWSサービスへの接続設定をまとめた設定用の構造体(リージョン、認証情報、などなど)。
Go SDKの多くの関数呼び出しで、このaws.Configの情報を参照してAWSへアクセスするらしい。構造体として各種設定を詰め込み、
session.NewSession(...) に渡すと、AWS SDKのサービスを使うのに必要な接続情報的なのが生成されるのだろう、きっと。
そしてそれをdynamodb.New(...)に渡せばDynamoDBクライアントが生成される。
*/

/*
(※3)
・なぜライブラリの作者はAPI Gatewayにポインタ型を返す設計にしたのか？(*events.APIGatewayProxyResponse)
＜ 理由 1 ＞ エラー時に nil を返せるので、「ゼロ値の構造体」と「レスポンスなし」を区別できる-------
エラー時や特定の状況で「レスポンスを返せない・不要」というケースが想定されるなら、(*events.APIGatewayProxyResponse, error) の形だとレスポンス用
の構造体を nil で返すことができる。値渡しだと「ゼロ値の構造体」を返す以外に手がなく、「ゼロ値の構造体」と「レスポンスなし」を区別するのが難しくなる。
たとえば return nil, errors.New("some error") と書けるのは、ポインタならではのメリット。nil を返せるおかげで「レスポンスなし」を示せる。

＜ 理由 2 ＞ 大きな構造体をコピーしないで済む-------
大きな構造体を「値」として返すと、呼び出しのたびに構造体全体をメモリコピーするコストがかかる。
ポインタならばアドレスだけを返すため、コピー負荷が少なくなる。

・events.APIGatewayProxyRequestについて
「events.APIGatewayProxyRequest」という構造体でユーザーからAPI Gateway経由でリクエストを受け取り、「*events.APIGatewayProxyResponse」という
構造体のメモリーアドレスをユーザーに返す。API Gatewayでマッピングテンプレートを作成して翻訳したりせず、ユーザーからのリクエストデータを直接ここ
に放り込んでここでさばく。
*/
