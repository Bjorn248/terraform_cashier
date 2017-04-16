resource "aws_db_instance" "Sample_MySQL" {
  allocated_storage    = 10
  storage_type         = "gp2"
  engine               = "mysql"
  engine_version       = "5.6.17"
  instance_class       = "db.t2.large"
  name                 = "mydb"
  username             = "foo"
  password             = "bar"
  db_subnet_group_name = "my_database_subnet_group"
  parameter_group_name = "default.mysql5.6"
}

resource "aws_db_instance" "Sample_MySQL_MultiAZ" {
  allocated_storage    = 10
  storage_type         = "gp2"
  multi_az             = true
  engine               = "mysql"
  engine_version       = "5.6.17"
  instance_class       = "db.t2.large"
  name                 = "mydb"
  username             = "foo"
  password             = "bar"
  db_subnet_group_name = "my_database_subnet_group"
  parameter_group_name = "default.mysql5.6"
}

resource "aws_db_instance" "Sample_MySQL_SingleAZ" {
  allocated_storage    = 10
  storage_type         = "gp2"
  multi_az             = false
  engine               = "mysql"
  engine_version       = "5.6.17"
  instance_class       = "db.t2.large"
  name                 = "mydb"
  username             = "foo"
  password             = "bar"
  db_subnet_group_name = "my_database_subnet_group"
  parameter_group_name = "default.mysql5.6"
}
