package hime_test

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/moonrhythm/hime"
)

// newExampleContext builds a Context wired to an httptest recorder so the
// examples below can show output without starting a real server.
func newExampleContext(method, target string, body string) (*hime.Context, *httptest.ResponseRecorder) {
	var r *http.Request
	if body == "" {
		r = httptest.NewRequest(method, target, nil)
	} else {
		r = httptest.NewRequest(method, target, strings.NewReader(body))
	}
	w := httptest.NewRecorder()
	return hime.NewAppContext(hime.New(), w, r), w
}

func Example() {
	app := hime.New()
	app.Handler(hime.Handler(func(ctx *hime.Context) error {
		return ctx.String("Hello, hime")
	}))

	// drive the app with an httptest recorder instead of ListenAndServe
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	app.ServeHTTP(w, r)

	fmt.Print(w.Body.String())
	// Output: Hello, hime
}

func ExampleApp_Route() {
	app := hime.New()
	app.Routes(hime.Routes{"user": "/users"})

	fmt.Println(app.Route("user"))
	fmt.Println(app.Route("user", 42))
	fmt.Println(app.Route("user", 42, &hime.Param{Name: "tab", Value: "posts"}))
	// Output:
	// /users
	// /users/42
	// /users/42?tab=posts
}

func ExampleApp_Globals() {
	app := hime.New()
	app.Globals(hime.Globals{"AppName": "hime"})

	fmt.Println(app.Global("AppName"))
	// Output: hime
}

func ExampleContext_JSON() {
	ctx, w := newExampleContext(http.MethodGet, "/", "")

	ctx.JSON(struct {
		Message string `json:"message"`
	}{Message: "hello"})

	fmt.Print(w.Body.String())
	// Output: {"message":"hello"}
}

func ExampleContext_XML() {
	ctx, w := newExampleContext(http.MethodGet, "/", "")

	ctx.XML(struct {
		XMLName xml.Name `xml:"greeting"`
		Message string   `xml:"message"`
	}{Message: "hello"})

	fmt.Print(w.Body.String())
	// Output: <greeting><message>hello</message></greeting>
}

func ExampleContext_BindJSON() {
	ctx, _ := newExampleContext(http.MethodPost, "/", `{"name":"hime"}`)

	var body struct {
		Name string `json:"name"`
	}
	ctx.BindJSON(&body)

	fmt.Println(body.Name)
	// Output: hime
}

func ExampleContext_View() {
	app := hime.New()
	app.Template().Parse("index", "Hello, {{.}}")

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := hime.NewAppContext(app, w, r)
	ctx.View("index", "hime")

	fmt.Print(w.Body.String())
	// Output: Hello, hime
}

func ExampleContext_QueryValueInt() {
	ctx, _ := newExampleContext(http.MethodGet, "/search?page=2", "")

	fmt.Println(ctx.QueryValueInt("page"))
	// Output: 2
}

func ExampleContext_QueryValues() {
	ctx, _ := newExampleContext(http.MethodGet, "/?tag=go&tag=web", "")

	fmt.Println(ctx.QueryValues("tag"))
	// Output: [go web]
}

func ExampleContext_StatusCode() {
	ctx, _ := newExampleContext(http.MethodGet, "/", "")

	fmt.Println(ctx.StatusCode()) // defaults to 200 when unset
	ctx.Status(http.StatusTeapot)
	fmt.Println(ctx.StatusCode())
	// Output:
	// 200
	// 418
}

func ExampleContext_Redirect() {
	ctx, w := newExampleContext(http.MethodGet, "/", "")

	ctx.Redirect("/login")

	fmt.Println(w.Code)
	fmt.Println(w.Header().Get("Location"))
	// Output:
	// 302
	// /login
}

func ExampleSafeRedirectPath() {
	// SafeRedirectPath strips the host, defeating open-redirect attempts.
	fmt.Println(hime.SafeRedirectPath("https://evil.example.com/account"))
	// Output: /account
}

func ExampleHMACCookieSigner() {
	signer := hime.NewHMACCookieSigner([]byte("a-32-byte-or-longer-secret-key!!"))

	signed, _ := signer.Sign("session", "user42")
	value, err := signer.Verify("session", signed)
	fmt.Println(value, err == nil)

	// the cookie name is bound into the signature, so reusing the value
	// under a different name fails verification
	_, err = signer.Verify("other", signed)
	fmt.Println(err)
	// Output:
	// user42 true
	// hime: invalid signed cookie
}
