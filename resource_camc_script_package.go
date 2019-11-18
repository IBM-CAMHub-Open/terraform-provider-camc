//
// Copyright : IBM Corporation 2016, 2016
//

package main

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/IBM-CAMHub-Open/terraform-provider-camc/common"
)

func resourceCamcScriptPackage() *schema.Resource {
	return &schema.Resource{
		Create: resourceCamcScriptPackageCreate,
		Read:   resourceCamcScriptPackageRead,
		Update: resourceCamcScriptPackageUpdate,
		Delete: resourceCamcScriptPackageDelete,

		Schema: map[string]*schema.Schema{
			"program": &schema.Schema{
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				ForceNew: true,
			},

			"program_sensitive": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					Sensitive: true,
				},
				ForceNew: true,
				Sensitive: true,
			},

			"source": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"source_user": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"source_password": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Sensitive: true,
			},

			"destination": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"query": &schema.Schema{
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				ForceNew: true,
			},

			"query_sensitive": &schema.Schema{
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					Sensitive: true,
				},
				ForceNew: true,
				Sensitive: true,
			},

			"result": &schema.Schema{
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"on_create": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"on_update": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"on_delete": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"remote_host": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"remote_user": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"remote_password": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Sensitive: true,
			},

			"remote_key": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Sensitive: true,
			},

			"bastion_host": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"bastion_user": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"bastion_password": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Sensitive: true,
			},

			"bastion_private_key": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Sensitive: true,
			},	
					
			"bastion_port": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},			

			"trace": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},

			"source_no_check_cert": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
		},
	}
}

func resourceCamcScriptPackageCreate(d *schema.ResourceData, m interface{}) error {
  if (! d.Get("on_create").(bool)){
		// Need to set an ID so that the resource gets created in Terraform
		d.SetId(common.GenUUID())
		var emptyResult map[string]string
		emptyResult = make(map[string]string)
		d.Set("result", emptyResult)
		return nil
	}
	result, err := common.RunScript(d, m)

	if err != nil {
		return err
	}

  d.Set("result", result)
	d.SetId(common.GenUUID())
  return nil
}

func resourceCamcScriptPackageRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceCamcScriptPackageUpdate(d *schema.ResourceData, m interface{}) error {
	if (! d.Get("on_update").(bool)){
		var emptyResult map[string]string
		emptyResult = make(map[string]string)
		d.Set("result", emptyResult)				
		return nil
	}
	return runRequest(d, m)
}

func resourceCamcScriptPackageDelete(d *schema.ResourceData, m interface{}) error {
	if (! d.Get("on_delete").(bool)){
		var emptyResult map[string]string
		emptyResult = make(map[string]string)
		d.Set("result", emptyResult)		
		return nil
	}
	return runRequest(d, m)
}

func runRequest(d *schema.ResourceData, m interface{}) error {
	result, err := common.RunScript(d, m)

	if err != nil {
		return err
	}

  d.Set("result", result)
  return nil
}
