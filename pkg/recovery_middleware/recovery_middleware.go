package recovery_middleware

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
)

type RecoverMiddleware func(h httprouter.Handle) httprouter.Handle

func Recover(logger *zap.SugaredLogger, h httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		defer func() {
			rec := recover()
			if rec != nil {
				logger.Infof("recovered from panic. %v", rec)
			}
		}()
		h(w, r, params)
	}
}
