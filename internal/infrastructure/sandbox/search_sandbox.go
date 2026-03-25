package sandbox

import (
	"modernc.org/quickjs"
)

// injectSearchHelpers registers search(), getEndpoints(), getSpec() into the VM.
func injectSearchHelpers(vm *quickjs.VM, s *Sandbox, apiName string) {
	// search(query) -> JSON string
	_ = vm.RegisterFunc("search", func(query string) string {
		results, err := s.search.Search(apiName, query)
		if err != nil {
			return jsonMarshal(map[string]string{"error": err.Error()})
		}
		return jsonMarshal(results)
	}, false)

	// getEndpoints() -> JSON string
	_ = vm.RegisterFunc("getEndpoints", func() string {
		endpoints, err := s.registry.GetEndpoints(apiName)
		if err != nil {
			return jsonMarshal(map[string]string{"error": err.Error()})
		}
		return jsonMarshal(endpoints)
	}, false)

	// getSpec() -> JSON string
	_ = vm.RegisterFunc("getSpec", func() string {
		spec, err := s.registry.GetAPI(apiName)
		if err != nil {
			return jsonMarshal(map[string]string{"error": err.Error()})
		}
		return jsonMarshal(spec)
	}, false)
}
