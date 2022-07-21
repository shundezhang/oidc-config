
## Usage
The following assumes you have the plugin installed via

```shell
kubectl krew install oidc-config
```

### Show content of K8S_API/.well-known/openid-configuration and K8S_API/openid/v1/jwks from K8s API server
Can use `-oyaml` or `-ojson` to specify output format, by default it is just plain text.
```shell
kubectl oidc-config get
```

### Show content of OIDC config files and upload to S3
S3 bucket and paths are from openid-configuration, thus please set it to the right URL before running this.
```shell
kubectl oidc-config get --upload
```

### Show content of OIDC config files, upload to S3 and create an OIDC provider in IAM with the uploaded contents

```shell
kubectl oidc-config get --upload --create-oidc-provider
```

### Create IAM role for k8s to assume
Only service accounts in sa-namespace can assume this role.
If --allow-all-sas is set, all service accounts in sa-namespace can assuem the role, otherwise only sa-name can assume.
If --create-sa is set, the service account sa-namespace/sa-name will be created in Kubernetes.
```shell
kubectl oidc-config create-role -r [role-name] -p [policy-name] -sa-name [sa-name] -sa-namespace [sa-namespace] --create-sa --allow-all-sas
```

## How it works
Write a brief description of your plugin here.