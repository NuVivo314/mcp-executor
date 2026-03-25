package sandbox

import (
	"context"
	"encoding/json"

	"modernc.org/quickjs"
)

// injectHTTPHelpers registers httpGet, httpPost, httpPut, httpPatch, httpDelete,
// httpRequest into the VM. All HTTP calls are proxied through Go.
func injectHTTPHelpers(ctx context.Context, vm *quickjs.VM, s *Sandbox, baseURL string) {
	doHTTP := func(method, path, body string, rawHeaders string) string {
		headers := map[string]string{}
		if rawHeaders != "" {
			_ = json.Unmarshal([]byte(rawHeaders), &headers)
		}
		resp, err := s.http.Do(ctx, method, baseURL+path, headers, body)
		if err != nil {
			return jsonMarshal(map[string]string{"error": err.Error()})
		}
		return jsonMarshal(map[string]any{
			"statusCode": resp.StatusCode,
			"body":       resp.Body,
			"headers":    resp.Headers,
		})
	}

	// httpRequest(method, path, body?, headersJSON?) -> JSON string
	_ = vm.RegisterFunc("httpRequest", func(method, path string, args ...string) string {
		body, headers := "", ""
		if len(args) > 0 {
			body = args[0]
		}
		if len(args) > 1 {
			headers = args[1]
		}
		return doHTTP(method, path, body, headers)
	}, false)

	// httpGet(path, headersJSON?) -> JSON string
	_ = vm.RegisterFunc("httpGet", func(path string, args ...string) string {
		headers := ""
		if len(args) > 0 {
			headers = args[0]
		}
		return doHTTP("GET", path, "", headers)
	}, false)

	// httpPost(path, body?, headersJSON?) -> JSON string
	_ = vm.RegisterFunc("httpPost", func(path string, args ...string) string {
		body, headers := "", ""
		if len(args) > 0 {
			body = args[0]
		}
		if len(args) > 1 {
			headers = args[1]
		}
		return doHTTP("POST", path, body, headers)
	}, false)

	// httpPut(path, body?, headersJSON?) -> JSON string
	_ = vm.RegisterFunc("httpPut", func(path string, args ...string) string {
		body, headers := "", ""
		if len(args) > 0 {
			body = args[0]
		}
		if len(args) > 1 {
			headers = args[1]
		}
		return doHTTP("PUT", path, body, headers)
	}, false)

	// httpPatch(path, body?, headersJSON?) -> JSON string
	_ = vm.RegisterFunc("httpPatch", func(path string, args ...string) string {
		body, headers := "", ""
		if len(args) > 0 {
			body = args[0]
		}
		if len(args) > 1 {
			headers = args[1]
		}
		return doHTTP("PATCH", path, body, headers)
	}, false)

	// httpDelete(path, headersJSON?) -> JSON string
	_ = vm.RegisterFunc("httpDelete", func(path string, args ...string) string {
		headers := ""
		if len(args) > 0 {
			headers = args[0]
		}
		return doHTTP("DELETE", path, "", headers)
	}, false)
}
