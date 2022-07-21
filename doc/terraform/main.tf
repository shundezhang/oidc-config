provider "aws" {
  region = var.region
  profile = var.profile
}

module "nat-instance" {
  source  = "int128/nat-instance/aws"
  version = "2.0.1"
  name = "nat-inst"
  private_subnets_cidr_blocks = var.private_subnets
  public_subnet = var.public_subnet
  vpc_id = var.vpc_id
  private_route_table_ids = var.private_route_table_ids
  key_name = var.key_name
}

resource "aws_security_group_rule" "nat_ssh" {
  security_group_id = module.nat-instance.sg_id
  type              = "ingress"
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 22
  to_port           = 22
  protocol          = "tcp"
}

resource "aws_eip" "nat" {
  network_interface = module.nat-instance.eni_id
  vpc      = true
  tags = {
    "Name" = "nat-instance-main"
  }
}