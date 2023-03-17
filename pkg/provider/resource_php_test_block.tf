resource "clevercloud_php" "%s" {
	name = "%s"
	region = "par"
	min_instance_count = 1
	max_instance_count = 2
	smallest_flavor = "XS"
	biggest_flavor = "M"
    php_version = "8"
	additional_vhosts = [ "toto-tf5283457829345.com" ]
}
