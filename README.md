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
        "host": "localhost",
        "port": ":3000",
        "redisHost": "localhost",
        "redisPort": "6379",
        "redisPass": "<password>"  
    },
    "mailSender": {
        "email": "<email>",
        "sendgrid_key": "<sendgrid-key>",
        "timeout": 20 
    },
    "database": {
        "file": "./database.db"
    },
    "token": {
        "secret": "<secret>",
        "timeout": 100
    },
    "account": {
        "mnemonics": "<mnemonics>",
        "network": "<grid-network>"
    },
    "version": "v1",
    "salt": "<salt>",
    "admins": [],
    "notifyAdminsIntervalHours": 6
}
```

## Frontend

check frontend [README](client/README.md)

## Backend

check backend [README](server/README.md)
