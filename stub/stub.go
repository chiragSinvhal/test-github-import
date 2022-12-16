// Copyright 2022 Hewlett Packard Enterprise Development LP.
package stub

import (
	"net/http"

	"github.com/unrolled/render"
)

func JSON(w http.ResponseWriter, status int, result interface{}) {
	if result == nil {
		w.WriteHeader(status)

		return
	}
	_ = r.JSON(w, status, result)
}

var r = render.New(render.Options{
	IndentJSON: true,
})
