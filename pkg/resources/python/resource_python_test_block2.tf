resource "clevercloud_python" "%s" {
	name = "%s"
	region = "par"
	min_instance_count = 1
	max_instance_count = 2
	smallest_flavor = "XS"
	biggest_flavor = "M"
	deployment {
		repository = "https://github.com/CleverCloud/flask-example.git"
	}
}
