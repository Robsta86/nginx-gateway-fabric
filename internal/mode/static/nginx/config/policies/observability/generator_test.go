package observability_test

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	ngfAPI "github.com/nginxinc/nginx-gateway-fabric/apis/v1alpha1"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/helpers"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/config/http"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/config/policies"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/config/policies/observability"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/dataplane"
)

func TestGenerate(t *testing.T) {
	ratio := helpers.GetPointer[int32](25)
	zeroRatio := helpers.GetPointer[int32](0)
	context := helpers.GetPointer[ngfAPI.TraceContext](ngfAPI.TraceContextExtract)
	spanName := helpers.GetPointer("my-span")

	tests := []struct {
		name               string
		expExternalStrings []string
		expRedirectStrings []string
		expInternalStrings []string
		policy             policies.Policy
		telemetryConf      dataplane.Telemetry
	}{
		{
			name: "strategy set to default ratio",
			policy: &ngfAPI.ObservabilityPolicy{
				Spec: ngfAPI.ObservabilityPolicySpec{
					Tracing: &ngfAPI.Tracing{
						Strategy: ngfAPI.TraceStrategyRatio,
					},
				},
			},
			expExternalStrings: []string{
				"otel_trace on;",
			},
			expRedirectStrings: []string{
				"otel_trace on;",
			},
			expInternalStrings: []string{
				"otel_span_name $request_uri_path;",
			},
		},
		{
			name: "strategy set to custom ratio",
			policy: &ngfAPI.ObservabilityPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-policy",
					Namespace: "test-namespace",
				},
				Spec: ngfAPI.ObservabilityPolicySpec{
					Tracing: &ngfAPI.Tracing{
						Strategy: ngfAPI.TraceStrategyRatio,
						Ratio:    ratio,
					},
				},
			},
			expExternalStrings: []string{
				"otel_trace $otel_ratio_25;",
			},
			expRedirectStrings: []string{
				"otel_trace $otel_ratio_25;",
			},
			expInternalStrings: []string{
				"otel_span_name $request_uri_path;",
			},
		},
		{
			name: "strategy set to zero ratio",
			policy: &ngfAPI.ObservabilityPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-policy",
					Namespace: "test-namespace",
				},
				Spec: ngfAPI.ObservabilityPolicySpec{
					Tracing: &ngfAPI.Tracing{
						Strategy: ngfAPI.TraceStrategyRatio,
						Ratio:    zeroRatio,
					},
				},
			},
			expExternalStrings: []string{
				"otel_trace off;",
			},
			expRedirectStrings: []string{
				"otel_trace off;",
			},
			expInternalStrings: []string{
				"otel_span_name $request_uri_path;",
			},
		},
		{
			name: "strategy set to parent",
			policy: &ngfAPI.ObservabilityPolicy{
				Spec: ngfAPI.ObservabilityPolicySpec{
					Tracing: &ngfAPI.Tracing{
						Strategy: ngfAPI.TraceStrategyParent,
					},
				},
			},
			expExternalStrings: []string{
				"otel_trace $otel_parent_sampled;",
			},
			expRedirectStrings: []string{
				"otel_trace $otel_parent_sampled;",
			},
			expInternalStrings: []string{
				"otel_span_name $request_uri_path;",
			},
		},
		{
			name: "context is set",
			policy: &ngfAPI.ObservabilityPolicy{
				Spec: ngfAPI.ObservabilityPolicySpec{
					Tracing: &ngfAPI.Tracing{
						Context: context,
					},
				},
			},
			expExternalStrings: []string{
				"otel_trace off;",
				"otel_trace_context extract;",
			},
			expRedirectStrings: []string{
				"otel_trace off;",
				"otel_trace_context extract;",
			},
			expInternalStrings: []string{
				"otel_span_name $request_uri_path;",
			},
		},
		{
			name: "spanName is set",
			policy: &ngfAPI.ObservabilityPolicy{
				Spec: ngfAPI.ObservabilityPolicySpec{
					Tracing: &ngfAPI.Tracing{
						SpanName: spanName,
					},
				},
			},
			expExternalStrings: []string{
				"otel_trace off;",
				"otel_span_name \"my-span\";",
			},
			expRedirectStrings: []string{
				"otel_trace off;",
			},
			expInternalStrings: []string{
				"otel_span_name \"my-span\";",
			},
		},
		{
			name: "span attributes set",
			policy: &ngfAPI.ObservabilityPolicy{
				Spec: ngfAPI.ObservabilityPolicySpec{
					Tracing: &ngfAPI.Tracing{
						SpanAttributes: []ngfAPI.SpanAttribute{
							{Key: "test-key", Value: "test-value"},
						},
					},
				},
			},
			expExternalStrings: []string{
				"otel_trace off;",
				"otel_span_attr \"test-key\" \"test-value\";",
			},
			expRedirectStrings: []string{
				"otel_trace off;",
			},
			expInternalStrings: []string{
				"otel_span_name $request_uri_path;",
				"otel_span_attr \"test-key\" \"test-value\";",
			},
		},
		{
			name: "global span attributes set",
			policy: &ngfAPI.ObservabilityPolicy{
				Spec: ngfAPI.ObservabilityPolicySpec{
					Tracing: &ngfAPI.Tracing{},
				},
			},
			telemetryConf: dataplane.Telemetry{
				SpanAttributes: []dataplane.SpanAttribute{
					{Key: "test-global-key", Value: "test-global-value"},
				},
			},
			expExternalStrings: []string{
				"otel_trace off;",
				"otel_span_attr \"test-global-key\" \"test-global-value\";",
			},
			expRedirectStrings: []string{
				"otel_trace off;",
			},
			expInternalStrings: []string{
				"otel_span_name $request_uri_path;",
				"otel_span_attr \"test-global-key\" \"test-global-value\";",
			},
		},
		{
			name: "all fields populated",
			policy: &ngfAPI.ObservabilityPolicy{
				Spec: ngfAPI.ObservabilityPolicySpec{
					Tracing: &ngfAPI.Tracing{
						Strategy: ngfAPI.TraceStrategyRatio,
						Context:  context,
						SpanName: spanName,
						SpanAttributes: []ngfAPI.SpanAttribute{
							{Key: "test-key", Value: "test-value"},
						},
					},
				},
			},
			telemetryConf: dataplane.Telemetry{
				SpanAttributes: []dataplane.SpanAttribute{
					{Key: "test-global-key", Value: "test-global-value"},
				},
			},
			expExternalStrings: []string{
				"otel_trace on;",
				"otel_trace_context extract;",
				"otel_span_name \"my-span\";",
				"otel_span_attr \"test-key\" \"test-value\";",
				"otel_span_attr \"test-global-key\" \"test-global-value\";",
			},
			expRedirectStrings: []string{
				"otel_trace on;",
				"otel_trace_context extract;",
			},
			expInternalStrings: []string{
				"otel_span_name \"my-span\";",
				"otel_span_attr \"test-key\" \"test-value\";",
				"otel_span_attr \"test-global-key\" \"test-global-value\";",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)

			generator := observability.NewGenerator(test.telemetryConf)

			for _, locType := range []http.LocationType{
				http.ExternalLocationType, http.RedirectLocationType, http.InternalLocationType,
			} {
				var expStrings []string
				var resFiles policies.GenerateResultFiles
				switch locType {
				case http.ExternalLocationType:
					expStrings = test.expExternalStrings
					resFiles = generator.GenerateForLocation([]policies.Policy{test.policy}, http.Location{Type: locType})
				case http.RedirectLocationType:
					expStrings = test.expRedirectStrings
					resFiles = generator.GenerateForLocation([]policies.Policy{test.policy}, http.Location{Type: locType})
				case http.InternalLocationType:
					expStrings = test.expInternalStrings
					resFiles = generator.GenerateForInternalLocation([]policies.Policy{test.policy})
				}

				g.Expect(resFiles).To(HaveLen(1))

				content := string(resFiles[0].Content)

				if len(expStrings) == 0 {
					g.Expect(content).To(Equal(""))
				}

				for _, str := range expStrings {
					g.Expect(content).To(ContainSubstring(str))
				}
			}
		})
	}
}

func TestGenerateNoPolicies(t *testing.T) {
	g := NewWithT(t)

	generator := observability.NewGenerator(dataplane.Telemetry{})

	resFiles := generator.GenerateForLocation([]policies.Policy{}, http.Location{})
	g.Expect(resFiles).To(BeEmpty())

	resFiles = generator.GenerateForLocation([]policies.Policy{&ngfAPI.ClientSettingsPolicy{}}, http.Location{})
	g.Expect(resFiles).To(BeEmpty())

	resFiles = generator.GenerateForInternalLocation([]policies.Policy{})
	g.Expect(resFiles).To(BeEmpty())

	resFiles = generator.GenerateForInternalLocation([]policies.Policy{&ngfAPI.ClientSettingsPolicy{}})
	g.Expect(resFiles).To(BeEmpty())
}
