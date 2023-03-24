resource "clevercloud_nodejs" "%s" {
	name = "%s"
	region = "par"
	min_instance_count = 1
	max_instance_count = 2
	smallest_flavor = "XS"
	biggest_flavor = "M"
	app_folder = "./app"
	environment = {
		MY_KEY = "myval"
	}
}