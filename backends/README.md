# Backends

Backends are the set of supported backend databases.

In the backend, You should provide the following implementations:

- Connector: build backend connection according to connection string
- Requester: reuse the Connection and fire request to the target database
 
