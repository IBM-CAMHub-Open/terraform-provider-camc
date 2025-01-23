1. Bump up the version: 
    - https://github.ibm.com/OpenContent/terraform-provider-camc/blob/master/.camhub.yml#L29
    - https://github.ibm.com/OpenContent/terraform-provider-camc/blob/master/.travis.yml#L48
2. Tag as publish: https://github.ibm.com/OpenContent/terraform-provider-camc
```
git tag -a -m "tagging it to publish" publish
git commit --allow-empty -m "retagging to publish"
git push origin master --tags
```
3. It will trigger the publish branch build and create tag in github.com
    - https://v3.travis.ibm.com/github/OpenContent/terraform-provider-camc/branches
4. Create a tag with v… at github.com
    - https://github.com/IBM-CAMHub-Open/terraform-provider-camc/releases/tag/v0.3.0
5. Once all the binaries are ready, create release from tag and attach the binaries.
    - https://na.artifactory.swg-devops.com/artifactory/orpheus-local-generic/opencontent/terraform-provider/v1.X.X/intel_release-4.1/terraform-provider-camc_v0.3.0
    - https://na.artifactory.swg-devops.com/artifactory/orpheus-local-generic/opencontent/terraform-provider/v1.X.X/power_release-4.1/terraform-provider-camc_v0.3.0_ppc64le
    - https://na.artifactory.swg-devops.com/artifactory/orpheus-local-generic/opencontent/terraform-provider/v1.X.X/s390x_release-4.1/terraform-provider-camc_v0.3.0_s390x
