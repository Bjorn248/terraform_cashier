resource "aws_instance" "C1_MongoDB1_West" {
  # Custom AMI for Mongo Instances
  ami                    = "${lookup(var.Mongo_AMIs, var.primary_region)}"
  instance_type          = "r3.xlarge"
  subnet_id              = "${aws_subnet.Dev_mongo_west.id}"
  key_name               = "REDACTED"
  vpc_security_group_ids = ["${aws_security_group.REDACTED_Dev_PoC_SG_West_Internal.id}"]
  user_data              = "${file("userdata_hosts.sh")}"
  ebs_optimized          = true

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

resource "aws_instance" "C1_MongoDB2_West" {
  # Custom AMI for Mongo Instances
  ami                    = "${lookup(var.Mongo_AMIs, var.primary_region)}"
  instance_type          = "r3.xlarge"
  subnet_id              = "${aws_subnet.Dev_mongo_west.id}"
  key_name               = "REDACTED"
  vpc_security_group_ids = ["${aws_security_group.REDACTED_Dev_PoC_SG_West_Internal.id}"]
  user_data              = "${file("userdata_hosts.sh")}"
  ebs_optimized          = true

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
    Name                 = "C1_MongoDB2_West"
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
    node_name       = "C1_MongoDB2_West_${self.id}"
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

resource "aws_instance" "C1_MongoDB3_West" {
  # Custom AMI for Mongo Instances
  ami                    = "${lookup(var.Mongo_AMIs, var.primary_region)}"
  instance_type          = "r3.xlarge"
  subnet_id              = "${aws_subnet.Dev_mongo_west.id}"
  key_name               = "REDACTED"
  vpc_security_group_ids = ["${aws_security_group.REDACTED_Dev_PoC_SG_West_Internal.id}"]
  user_data              = "${file("userdata_hosts.sh")}"
  ebs_optimized          = true

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
    Name                 = "C1_MongoDB3_West"
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
    node_name       = "C1_MongoDB3_West_${self.id}"
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

resource "aws_instance" "Smaller_Instance" {
  # Custom AMI for Mongo Instances
  ami                    = "${lookup(var.Mongo_AMIs, var.primary_region)}"
  instance_type          = "m3.large"
  subnet_id              = "${aws_subnet.Dev_mongo_west.id}"
  key_name               = "REDACTED"
  vpc_security_group_ids = ["${aws_security_group.REDACTED_Dev_PoC_SG_West_Internal.id}"]
  user_data              = "${file("userdata_hosts.sh")}"
  ebs_optimized          = true

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
    Name                 = "Smaller_Instance"
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
}

resource "null_resource" "configure_mongo_replicaset_and_backup" {
  triggers {
    replicaset_instance_ids = "${aws_instance.C1_MongoDB1_West.id},${aws_instance.C1_MongoDB2_West.id},${aws_instance.C1_MongoDB3_West.id}"
  }

  provisioner "chef" {
    attributes_json = <<-EOF
        {
            "mongodb": {
                "replicaset_members": {
                    "0": "${aws_instance.C1_MongoDB1_West.private_ip}",
                    "1": "${aws_instance.C1_MongoDB2_West.private_ip}"
                },
                "replicaset_member_low_priority": {
                    "2": "${aws_instance.C1_MongoDB3_West.private_ip}"
                }
            }
        }
    EOF

    environment     = "c1"
    run_list        = ["mongodb::replicate", "mongodb::auth"]
    node_name       = "C1_MongoDB1_West_${aws_instance.C1_MongoDB1_West.id}"
    secret_key      = "${file("./dbsec")}"
    server_url      = "https://api.chef.io/organizations/REDACTED"
    user_name       = "terraform_validator"
    recreate_client = false
    user_key        = "${file("../../terraform_validator.pem")}"
    version         = "12.13.37"

    connection {
      type = "ssh"
      user = "centos"
      host = "${aws_instance.C1_MongoDB1_West.public_ip}"

      # TODO: Figure out how to get the path to load from a variable
      private_key = "${file("/Users/bstange/.ssh/REDACTED.pem")}"
    }
  }

  provisioner "chef" {
    environment     = "c1"
    run_list        = ["mongodb::backup"]
    node_name       = "C1_MongoDB3_West_${aws_instance.C1_MongoDB3_West.id}"
    secret_key      = "${file("./dbsec")}"
    server_url      = "https://api.chef.io/organizations/REDACTED"
    user_name       = "terraform_validator"
    recreate_client = false
    user_key        = "${file("../../terraform_validator.pem")}"
    version         = "12.13.37"

    connection {
      type = "ssh"
      user = "centos"
      host = "${aws_instance.C1_MongoDB3_West.public_ip}"

      # TODO: Figure out how to get the path to load from a variable
      private_key = "${file("/Users/bstange/.ssh/REDACTED.pem")}"
    }
  }
}
