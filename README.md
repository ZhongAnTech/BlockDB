# BlockDB
Are you looking for a super easy-to-use blockchain database that can record anything to it, feeling just like MySQL or MongoDB, with the features provided by blockchain such as immutability, distributed ledger and high availability being integrated? You got to the right place. 

BlockDB is an out-of-box database that is built on blockchain, providing the SQL/NoSQL query ability seamlessly.


# Features
## Blockchain (Distributed Ledger Technology) based
As a blockchain(DLT) based database, BlockDB will honestly record all activities performed on the database. Full operation logs and change histories will be recorded in an append-only style to provide immutability, security and auditability.

We provide a modern DLT implementation: [Annchain.OG](https://github.com/annchain/OG) as the default BlockDB backend. Annchain.OG is leveraging DAG (Directed Acyclic Graph), DKG (Distributed Key Generation), BLS threshold signature and so many other modern technologies to support 10k+ TPS (Transactions per second) upon public network, comparing to the traditional chain-based DLT that can only reach 100 or less.

## Immutability
BlockDB provides two ways of auditing: operation auditing and data auditing. Both auditing methods are leveraging the capability of immutability provided by DLT.

Users may configure the BlockDB to either way, according to their purpose.
### Immutable Data Storage
In this mode, BlockDB acts as an immutable append-only database. Currently the majority of successful applications of blockchain uses blockchain as "immutable append-only distributed data storage". Proofs, certificates, transactions are all those records that will never be changed and will need complete immutability. However the complexity of setting up distributed ledger and adapting records onto it is too high to welcome normal developers.

**We greatly simplify the process of immutable storage in BlockDB**:
 In BlockDB, all processes of DLT are well-encapsulated. Developers with little knowledge of cryptographic, distributed systems, p2p communications could still setup BlockDB in a really short time. Building systems upon BlockDB is as easy as building one upon traditional SQL/NoSQL databases. Developers may focus on the business logic and let BlockDB handle all the rest tricky part of the work.

### CRUD Operation Auditing
In this mode, BlockDB is acting like a proxy. It will delegate all CRUD operations to the target database (like MongoDB). During the delegation, these operations will be sent to DLT and recorded. If you have a legacy system using traditional database, feel free to configure BlockDB as a middle proxy to enable full audit capability of every insertion, update,deletion or even query.

## Database Sharing
Since BlockDB is based on DLT, it is really natural to implement trustless high-availability data sharing among multiple parties around the world. There will be no single center for the data storage, thus modifications to data without complete consensus are not possible in BlockDB.


## Interfaces
BlockDB supports multiple ways for you to send data. All ways listed here are configurable to enable/disable.

### Kafka
BlockDB can listen to a Kafka MQ and consume the data. 

### Log4j
BlockDB opens a Log4j appender receiver to allow direct write from log frameworks.

### Socket
BlockDB opens a Socket listener to receive JSON stream.

### HTTP
BlockDB opens an HTTP listener to receive JSON requests.

### Intercept
*For CRUD Operation Auditing only*: change the target database URL from original one to the one provided by BlockDB so that all operations will be fully audited.


## Queries
BlockDB supports MongoDB style query on all data records. There are mainly three parts in the queriable data:

+ *DLT info*: Info of the chain structure. e.g., the height, hash, etc.
+ *Audit info*: Info of the audit meta. e.g., operator, timestamp, IP, browser, etc.
+ *User defined info*: Customizable info provided by user. Can be anything.

Note that all data are recorded **directly and completely** onto the ledger. BlockDB stores not only the hash of content (which most of the certificate applications do), but also the **full content** user provides. This enables data flowing among parties. 