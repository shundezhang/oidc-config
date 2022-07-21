variable "region" {
  type        = string
  default = "us-east-1"
}
variable "profile" {
  type        = string
  default = "default"
}
variable "private_subnets" {
  type        = list
  default = ["172.30.100.0/24","172.30.101.0/24"]
}
variable "public_subnet" {
  type        = string
  default = "subnet-080da3e9f9d88bc72"
}
variable "key_name" {
  type        = string
  default = "lp-key"
}
variable "vpc_id" {
  type        = string
  default = "vpc-009f7ecec56e66e34"
}
variable "private_route_table_ids" {
  type        = list
  default = ["rtb-0902e8477cb11199d"]
}