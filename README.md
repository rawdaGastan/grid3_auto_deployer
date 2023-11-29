# cloud4students

[![Codacy Badge](https://app.codacy.com/project/badge/Grade/cd6e18aac6be404ab89ec160b4b36671)](https://www.codacy.com/gh/codescalers/cloud4students/dashboard?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=codescalers/cloud4students&amp;utm_campaign=Badge_Grade) <a href='https://github.com/jpoles1/gopherbadger' target='_blank'>![gopherbadger-tag-do-not-edit](https://img.shields.io/badge/Go%20Coverage-63.1%25-yellow.svg?longCache=true&style=flat)</a> [![Testing](https://github.com/codescalers/cloud4students/actions/workflows/gotest.yml/badge.svg?branch=development)](https://github.com/codescalers/cloud4students/actions/workflows/gotest.yml) [![Testing](https://github.com/codescalers/cloud4students/actions/workflows/golint.yml/badge.svg?branch=development)](https://github.com/codescalers/cloud4students/actions/workflows/golint.yml) [![Testing](https://github.com/codescalers/cloud4students/actions/workflows/vuelint.yml/badge.svg?branch=development)](https://github.com/codescalers/cloud4students/actions/workflows/vuelint.yml)

cloud4students aims to help students deploy their projects on Threefold Grid.

## Requirements

- docker-compose

## Build

First create `config.json` check [configuration](#configuration)

To build backend and frontend images

```bash
docker-compose build
```

## Run

First create `config.json` check [configuration](#configuration)

To run backend and frontend:

```bash
docker-compose up
```

### Configuration

Before building or running backend image create `config.json` in `server` dir.

example `config.json`:

```json
{
    "server": {
        "host": "localhost, required",
        "port": ":3000, required",
        "redisHost": "redis host like `localhost`, required",
        "redisPort": "redis port like `6379`, required",
        "redisPass": "<redis password, required>"  
    },
    "mailSender": {
        "email": "your sendgrid account sender, required",
        "sendgrid_key": "<sendgrid-key>, required",
        "timeout": "<the timeout for app mail verification codes in seconds, required>"
    },
    "database": {
        "file": "<the path of the database file you have or you want to create, required>"
    },
    "token": {
        "secret": "<your secret for the jwt tokens, required>",
        "timeout": "<the timeout of the jwt token in seconds, required>"
    },
    "account": {
        "mnemonics": "<your account mnemonic to be used for the deployments, required>",
        "network": "<grid-network, It can be main, qa, test, dev only, required>"
    },
    "version": "the version of your api like `v1`, required",
    "admins": ["<a set of the user emails you want to make admins>"],
    "notifyAdminsIntervalHours": "<the interval between admins notifications in hours, optional>",
    "adminSSHKey": "<an ssh key to be put with every deployment to prevent losing the vm if the user changed his ssh keys. optional>"
}
```

## Frontend

check frontend [README](client/README.md)

## Backend

check backend [README](server/README.md)
