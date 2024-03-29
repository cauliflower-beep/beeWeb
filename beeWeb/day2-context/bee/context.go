package bee

import (
	"encoding/json"
	"fmt"
	"net/http"
)

/*
 	设计 context 的必要性：
		1.对Web服务来说，无非是根据请求*http.Request，构造响应http.ResponseWriter
		但是这两个对象提供的接口粒度太细，比如我们要构造一个完整的响应，需要考虑消息头(Header)和消息体(Body)，
		而 Header 包含了状态码(StatusCode)，消息类型(ContentType)等几乎每次请求都需要设置的信息。
		因此，如果不进行有效的封装，那么框架的用户将需要写大量重复，繁杂的代码，而且容易出错。
		针对常用场景，能够高效地构造出 HTTP 响应是一个好的框架必须考虑的点。

		2.针对使用场景，封装*http.Request和http.ResponseWriter的方法，简化相关接口的调用，只是设计 Context 的原因之一。
		对于框架来说，还需要支撑额外的功能。
		例如，将来解析动态路由/hello/:name，参数:name的值放在哪呢？
		再比如，框架需要支持中间件，那中间件产生的信息放在哪呢？
		Context 随着每一个请求的出现而产生，请求的结束而销毁，和当前请求强相关的信息都应由 Context 承载。
		因此，设计 Context 结构，扩展性和复杂性留在了内部，而对外简化了接口。
		路由的处理函数，以及将要实现的中间件，参数都统一使用 Context 实例， Context 就像一次会话的百宝箱，可以找到任何东西。
*/

// H 取别名，构建JSON 数据时，显得更简洁
type H map[string]interface{}

type Context struct {
	// origin objects
	Writer http.ResponseWriter
	Req    *http.Request
	// request info 提供对 Path Method 这两个常用属性的直接访问
	Path   string
	Method string
	// response info
	StatusCode int
}

func newContext(w http.ResponseWriter, req *http.Request) *Context {
	return &Context{
		Writer: w,
		Req:    req,
		Path:   req.URL.Path,
		Method: req.Method,
	}
}

func (c *Context) PostForm(key string) string {
	return c.Req.FormValue(key)
}

func (c *Context) Query(key string) string {
	return c.Req.URL.Query().Get(key)
}

/*-------完整响应需要考虑的基本信息-----------*/

func (c *Context) Status(code int) {
	c.StatusCode = code
	// 设置http回包数据中的响应行
	c.Writer.WriteHeader(code)
}

func (c *Context) SetHeader(key string, value string) {
	// 设置http回包数据中的响应头
	c.Writer.Header().Set(key, value)
}

/*------快速构造响应-------*/
func (c *Context) String(code int, format string, values ...interface{}) {
	// 注意调用顺序，必须是 Header().Set -> WriteHeader -> Write
	c.SetHeader("Content-Type", "text/plain")
	c.Status(code)
	// 设置http回包数据中的响应体
	c.Writer.Write([]byte(fmt.Sprintf(format, values...)))
}

func (c *Context) JSON(code int, obj interface{}) {
	c.SetHeader("Content-Type", "application/json")
	c.Status(code)
	// 创建一个json编码器，将数据编码并写入c.Writer
	encoder := json.NewEncoder(c.Writer)
	if err := encoder.Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), 500)
	}
}

func (c *Context) Date(code int, data []byte) {
	c.Status(code)
	c.Writer.Write(data)
}

func (c *Context) HTML(code int, html string) {
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)
	c.Writer.Write([]byte(html))
}
