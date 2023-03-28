# cloud4students

[![Codacy Badge](https://app.codacy.com/project/badge/Grade/cd6e18aac6be404ab89ec160b4b36671)](https://www.codacy.com/gh/threefoldtech/grid3-go/dashboard?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=threefoldtech/grid3-go&amp;utm_campaign=Badge_Grade) [![Testing](https://github.com/codescalers/cloud4students/actions/workflows/gotest.yml/badge.svg?branch=development)](https://github.com/codescalers/cloud4students/actions/workflows/gotest.yml)

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
        "port": ":3000"
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
        "secret": "mysecret",
        "timeout": 100
    },
    "account": {
        "mnemonics": "<mnemonics>",
        "network": "<grid-network>"
    },
    "version": "v1"
}
```

## Frontend

check frontend [README](client/README.md)

## Backend

check backend [README](server/README.md)
