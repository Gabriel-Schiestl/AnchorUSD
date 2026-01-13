Arquitetura:

- blockchain: clients and interfaces for on-chain communication
- config: many different configs
- domain: calculations and aggregations, nothing about I/O, just pure code
- handler: handlers that receives http request and forward them to services
- http: server and routes
- model: data representation(DB entities, DTOs)
- service: orchestrate data workflow
- storage: DB communication and operations
