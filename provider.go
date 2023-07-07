//
// Copyright : IBM Corporation 2016, 2016
//

package main

import (
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			"camc_bootstrap":               resourceCamcBootstrap(),
			"camc_scriptpackage":           resourceCamcScriptPackage(),
			"camc_updatable_scriptpackage": resourceCamcUpdatableScriptPackage(),
			"camc_softwaredeploy":          resourceCamcSoftwaredeploy(),
			"camc_vaultitem":               resourceCamcVaultitem(),
		},
	}
}

func jsonStateFunc(value interface{}) string {
	// Parse and re-stringify the JSON to make sure it's always kept
	// in a normalized form.
	in, ok := value.(string)
	if !ok {
		return "null"
	}
	var tmp map[string]interface{}

	json.Unmarshal([]byte(in), &tmp)

	jsonValue, _ := json.Marshal(&tmp)
	return string(jsonValue)
}
