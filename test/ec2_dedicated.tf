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

  ebs_block_device {
    device_name = "/dev/sdb"
    snapshot_id = "${lookup(var.Mongo_Snapshots, var.primary_region)}"
    volume_type = "io1"
    volume_size = 100
    iops        = 500
  }

  tags {
    Name                 = "C1_MongoDB1_West"
    managed_by_terraform = true
  }

  # Clear iptables and initialize data EBS volume
  provisioner "remote-exec" {
    inline = [
      "sudo iptables -F",
      "sudo chkconfig iptables off",
      "sudo /root/init_data_volume.sh",
    ]

    connection {
      type        = "ssh"
      user        = "centos"
      private_key = "${file("/Users/bstange/.ssh/REDACTED.pem")}"
    }
  }

  provisioner "chef" {
    environment     = "c1"
    run_list        = ["mongodb::default"]
    node_name       = "C1_MongoDB1_West_${self.id}"
    secret_key      = "${file("./dbsec")}"
    server_url      = "https://api.chef.io/organizations/REDACTED"
    user_name       = "terraform_validator"
    recreate_client = false
    user_key        = "${file("../../terraform_validator.pem")}"
    version         = "12.13.37"

    connection {
      type = "ssh"
      user = "centos"

      # TODO: Figure out how to get the path to load from a variable
      private_key = "${file("/Users/bstange/.ssh/REDACTED.pem")}"
    }
  }
}
