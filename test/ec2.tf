resource "aws_instance" "shared1" {
  # Custom AMI for Mongo Instances
  ami           = "ami-c58c1dd3"
  instance_type = "r4.xlarge"
  key_name      = "REDACTED"
  ebs_optimized = true

  root_block_device {
    volume_type = "io1"
    volume_size = 50

    iops = 250
  }
}

resource "aws_instance" "shared2" {
  # Custom AMI for Mongo Instances
  ami           = "ami-c58c1dd3"
  instance_type = "r4.xlarge"
  key_name      = "REDACTED"
  ebs_optimized = true

  root_block_device {
    volume_type = "io1"
    volume_size = 50

    iops = 250
  }
}

resource "aws_instance" "shared3" {
  # Custom AMI for Mongo Instances
  ami           = "ami-c58c1dd3"
  instance_type = "r4.xlarge"
  key_name      = "REDACTED"
  ebs_optimized = true

  root_block_device {
    volume_type = "io1"
    volume_size = 50

    iops = 250
  }
}

resource "aws_instance" "shared4" {
  # Custom AMI for Mongo Instances
  ami           = "ami-c58c1dd3"
  instance_type = "r3.xlarge"
  key_name      = "REDACTED"
  ebs_optimized = true

  root_block_device {
    volume_type = "io1"
    volume_size = 50

    iops = 250
  }
}

resource "aws_instance" "shared5" {
  # Custom AMI for Mongo Instances
  ami           = "ami-c58c1dd3"
  instance_type = "r3.xlarge"
  key_name      = "REDACTED"
  ebs_optimized = true

  root_block_device {
    volume_type = "io1"
    volume_size = 50

    iops = 250
  }
}

resource "aws_instance" "shared6" {
  # Custom AMI for Mongo Instances
  ami           = "ami-c58c1dd3"
  instance_type = "r3.xlarge"
  key_name      = "REDACTED"
  ebs_optimized = true

  root_block_device {
    volume_type = "io1"
    volume_size = 50

    iops = 250
  }
}

resource "aws_instance" "shared7" {
  # Custom AMI for Mongo Instances
  ami           = "ami-c58c1dd3"
  instance_type = "m4.large"
  key_name      = "REDACTED"
  ebs_optimized = true

  root_block_device {
    volume_type = "io1"
    volume_size = 50

    iops = 250
  }
}

resource "aws_instance" "shared8" {
  # Custom AMI for Mongo Instances
  ami           = "ami-c58c1dd3"
  instance_type = "m4.xlarge"
  key_name      = "REDACTED"
  ebs_optimized = true

  root_block_device {
    volume_type = "io1"
    volume_size = 50

    iops = 250
  }
}

resource "null_resource" "for_tests" {}
