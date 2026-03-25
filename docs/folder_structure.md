# Workspace Folder Structure

Here is the folder structure tree for the `2hand-shop` workspace.

```text
.
├── README.md
├── deployments/
│   └── docker-compose/
│       └── docker-compose.yml
├── docs/
├── gateway/
│   ├── tyk.conf
│   └── apps/
│       └── identity-api.json
├── internal/
│   ├── pkg/
│   │   ├── errors/
│   │   ├── logger/
│   │   ├── migration/
│   │   ├── postgresqlx/
│   │   ├── rabbitmq/
│   │   ├── utils/
│   │   ├── go.mod
│   │   └── go.sum
│   └── services/
│       ├── catalog/
│       └── identity/
│           ├── cmd/
│           ├── config/
│           ├── internal/
│           ├── test/
│           ├── Dockerfile
│           ├── .env
│           ├── go.mod
│           └── go.sum
├── scripts/
│   └── init_db.sh
└── web/ (empty)
```

## Summary of Components
- **gateway**: Infrastructure for managing APIs (using Tyk).
- **internal/pkg**: Shared Go packages used across different services.
- **internal/services**: Individual microservices (Identity, Catalog).
- **deployments**: Deployment configurations.
- **scripts**: Utility scripts for environment setup.
- **web**: Intended for the frontend application.
