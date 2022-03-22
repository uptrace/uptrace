package main

import (
	"context"
	"errors"
	"html/template"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace-go/uptrace"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const (
	indexTmpl   = "index"
	profileTmpl = "profile"
)

func main() {
	ctx := context.Background()

	uptrace.ConfigureOpentelemetry()
	defer uptrace.Shutdown(ctx)

	db, err := gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	if err := db.Use(otelgorm.NewPlugin()); err != nil {
		panic(err)
	}

	_ = db.AutoMigrate(&User{})
	db.Create(&User{Username: "world"})
	db.Create(&User{Username: "foo-bar"})

	handler := &Handler{
		db:  db,
		log: otelzap.New(zap.NewExample()),
	}

	router := gin.Default()
	router.SetHTMLTemplate(parseTemplates())
	router.Use(otelgin.Middleware("service-name"))
	router.GET("/", handler.Index)
	router.GET("/hello/:username", handler.Hello)

	if err := router.Run("localhost:9999"); err != nil {
		log.Print(err)
	}
}

type Handler struct {
	db  *gorm.DB
	log *otelzap.Logger
}

func (h *Handler) Index(c *gin.Context) {
	ctx := c.Request.Context()

	// Extract span from the request context.
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.SetAttributes(
			attribute.String("string_key", "string_value"),
			attribute.Int("int_key", 42),
			attribute.StringSlice("string_slice_key", []string{"foo", "bar"}),
		)
	}

	h.log.Ctx(ctx).Error("hello from zap",
		zap.Error(errors.New("hello world")),
		zap.String("foo", "bar"))

	otelgin.HTML(c, http.StatusOK, indexTmpl, gin.H{
		"traceURL": uptrace.TraceURL(trace.SpanFromContext(ctx)),
	})
}

func (h *Handler) Hello(c *gin.Context) {
	ctx := c.Request.Context()

	username := c.Param("username")
	user := new(User)
	if err := h.db.WithContext(ctx).Where("username = ?", username).First(user).Error; err != nil {
		_ = c.Error(err)
		return
	}

	otelgin.HTML(c, http.StatusOK, profileTmpl, gin.H{
		"username": user.Username,
		"traceURL": uptrace.TraceURL(trace.SpanFromContext(ctx)),
	})
}

type User struct {
	ID       int64 `gorm:"primaryKey"`
	Username string
}

func parseTemplates() *template.Template {
	indexTemplate := `
		<html>
		<p>Here are some routes for you:</p>
		<ul>
			<li><a href="/hello/world">Hello world</a></li>
			<li><a href="/hello/foo-bar">Hello foo-bar</a></li>
		</ul>
		<p>View trace: <a href="{{ .traceURL }}" target="_blank">{{ .traceURL }}</a></p>
		</html>
	`
	t := template.Must(template.New(indexTmpl).Parse(indexTemplate))

	profileTemplate := `
		<html>
		<h3>Hello {{ .username }}</h3>
		<p>View trace: <a href="{{ .traceURL }}" target="_blank">{{ .traceURL }}</a></p>
		</html>
	`
	return template.Must(t.New(profileTmpl).Parse(profileTemplate))
}
