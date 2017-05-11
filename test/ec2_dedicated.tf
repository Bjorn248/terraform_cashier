resource "aws_instance" "dedicated_instance" {
  # Custom AMI for Mongo Instances
  ami           = "ami-c58c1dd3"
  instance_type = "r3.xlarge"
  key_name      = "REDACTED"
  ebs_optimized = true
  tenancy       = "dedicated"

  root_block_device {
    volume_type = "io1"
    volume_size = 50

    iops = 250
  }
}
