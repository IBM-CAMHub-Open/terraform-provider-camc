# Terrafrom Provider for IBM Cloud Automation Manager Content

## Maintainers

This provider is maintained by IBM Corp. 

## Requirements

Terraform 0.11.7 or 0.12.21
GO (GO version for Terrafrom version 0.12.21 must be 1.13.x)

## Using the provider

Follow the steps below to get the information on how to use the provider

* Open [IBM Cloud Automation Manager Documentation.](https://www.ibm.com/support/knowledgecenter/en/SS2L37/product_welcome_cloud_automation_manager.html)
* Select the version of CAM you are using 
* Navigate to  Reference > Terraform CAMC provider

## Building the provider

  #Set the variables for terrafrom version,   
  #terraform-provider-camc branch and to turn off GO Modules.  
  export GOPATH=<your_go_path>  
  export CAM_TERRAFORM_VERSION=0.11.7    
  export BRANCH_TO_BUILD=master  
  export GO111MODULE=off  
  export PROVIDER_VERSION=<your_new_provider_version>
  
  #Create directories  
  mkdir -p $GOPATH/src/github.com/hashicorp  
  mkdir -p $GOPATH/src/github.com/IBM-CAMHub-Open  
  mkdir -p $GOPATH/bin  
  cd $GOPATH/src/github.com/hashicorp  
  
  #Clone terrafrom  
  git clone https://github.com/hashicorp/terraform.git
  
  #Checkout the branch set in CAM_TERRAFORM_VERSION  
  cd terraform/  
  git checkout v${CAM_TERRAFORM_VERSION}  
  
  cd $GOPATH/src/github.com/IBM-CAMHub-Open
  
  #Clone terrafrom provider camc  
  git clone https://github.com/IBM-CAMHub-Open/terraform-provider-camc.git  
  
  #Checkout the branch set in BRANCH_TO_BUILD  
  cd terraform-provider-camc/   
  git checkout ${BRANCH_TO_BUILD}  
  
  #Get additional GO package  
  go get github.com/pkg/sftp  
  go get -u golang.org/x/crypto/ssh  
  
  #Build  
  cd $GOPATH/src/github.com/IBM-CAMHub-Open/terraform-provider-camc  
  go build -o terraform-provider-camc  
  mv $GOPATH/src/github.com/IBM-CAMHub-Open/terraform-provider-camc/terraform-provider-camc $GOPATH/bin/terraform-provider-camc_v${PROVIDER_VERSION}


Copyright IBM Corp. 2019




