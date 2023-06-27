# Project Command Runner
## About
Project Command Runner is designed to run a given command in specified projects.
The projects are determined by the provided configuration and the given filters
to the command.

## Creating a configuration
Example configuration:
```yaml
projects:
  authentication:
    path: http/authentication
    tags:
      - http
      - nodejs
  posts:
    path: http/posts
    tags:
      - http
      - golang
  invoices:
    path: cron/invoices
    tags:
      - cron
      - golang
```
Every defined project should have a path to the folder of the project.
You could provide tags by which to later filter out or include projects when
executing a command.

## Using the command runner
You need to create a configuration file with your projects. Once you have your
configuration file ready you can use the commander as follows:
```sh
# This command will run the command 'pwd' in every project
./command_runner -c my_config.yml -X 'pwd'

# This command will run the command 'pwd' in every project except the 'invoices'
# project. You can pass multiple '-e' flags
./command_runner -c my_config.yml -e invoices -X 'pwd'

# Will run the 'pwd' command in every project which has the tag 'http' but not
# in the posts project if it is filtered in
./command_runner -c my_config.yml --tag-search http -e posts -X 'pwd'

# Will run the 'pwd' command in every project which does not have the 'http' tag
./command_runner -c my_config.yml --tag-exclude http -X 'pwd'
```

> The `-c` flag is optional. You could omit it but you would have to name your
> configuration file `commander.yml`

