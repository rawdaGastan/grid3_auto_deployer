# Backend Server

Go backend server using sqlite3 db.

## Requirements

- Go >= 1.21
- make
- docker

## Configuration

Before building or running backend create `config.json` in `server` dir.

example `config.json`:

```json
{
    "server": {
        "host": "localhost, required",
        "port": ":3000, required",
        "redisHost": "redis-db, make sure to change it in docker compose if you have other redis configurations, required",
        "redisPort": "6379, make sure to change it in docker compose if you have other redis configurations, required",
        "redisPass": "pass, make sure to change it in docker compose if you have other redis configurations, required" 
    },
    "mailSender": {
        "email": "your sendgrid account sender, required",
        "sendgrid_key": "<sendgrid-key>, required",
        "timeout": "<the timeout for app mail verification codes in seconds, required>"
    },
    "database": {
        "file": "<the path of the database file you have or you want to create, default is `database.sql`>"
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

## Build

```bash
make build
```

### Build Using Docker

```bash
docker build -t cloud4students .
```

## Run

```bash
make run
```

### Run Using Docker

```bash
docker run cloud4students
```

### Swagger

- Install swag binary

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

- Generate swagger docs

```bash
swag init
```

- You can access swagger through `/swagger/index.html`.
- Example: if your port is `3000` and host is `localhost`, then you can access swagger using `http://localhost:3000/swagger/index.html`
