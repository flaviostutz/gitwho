# CONTRIBUTING

Thank you for being interested in helping with GitWho!

## The best way to start

- Look at the "Issues" and choose something to implement
- Fork this project to your account
- Develop your code and create unit tests
- Create a PR pointing to main branch
- Use make targets for common tasks (they are the same that are run during pipelines)

```sh
make build
make lint
make test
```

## Questions and discussions

- Discuss design and implementation details on the related Issue comments
- If you have a more generic question, create a new Issue

## Bugs and feature requests

- If you find any bug, create a new Issue describing the usage context and what kind of error you had. If possible, send a link of the repo you are analysing with GitWho or a screenshot with details about it.

- If you want a new feature, open an Issue, explain your real use case, detail the kind of problems you have nowadays and how you think GitWho could help you in practice.

## Prepare your development environment

- Install VSCode and the Golang extension
- Git clone this repo
- Type `make run-ownership` to run Gitwho and check the output

## Pipelines

- Everytime a PR is created, testing, linting and building will be run. Your PR will only be analysed after a successful pipeline run
- Tag the repo when you are prepared to release a new version. The pipeline will be run automatically and do the publishing
