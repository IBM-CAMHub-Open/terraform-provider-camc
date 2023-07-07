# Terrafrom Provider for IBM Cloud Automation Manager Content

## Maintainers

This provider is maintained by IBM Corp. 

## Requirements

- Terraform Plugin SDK v2
- GO (GO version must be 1.20.x or greater)

## Using the provider

Follow the steps below to get the information on how to use the provider

* Open [IBM Cloud Pak for Watson AIOps documentation.](https://www.ibm.com/docs/en/cloud-paks/cloud-pak-watson-aiops)
* Select the version of CPWAIOPs(viz CAM) you are using 
* Navigate to  Reference > Infrastructure Automation Managed services > Terraform CAMC provider

## Building the provider

  #Set the variables for terrafrom version,   
  #terraform-provider-camc branch and to turn off GO Modules.  
  export GOPATH=<your_go_path>  
  export BRANCH_TO_BUILD=master  
  export GO111MODULE=auto  
  export PROVIDER_VERSION=<your_new_provider_version>
      
  #Clone terrafrom provider camc  
  git clone https://github.com/IBM-CAMHub-Open/terraform-provider-camc.git  
  
  #Checkout the branch set in BRANCH_TO_BUILD  
  cd terraform-provider-camc/   
  git checkout ${BRANCH_TO_BUILD}  
    
  #Build  
  go build -o terraform-provider-camc  
  mv $GOPATH/src/github.com/IBM-CAMHub-Open/terraform-provider-camc/terraform-provider-camc $GOPATH/bin/terraform-provider-camc_v${PROVIDER_VERSION}


Copyright IBM Corp. 2023




