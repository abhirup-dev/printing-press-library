// Copyright 2026 abhirup and contributors. Licensed under Apache-2.0. See LICENSE.
// Hand-authored tests for the novel diff helpers (printing-press preserved file).

package cli

import (
	"encoding/json"
	"testing"
)

func TestUptFlattenJSON(t *testing.T) {
	rec := json.RawMessage(`{"modelName":"X","aiEngineConfig":{"inputs":[{"name":"A","inputType":"String"}],"predictEngineType":"AppEngine"},"active":false}`)
	var v any
	if err := json.Unmarshal(rec, &v); err != nil {
		t.Fatal(err)
	}
	flat := map[string]string{}
	uptFlattenJSON("", v, flat)
	want := map[string]string{
		"modelName":                          "X",
		"aiEngineConfig.predictEngineType":   "AppEngine",
		"aiEngineConfig.inputs[0].name":      "A",
		"aiEngineConfig.inputs[0].inputType": "String",
		"active":                             "false",
	}
	for k, w := range want {
		if flat[k] != w {
			t.Errorf("flat[%s] = %q, want %q", k, flat[k], w)
		}
	}
}

func TestUptNormalizeForDiffNeutralizesEnvHosts(t *testing.T) {
	dev := "https://api.dev.unitypredict.net/api/public/models/x/files/MODELDESCRIPTION.html"
	prod := "https://api.prod.unitypredict.com/api/public/models/x/files/MODELDESCRIPTION.html"
	if uptNormalizeForDiff(dev) != uptNormalizeForDiff(prod) {
		t.Errorf("env-host URLs should normalize equal:\n %q\n %q", uptNormalizeForDiff(dev), uptNormalizeForDiff(prod))
	}
}

func TestUptVolatileExclusions(t *testing.T) {
	for _, k := range []string{"stateInfo", "modifiedDate", "totalInferences"} {
		if !uptVolatileFields[k] {
			t.Errorf("%s should be volatile-excluded", k)
		}
	}
	if uptVolatileFields["modelName"] {
		t.Error("modelName must not be excluded")
	}
	if topSegment("aiEngineConfig.inputs[0].name") != "aiEngineConfig" {
		t.Error("topSegment broken")
	}
}
