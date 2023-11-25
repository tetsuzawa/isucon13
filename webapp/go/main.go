package main

// ISUCON的な参考: https://github.com/isucon/isucon12-qualify/blob/main/webapp/go/isuports.go#L336
// sqlx的な参考: https://jmoiron.github.io/sqlx/

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	echolog "github.com/labstack/gommon/log"
)

const (
	listenPort                     = 8080
	powerDNSSubdomainAddressEnvKey = "ISUCON13_POWERDNS_SUBDOMAIN_ADDRESS"
)

var (
	powerDNSSubdomainAddress string
	dbConn                   *sqlx.DB
	secret                   = []byte("isucon13_session_cookiestore_defaultsecret")
	AppVersion               string
)

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	if secretKey, ok := os.LookupEnv("ISUCON13_SESSION_SECRETKEY"); ok {
		secret = []byte(secretKey)
	}
}

type InitializeResponse struct {
	Language string `json:"language"`
}

func initializeHandler(c echo.Context) error {
	if out, err := exec.Command("../sql/pg/init.sh").CombinedOutput(); err != nil {
		c.Logger().Warnf("init.sh failed with err=%s", string(out))
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to initialize: "+err.Error())
	}

	c.Request().Header.Add("Content-Type", "application/json;charset=utf-8")

	ctx := c.Request().Context()
	if err := rdb.FlushAll(ctx).Err(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to flush redis: "+err.Error())
	}
	type TotalReaction struct {
		TotalReactionCount int64 `db:"total_reaction_count"`
		UserID             int64 `db:"user_id"`
	}
	var trs []TotalReaction
	if err := dbConn.SelectContext(ctx, &trs, "SELECT l.user_id, COUNT(*) AS total_reaction_count FROM livestreams l INNER JOIN reactions r ON l.id = r.livestream_id GROUP BY l.user_id"); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to select reactions: "+err.Error())
	}
	for _, tr := range trs {
		if err := rdb.IncrBy(ctx, fmt.Sprintf("total_reaction:%d", tr.UserID), tr.TotalReactionCount).Err(); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to increment total_reaction: "+err.Error())
		}
		if err := rdb.ZIncrBy(ctx, "ranking", 1, strconv.FormatInt(tr.UserID, 10)).Err(); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to increment ranking: "+err.Error())
		}
	}
	type TotalViewer struct {
		TotalViewerCount int64 `db:"total_viewer_count"`
		UserID           int64 `db:"user_id"`
	}
	var tvs []TotalViewer
	if err := dbConn.SelectContext(ctx, &tvs, "SELECT l.user_id, COUNT(*) AS total_viewer_count FROM livestream_viewers_history lv INNER JOIN livestreams l ON l.id = lv.livestream_id GROUP BY l.user_id"); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to select viewers: "+err.Error())
	}
	for _, tv := range tvs {
		if err := rdb.IncrBy(ctx, fmt.Sprintf("livestream_viewers:%d", tv.UserID), tv.TotalViewerCount).Err(); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to increment livestream_viewers: "+err.Error())
		}
	}
	var users []User
	if err := dbConn.SelectContext(ctx, &users, "SELECT id, name FROM users"); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to select users: "+err.Error())
	}
	for _, user := range users {
		type EmojiCount struct {
			EmojiName string `db:"emoji_name"`
			Count     int64  `db:"cnt"`
		}
		q := `
SELECT r.emoji_name, COUNT(*) AS cnt
FROM users u
INNER JOIN livestreams l ON l.user_id = u.id
INNER JOIN reactions r ON r.livestream_id = l.id
WHERE u.name = ?
GROUP BY emoji_name`
		var ecs []EmojiCount
		if err := dbConn.SelectContext(ctx, &ecs, q, user.Name); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to select emoji_name: "+err.Error())
		}
		for _, ec := range ecs {
			if err := rdb.ZIncrBy(ctx, fmt.Sprintf("favorite_emoji:%d", user.ID), float64(ec.Count), ec.EmojiName).Err(); err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "failed to increment reaction: "+err.Error())
			}
		}
	}

	type TotalComment struct {
		TotalCommentCount int64 `db:"total_comment_count"`
		TotalTip          int64 `db:"total_tip"`
		UserID            int64 `db:"user_id"`
	}
	var tcs []TotalComment
	if err := dbConn.SelectContext(ctx, &tcs, "SELECT l.user_id, COUNT(*) AS total_comment_count, SUM(tip) AS total_tip FROM livecomments c INNER JOIN livestreams l ON c.livestream_id = l.id GROUP BY l.user_id"); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to select comments: "+err.Error())
	}
	for _, tc := range tcs {
		if err := rdb.IncrBy(ctx, fmt.Sprintf("total_comment:%d", tc.UserID), tc.TotalCommentCount).Err(); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to increment total_comment: "+err.Error())
		}
		if err := rdb.ZIncrBy(ctx, "ranking", float64(tc.TotalTip), strconv.FormatInt(tc.UserID, 10)).Err(); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to increment ranking: "+err.Error())
		}
	}

	return c.JSON(http.StatusOK, InitializeResponse{
		Language: "golang",
	})
}

