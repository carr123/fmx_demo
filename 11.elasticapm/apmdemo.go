package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/carr123/fmx"
	"go.elastic.co/apm/v2"
)

//this demo shows how to incorporate elastic apm package into fmx.
//thus, every http request will be logged to Elastic Search engine.

func main() {
	os.Setenv("ELASTIC_APM_SERVER_URL", "http://127.0.0.1:8200")
	os.Setenv("ELASTIC_APM_SERVICE_NAME", "appname")
	os.Setenv("ELASTIC_APM_API_KEY", "")
	os.Setenv("ELASTIC_APM_SECRET_TOKEN", "")
	os.Setenv("ELASTIC_APM_VERIFY_SERVER_CERT", "false")

	router := fmx.NewServeMux()
	router.GET("/api/v1", HTTPLoggerApm(true, true), func(c *fmx.Context) {
		tx := c.MustGet("apmtx").(*apm.Transaction)
		//tracer := c.MustGet("apmtracer").(*apm.Tracer)

		span1 := tx.StartSpan("readcache", "readcache", nil)
		time.Sleep(time.Millisecond * 100)
		span1.End()

		span2 := tx.StartSpan("readdb", "readcockroach", nil)
		time.Sleep(time.Millisecond * 150)
		span2.End()

		//c.AddError(fmt.Errorf("read cache fail"))

		c.JSON(200, fmx.H{"result": "ok"})
	})

	err := http.ListenAndServe("127.0.0.1:8089", router)
	if err != nil {
		fmt.Println(err)
	}
}

func setContext(ctx *apm.Context, c *fmx.Context, reqBody []byte, respBody []byte) {
	ctx.SetHTTPRequest(c.Request)
	ctx.SetHTTPStatusCode(c.Writer.GetStatusCode())
	ctx.SetHTTPResponseHeaders(c.Writer.Header())

	if reqBody != nil {
		ctx.SetCustom("reqbody", string(reqBody))
	}

	if respBody != nil {
		ctx.SetCustom("respbody", string(respBody))
	}
}

func HTTPLoggerApm(bShowReqBody bool, bShowRespBody bool) fmx.HandlerFunc {
	return func(c *fmx.Context) {
		tracer := apm.DefaultTracer()
		requestName := fmt.Sprintf("%s %s", c.Request.Method, c.Request.URL.String())

		tx := tracer.StartTransaction(requestName, "request")
		defer tx.End()

		var reqBody []byte
		var respBody []byte

		if bShowReqBody && c.Request.Body != nil {
			buff := &bytes.Buffer{}
			io.Copy(buff, c.Request.Body)
			c.Request.Body.Close()
			c.Request.Body = ioutil.NopCloser(buff)
			reqBody = buff.Bytes()
		}

		if bShowRespBody {
			c.Writer.SetRecordRespBody(bShowRespBody)
		}

		c.Set("apmtx", tx)
		c.Set("apmtracer", tracer)

		defer func() {
			if v := recover(); v != nil {
				if c.Writer.GetStatusCode() == 0 {
					c.AbortWithStatus(500)
				} else {
					c.Abort()
				}
				e := tracer.Recovered(v)
				e.SetTransaction(tx)
				setContext(&e.Context, c, reqBody, respBody)
				e.Send()
				return
			}

			if bShowRespBody {
				if tmpData, err := ioutil.ReadAll(c.Writer.GetRespBody()); err == nil {
					respBody = tmpData
				}
			}

			if tx.Sampled() {
				setContext(&tx.Context, c, reqBody, respBody)
			}

			for _, err := range c.GetErrors() {
				var errstr string
				if e, ok := err.(fmx.ErrorWithPos); ok {
					errstr = e.String()
				} else {
					errstr = err.Error()
				}

				e := tracer.NewError(fmt.Errorf("%s", errstr))
				e.SetTransaction(tx)
				setContext(&e.Context, c, reqBody, respBody)
				e.Send()
			}
		}()

		c.Next()
	}
}
