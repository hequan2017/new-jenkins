package workflow

import (
	"encoding/json"
	"strings"
	"testing"

	modelWorkflow "github.com/flipped-aurora/gin-vue-admin/server/model/workflow"
)

func TestExpandConfigSupportsBothParamSyntaxes(t *testing.T) {
	input := []byte(`{"braced":"${param.env}","plain":"$param.version","shell":"$PATH","unknown":"${param.missing}"}`)
	got := string(expandConfig(input, map[string]string{"env": "prod", "version": "1.2.3"}))
	for _, want := range []string{`"braced":"prod"`, `"plain":"1.2.3"`, `"shell":"$PATH"`, `"unknown":"${param.missing}"`} {
		if !strings.Contains(got, want) {
			t.Fatalf("结果 %s 未包含 %s", got, want)
		}
	}
}

func TestExpandConfigEscapesParamValuesAsJSON(t *testing.T) {
	got := expandConfig([]byte(`{"body":"hello ${param.name}"}`), map[string]string{"name": `a"b`})
	var decoded map[string]string
	if err := json.Unmarshal(got, &decoded); err != nil {
		t.Fatalf("参数替换后 JSON 非法: %s: %v", got, err)
	}
	if decoded["body"] != `hello a"b` {
		t.Fatalf("参数值不正确: %q", decoded["body"])
	}
}

func TestValidateAndFillParamsRejectsUnknownAndDuplicateParams(t *testing.T) {
	schema := []byte(`[{"name":"env","type":"string"}]`)
	if _, err := validateAndFillParams(schema, []modelWorkflow.ParamValue{{Name: "extra", Value: "x"}}); err == nil {
		t.Fatal("未声明参数应被拒绝")
	}
	if _, err := validateAndFillParams(schema, []modelWorkflow.ParamValue{{Name: "env", Value: "a"}, {Name: "env", Value: "b"}}); err == nil {
		t.Fatal("重复参数应被拒绝")
	}
}
