package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestConfig_AddDefaultData(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	testApp.Session.Put(ctx, "flash", "flash")
	testApp.Session.Put(ctx, "warning", "warning")
	testApp.Session.Put(ctx, "error", "error")

	td := testApp.AddDefaultData(
		&TemplateData{},
		req,
	)

	if td.Flash != "flash" {
		t.Error("failed to get flash data!")
	}
	if td.Error != "error" {
		t.Error("failed to get error data!")
	}
	if td.Warning != "warning" {
		t.Error("failed to get warning data!")
	}
}

func TestConfig_IsAuthenticated(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	auth := testApp.IsAuthenticated(req)
	if auth {
		t.Error("should be not authenticated!")
	}

	testApp.Session.Put(ctx, "userId", 1)
	auth = testApp.IsAuthenticated(req)
	if !auth {
		t.Error("should be authenticated!")
	}

}

func TestConfig_render(t *testing.T) {
	path = "./templates"
	rr := httptest.NewRecorder()

	req, _ := http.NewRequest("GET", "/", nil)
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	testApp.render(
		rr,
		req,
		"home.page.gohtml",
		&TemplateData{},
	)

	if rr.Code != 200 {
		t.Error("failed to render home page")
	}
}
