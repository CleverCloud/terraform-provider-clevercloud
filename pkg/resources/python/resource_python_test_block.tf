resource "clevercloud_python" "%s" {
	name = "%s"
	region = "par"
	min_instance_count = 1
	max_instance_count = 2
	smallest_flavor = "XS"
	biggest_flavor = "M"
	redirect_https = true
	sticky_sessions = true
	app_folder = "./app"
	python_version = "2.7"
	pip_requirements = "requirements.txt"
	environment = {
		MY_KEY = "myval"
	}
	hooks {
		post_build = "echo \"build is OK!\""
	}
	dependencies = []
}
