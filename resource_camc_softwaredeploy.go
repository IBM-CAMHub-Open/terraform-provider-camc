//
// Copyright : IBM Corporation 2016, 2016
//

package main

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/IBM-CAMHub-Open/terraform-provider-camc/common"
)

func resourceCamcSoftwaredeploy() *schema.Resource {
	return &schema.Resource{
		Create: resourceCamcSoftwaredeployCreate,
		Read:   resourceCamcSoftwaredeployRead,
		Update: resourceCamcSoftwaredeployUpdate,
		Delete: resourceCamcSoftwaredeployDelete,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
				Required: false,
			},

			"camc_endpoint": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"data": &schema.Schema{
				Type:      schema.TypeString,
				Optional:  true,
				StateFunc: jsonStateFunc,
				Default:   "null",
				ForceNew:  true,
			},

			"username": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
				ForceNew: true,
			},

			"password": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
				ForceNew: true,
			},

			"skip_ssl_verify": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
				ForceNew: true,
			},

			"cert_file": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
				ForceNew: true,
			},

			"key_file": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
				ForceNew: true,
			},

			"ca_file": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
				ForceNew: true,
			},

			"trace": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},

			"access_token": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceCamcSoftwaredeployCreate(d *schema.ResourceData, m interface{}) error {
	//get create camc_endpoint
	camc_endpoint := d.Get("camc_endpoint").(string)
	if camc_endpoint != "" {
		_, err := common.MakeRequest(d, m, "POST")
		if err == nil {
			d.SetId(common.GenUUID())
			return nil
		} else {
			return err
		}
	} else {
		return nil
	}
}

func resourceCamcSoftwaredeployRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceCamcSoftwaredeployUpdate(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceCamcSoftwaredeployDelete(d *schema.ResourceData, m interface{}) error {
	return nil
}
