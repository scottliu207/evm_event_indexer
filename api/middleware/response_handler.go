package middleware

import (
	"evm_event_indexer/api/protocol"

	"github.com/gin-gonic/gin"
	"gitlab.com/heshuosg/system/cons/module/er.git"
)

// ResponseHandler :
//  1. 程式開頭需先將 Response 寫入 gin.Context 中
//     c.Set(env.APIResKeyInGinContext, res)
//  2. 程式發生錯誤時, 應將錯誤寫入 gin.error 中, 並直接回傳
//     c.Error(err)
//     return
func ResponseHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Process request
		c.Next()

		// 若已有回應, 則直接略過後續方法
		if c.Writer.Written() && !c.IsAborted() {
			return
		}

		// 取得API回應
		// Controller 執行時, 需先設定 c.Set(env.APIResKeyInGinContext, res)
		res := &protocol.Response{}

		// 設定一個空的物件 ， 用於有error時回傳一個 空的result
		type initResult struct{}
		result := new(initResult)

		if infRes, ok := c.Get(INTERCEPTOR_KEY); ok {
			if realRes, isProtocol := infRes.(*protocol.Response); isProtocol {
				res = realRes
			}
		}

		// 取得執行過程中發生的錯誤, 只處理第一個
		lastError := c.Errors.Last()
		if lastError == nil {
			// 成功執行完畢, 回傳成功訊息
			res.Code = 0
			if res.Result == nil {
				res.Result = result
			}
			c.JSON(c.Writer.Status(), res)
			return
		}

		// 自定義的 CustomerError 類型
		httpCode, code := er.GetAllCode(lastError.Err)
		res.Code = code
		res.Message = lastError.Err.Error()

		// if !SkipOverwriteRes(code, c.FullPath()) {
		// 	// 覆蓋空的struct到result
		// 	res.Result = result
		// }

		c.JSON(httpCode, res)
	}
}
