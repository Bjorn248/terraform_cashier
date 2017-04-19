resource "aws_instance" "dedicated_instance" {
  # Custom AMI for Mongo Instances
  ami                    = "${lookup(var.Mongo_AMIs, var.primary_region)}"
  instance_type          = "r3.xlarge"
  subnet_id              = "${aws_subnet.Dev_mongo_west.id}"
  key_name               = "REDACTED"
  vpc_security_group_ids = ["${aws_security_group.REDACTED_Dev_PoC_SG_West_Internal.id}"]
  user_data              = "${file("userdata_hosts.sh")}"
  ebs_optimized          = true
  tenancy                = "dedicated"

  root_block_device {
    volume_type = "io1"
    volume_size = 50

    iops = 250
  }
}