func main() {
	initProfile()
	tp, _ := initTracer(context.Background())
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			panic(err)
		}
	}()

	e := echo.New()
	e.Debug = true
	e.Logger.SetLevel(echolog.DEBUG)
	e.Use(middleware.Logger())
	cookieStore := sessions.NewCookieStore(secret)
	cookieStore.Options.Domain = "*.u.isucon.dev"
	e.Use(session.Middleware(cookieStore))
	// e.Use(middleware.Recover())
	e.Use(otelecho.Middleware("isupipe"))

	// 初期化
	e.POST("/api/initialize", initializeHandler)

	// top
	e.GET("/api/tag", getTagHandler)
	e.GET("/api/user/:username/theme", getStreamerThemeHandler)

	// livestream
	// reserve livestream
	e.POST("/api/livestream/reservation", reserveLivestreamHandler)
	// list livestream
	e.GET("/api/livestream/search", searchLivestreamsHandler)
	e.GET("/api/livestream", getMyLivestreamsHandler)
	e.GET("/api/user/:username/livestream", getUserLivestreamsHandler)
	// get livestream
	e.GET("/api/livestream/:livestream_id", getLivestreamHandler)
	// get polling livecomment timeline
	e.GET("/api/livestream/:livestream_id/livecomment", getLivecommentsHandler)
	// ライブコメント投稿
	e.POST("/api/livestream/:livestream_id/livecomment", postLivecommentHandler)
	e.POST("/api/livestream/:livestream_id/reaction", postReactionHandler)
	e.GET("/api/livestream/:livestream_id/reaction", getReactionsHandler)

	// (配信者向け)ライブコメントの報告一覧取得API
	e.GET("/api/livestream/:livestream_id/report", getLivecommentReportsHandler)
	e.GET("/api/livestream/:livestream_id/ngwords", getNgwords)
	// ライブコメント報告
	e.POST("/api/livestream/:livestream_id/livecomment/:livecomment_id/report", reportLivecommentHandler)
	// 配信者によるモデレーション (NGワード登録)
	e.POST("/api/livestream/:livestream_id/moderate", moderateHandler)

	// livestream_viewersにINSERTするため必要
	// ユーザ視聴開始 (viewer)
	e.POST("/api/livestream/:livestream_id/enter", enterLivestreamHandler)
	// ユーザ視聴終了 (viewer)
	e.DELETE("/api/livestream/:livestream_id/exit", exitLivestreamHandler)

	// user
	e.POST("/api/register", registerHandler)
	e.POST("/api/login", loginHandler)
	e.GET("/api/user/me", getMeHandler)
	// フロントエンドで、配信予約のコラボレーターを指定する際に必要
	e.GET("/api/user/:username", getUserHandler)
	e.GET("/api/user/:username/statistics", getUserStatisticsHandler)
	e.GET("/api/user/:username/icon", getIconHandler)
	e.POST("/api/icon", postIconHandler)

	// stats
	// ライブ配信統計情報
	e.GET("/api/livestream/:livestream_id/statistics", getLivestreamStatisticsHandler)

	// 課金情報
	e.GET("/api/payment", GetPaymentResult)

	e.HTTPErrorHandler = errorResponseHandler

	// DB接続
	conn, err := GetDB()
	if err != nil {
		e.Logger.Errorf("failed to connect db: %v", err)
		os.Exit(1)
	}
	defer conn.Close()
	dbConn = conn

	subdomainAddr, ok := os.LookupEnv(powerDNSSubdomainAddressEnvKey)
	if !ok {
		e.Logger.Errorf("environ %s must be provided", powerDNSSubdomainAddressEnvKey)
		os.Exit(1)
	}
	powerDNSSubdomainAddress = subdomainAddr

	// HTTPサーバ起動
	listenAddr := net.JoinHostPort("", strconv.Itoa(listenPort))
	if err := e.Start(listenAddr); err != nil {
		e.Logger.Errorf("failed to start HTTP server: %v", err)
		os.Exit(1)
	}
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func errorResponseHandler(err error, c echo.Context) {
	c.Logger().Errorf("error at %s: %+v", c.Path(), err)
	if he, ok := err.(*echo.HTTPError); ok {
		if e := c.JSON(he.Code, &ErrorResponse{Error: err.Error()}); e != nil {
			c.Logger().Errorf("%+v", e)
		}
		return
	}

	if e := c.JSON(http.StatusInternalServerError, &ErrorResponse{Error: err.Error()}); e != nil {
		c.Logger().Errorf("%+v", e)
	}
}
