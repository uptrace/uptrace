package handler

import (
	"net/http"

	"github.com/uptrace/uptrace/example/go-zero/api/internal/logic"
	"github.com/uptrace/uptrace/example/go-zero/api/internal/svc"
	"github.com/uptrace/uptrace/example/go-zero/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func ApiHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.Request
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewApiLogic(r.Context(), svcCtx)
		resp, err := l.Api(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
